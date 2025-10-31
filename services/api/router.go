package api

import (
	"net/http"
	"strings"

	"github.com/daleiii/podsync-web/pkg/db"
	"github.com/daleiii/podsync-web/pkg/feed"
	"github.com/daleiii/podsync-web/pkg/history"
	"github.com/daleiii/podsync-web/pkg/progress"
	"github.com/daleiii/podsync-web/pkg/ytdl"
	"github.com/daleiii/podsync-web/services/api/handlers"
	"github.com/daleiii/podsync-web/services/api/middleware"
	"github.com/daleiii/podsync-web/services/web"
)

// Router sets up all API routes
type Router struct {
	configHandler       *handlers.ConfigHandler
	configUpdateHandler *handlers.ConfigUpdateHandler
	feedsHandler        *handlers.FeedsHandler
	episodesHandler     *handlers.EpisodesHandler
	progressHandler     *handlers.ProgressHandler
	historyHandler      *handlers.HistoryHandler
	serverConfig        web.Config
}

// NewRouter creates a new API router
func NewRouter(feeds map[string]*feed.Config, server web.Config, database db.Storage, hostname string, configPath string, tokens map[string][]string, updater handlers.UpdateManager, downloader *ytdl.YoutubeDl, historyRetentionDays, historyMaxEntries int) *Router {
	var progressTracker *progress.Tracker
	var historyManager *history.Manager

	// Handle the Go nil interface gotcha: an interface holding a nil pointer is not nil itself
	// We need to check if updater is actually usable (not a nil pointer wrapped in an interface)
	if updater != nil {
		// Use recover to handle potential panic from calling methods on nil pointer receiver
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Silently handle panic - updater was a nil pointer in an interface
					progressTracker = nil
					historyManager = nil
				}
			}()
			progressTracker = updater.GetProgressTracker()
			historyManager = updater.GetHistoryManager()
		}()
	}

	return &Router{
		configHandler:       handlers.NewConfigHandler(feeds, server, database, tokens, configPath, downloader),
		configUpdateHandler: handlers.NewConfigUpdateHandler(configPath),
		feedsHandler:        handlers.NewFeedsHandler(feeds, database, configPath, updater),
		episodesHandler:     handlers.NewEpisodesHandler(feeds, database, hostname, updater),
		progressHandler:     handlers.NewProgressHandler(progressTracker),
		historyHandler:      handlers.NewHistoryHandler(database, historyManager, historyRetentionDays, historyMaxEntries),
		serverConfig:        server,
	}
}

// Handler returns an http.Handler for the API routes
func (router *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	// Configuration endpoints
	mux.HandleFunc("/api/v1/config", router.configHandler.GetConfig)

	// Feed endpoints
	mux.HandleFunc("/api/v1/feeds", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			router.feedsHandler.ListFeeds(w, r)
		case http.MethodPost:
			router.feedsHandler.CreateFeed(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/feeds/", func(w http.ResponseWriter, r *http.Request) {
		// Parse path to determine which handler to call
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/feeds/")
		if path == "" {
			http.Error(w, "Feed ID required", http.StatusBadRequest)
			return
		}

		// Check if this is a refresh action
		pathParts := strings.Split(path, "/")
		if len(pathParts) == 2 && pathParts[1] == "refresh" {
			router.feedsHandler.RefreshFeed(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			router.feedsHandler.GetFeed(w, r)
		case http.MethodPut:
			router.feedsHandler.UpdateFeed(w, r)
		case http.MethodDelete:
			router.feedsHandler.DeleteFeed(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Configuration update endpoints
	mux.HandleFunc("/api/v1/config/server", router.configUpdateHandler.UpdateServer)
	mux.HandleFunc("/api/v1/config/storage", router.configUpdateHandler.UpdateStorage)
	mux.HandleFunc("/api/v1/config/downloader", router.configUpdateHandler.UpdateDownloader)
	mux.HandleFunc("/api/v1/config/tokens", router.configUpdateHandler.UpdateTokens)
	mux.HandleFunc("/api/v1/config/auth", router.configUpdateHandler.UpdateAuth)
	mux.HandleFunc("/api/v1/config/history", router.configUpdateHandler.UpdateHistory)
	mux.HandleFunc("/api/v1/config/restart", router.configUpdateHandler.RestartServer)
	mux.HandleFunc("/api/v1/config/tls/upload", handlers.HandleTLSUpload)

	// Episode endpoints
	mux.HandleFunc("/api/v1/episodes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			router.episodesHandler.ListEpisodes(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/episodes/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a retry or block action
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) == 6 && pathParts[5] == "retry" {
			router.episodesHandler.RetryEpisode(w, r)
			return
		}
		if len(pathParts) == 6 && pathParts[5] == "block" {
			router.episodesHandler.BlockEpisode(w, r)
			return
		}

		if r.Method == http.MethodDelete {
			router.episodesHandler.DeleteEpisode(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Progress endpoints
	mux.HandleFunc("/api/v1/progress", router.progressHandler.GetProgress)
	mux.HandleFunc("/api/v1/progress/stream", router.progressHandler.StreamProgress)

	// History endpoints
	mux.HandleFunc("/api/v1/history", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			router.historyHandler.ListHistory(w, r)
		case http.MethodDelete:
			router.historyHandler.DeleteAllHistory(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/history/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			router.historyHandler.GetHistoryStats(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/history/cleanup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			router.historyHandler.CleanupHistory(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/history/", func(w http.ResponseWriter, r *http.Request) {
		// Parse path to get history ID
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/history/")
		if path == "" || path == "stats" || path == "cleanup" {
			http.Error(w, "History ID required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			router.historyHandler.GetHistory(w, r)
		case http.MethodDelete:
			router.historyHandler.DeleteHistory(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Apply middleware chain
	handler := middleware.CORS(mux)

	// Apply basic auth if configured
	if router.serverConfig.BasicAuth != nil && router.serverConfig.BasicAuth.Enabled {
		handler = middleware.BasicAuth(router.serverConfig.BasicAuth.Username, router.serverConfig.BasicAuth.Password)(handler)
	}

	return handler
}
