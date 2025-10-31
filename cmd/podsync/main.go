package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/daleiii/podsync-web/pkg/feed"
	"github.com/daleiii/podsync-web/pkg/history"
	"github.com/daleiii/podsync-web/pkg/model"
	"github.com/daleiii/podsync-web/services/api"
	"github.com/daleiii/podsync-web/services/update"
	"github.com/daleiii/podsync-web/services/web"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/daleiii/podsync-web/pkg/db"
	"github.com/daleiii/podsync-web/pkg/fs"
	"github.com/daleiii/podsync-web/pkg/ytdl"
)

type Opts struct {
	ConfigPath string `long:"config" short:"c" default:"config.toml" env:"PODSYNC_CONFIG_PATH"`
	Headless   bool   `long:"headless"`
	Debug      bool   `long:"debug"`
	NoBanner   bool   `long:"no-banner"`
}

const banner = `
 _______  _______  ______   _______           _        _______ 
(  ____ )(  ___  )(  __  \ (  ____ \|\     /|( (    /|(  ____ \
| (    )|| (   ) || (  \  )| (    \/( \   / )|  \  ( || (    \/
| (____)|| |   | || |   ) || (_____  \ (_) / |   \ | || |      
|  _____)| |   | || |   | |(_____  )  \   /  | (\ \) || |      
| (      | |   | || |   ) |      ) |   ) (   | | \   || |      
| )      | (___) || (__/  )/\____) |   | |   | )  \  || (____/\
|/       (_______)(______/ \_______)   \_/   |/    )_)(_______/
`

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	arch    = ""
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse args
	opts := Opts{}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.WithError(err).Fatal("failed to parse command line arguments")
	}

	if opts.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if !opts.NoBanner {
		log.Info(banner)
	}

	log.WithFields(log.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
		"arch":    arch,
	}).Info("running podsync")

	// Load TOML file
	log.Debugf("loading configuration %q", opts.ConfigPath)
	cfg, err := LoadConfig(opts.ConfigPath)
	if err != nil {
		log.WithError(err).Fatal("failed to load configuration file")
	}

	if cfg.Log.Filename != "" {
		log.Infof("Using log file: %s", cfg.Log.Filename)

		log.SetOutput(&lumberjack.Logger{
			Filename:   cfg.Log.Filename,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,
			Compress:   cfg.Log.Compress,
		})

		// Optionally enable debug mode from config.toml
		if cfg.Log.Debug {
			log.SetLevel(log.DebugLevel)
		}
	}

	downloader, err := ytdl.New(ctx, cfg.Downloader)
	if err != nil {
		log.WithError(err).Fatal("youtube-dl error")
	}

	database, err := db.NewBadger(&cfg.Database)
	if err != nil {
		log.WithError(err).Fatal("failed to open database")
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.WithError(err).Error("failed to close database")
		}
	}()

	var storage fs.Storage
	switch cfg.Storage.Type {
	case "local":
		storage, err = fs.NewLocal(cfg.Storage.Local.DataDir, cfg.Server.WebUIEnabled)
	case "s3":
		storage, err = fs.NewS3(cfg.Storage.S3) // serving files from S3 is not supported, so no WebUI either
	default:
		log.Fatalf("unknown storage type: %s", cfg.Storage.Type)
	}
	if err != nil {
		log.WithError(err).Fatal("failed to open storage")
	}

	// Run updater thread
	log.Debug("creating key providers")
	keys := map[model.Provider]feed.KeyProvider{}
	for name, list := range cfg.Tokens {
		provider, err := feed.NewKeyProvider(list)
		if err != nil {
			log.WithError(err).Fatalf("failed to create key provider for %q", name)
		}
		keys[name] = provider
	}

	// Construct full backend URL from hostname + port
	backendURL := fmt.Sprintf("%s:%d", cfg.Server.Hostname, cfg.Server.Port)

	// Create history manager
	historyManager := history.NewManager(database, cfg.History.Enabled)
	log.Infof("history tracking: enabled=%v, retention=%d days, max_entries=%d",
		cfg.History.Enabled, cfg.History.RetentionDays, cfg.History.MaxEntries)

	// Only create update manager if we have feeds
	var manager *update.Manager
	if len(cfg.Feeds) > 0 {
		log.Debug("creating update manager")
		manager, err = update.NewUpdater(cfg.Feeds, keys, backendURL, downloader, database, storage, historyManager)
		if err != nil {
			log.WithError(err).Fatal("failed to create updater")
		}

		// In Headless mode, do one round of feed updates and quit
		if opts.Headless {
			for _, _feed := range cfg.Feeds {
				if err := manager.Update(ctx, _feed); err != nil {
					log.WithError(err).Errorf("failed to update feed: %s", _feed.URL)
				}
			}
			return
		}
	} else {
		log.Info("no feeds configured, skipping feed updater")
		// In headless mode with no feeds, just exit
		if opts.Headless {
			log.Info("headless mode with no feeds, exiting")
			return
		}
	}

	// Queue of feeds to update
	updates := make(chan *feed.Config, 16)
	defer close(updates)

	group, ctx := errgroup.WithContext(ctx)
	defer func() {
		if err := group.Wait(); err != nil && (err != context.Canceled && err != http.ErrServerClosed) {
			log.WithError(err).Error("wait error")
		}
		log.Info("gracefully stopped")
	}()

	// Only run feed update goroutines if we have feeds
	if len(cfg.Feeds) > 0 {
		// Create Cron
		c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)))
		m := make(map[string]cron.EntryID)

		// Run updates listener
		group.Go(func() error {
			for {
				select {
				case _feed := <-updates:
					if err := manager.Update(ctx, _feed); err != nil {
						log.WithError(err).Errorf("failed to update feed: %s", _feed.URL)
					} else {
						log.Infof("next update of %s: %s", _feed.ID, c.Entry(m[_feed.ID]).Next)
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})

		// Run cron scheduler
		group.Go(func() error {
			var cronID cron.EntryID

			for _, _feed := range cfg.Feeds {
				// Track if this feed has an explicit cron schedule
				hasExplicitCronSchedule := _feed.CronSchedule != ""

				if _feed.CronSchedule == "" {
					_feed.CronSchedule = fmt.Sprintf("@every %s", _feed.UpdatePeriod.String())
				}
				cronFeed := _feed
				if cronID, err = c.AddFunc(cronFeed.CronSchedule, func() {
					log.Debugf("adding %q to update queue", cronFeed.ID)
					updates <- cronFeed
				}); err != nil {
					log.WithError(err).Fatalf("can't create cron task for feed: %s", cronFeed.ID)
				}

				m[cronFeed.ID] = cronID
				log.Debugf("-> %s (update '%s')", cronFeed.ID, cronFeed.CronSchedule)

				// Only perform initial update if no explicit cron schedule is configured
				// This prevents unwanted updates when using fixed schedules in Docker deployments
				if !hasExplicitCronSchedule {
					updates <- cronFeed
				}
			}

			c.Start()

			for {
				<-ctx.Done()

				log.Info("shutting down cron")
				c.Stop()

				return ctx.Err()
			}
		})
	}

	if cfg.Storage.Type == "s3" {
		return // S3 content is hosted externally
	}

	// Convert tokens map to string map for API
	tokensMap := make(map[string][]string)
	for provider, keys := range cfg.Tokens {
		tokensMap[string(provider)] = []string(keys)
	}

	// Create API router
	apiRouter := api.NewRouter(cfg.Feeds, cfg.Server, database, backendURL, opts.ConfigPath, tokensMap, manager, downloader, cfg.History.RetentionDays, cfg.History.MaxEntries)

	// Run web server with API
	srv := web.NewWithAPI(cfg.Server, storage, database, apiRouter.Handler())

	group.Go(func() error {
		log.Infof("running listener at %s", srv.Addr)
		if cfg.Server.TLS {
			return srv.ListenAndServeTLS(cfg.Server.CertificatePath, cfg.Server.KeyFilePath)
		} else {
			return srv.ListenAndServe()
		}
	})

	group.Go(func() error {
		// Shutdown web server
		defer func() {
			ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer func() {
				cancel()
			}()
			log.Info("shutting down web server")
			if err := srv.Shutdown(ctxShutDown); err != nil {
				log.WithError(err).Error("server shutdown failed")
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-stop:
				cancel()
				return nil
			}
		}
	})
}
