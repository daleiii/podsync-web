package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/daleiii/podsync-web/pkg/db"
	"github.com/daleiii/podsync-web/pkg/feed"
	"github.com/daleiii/podsync-web/pkg/ytdl"
	"github.com/daleiii/podsync-web/services/api/models"
	"github.com/daleiii/podsync-web/services/web"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

// ConfigHandler handles configuration-related API endpoints
type ConfigHandler struct {
	feeds      map[string]*feed.Config
	server     web.Config
	database   db.Storage
	tokens     map[string][]string
	configPath string
	downloader *ytdl.YoutubeDl
}

// NewConfigHandler creates a new configuration handler
func NewConfigHandler(feeds map[string]*feed.Config, server web.Config, database db.Storage, tokens map[string][]string, configPath string, downloader *ytdl.YoutubeDl) *ConfigHandler {
	return &ConfigHandler{
		feeds:      feeds,
		server:     server,
		database:   database,
		tokens:     tokens,
		configPath: configPath,
		downloader: downloader,
	}
}

// GetConfig returns the current configuration
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read config file to get latest saved values for server/storage/downloader/tokens
	var serverConfig models.ServerConfig
	var storageConfig models.StorageConfig
	var downloaderConfig models.DownloaderConfig
	var tokensConfig models.TokensConfig

	data, err := os.ReadFile(h.configPath)
	if err == nil {
		tree, err := toml.LoadBytes(data)
		if err == nil {
			// Read server config from file
			if serverTree := tree.Get("server"); serverTree != nil {
				if st, ok := serverTree.(*toml.Tree); ok {
					if v := st.Get("hostname"); v != nil {
						if s, ok := v.(string); ok {
							serverConfig.Hostname = s
						}
					}
					if v := st.Get("port"); v != nil {
						if i, ok := v.(int64); ok {
							serverConfig.Port = int(i)
						}
					}
					if v := st.Get("frontend_port"); v != nil {
						if i, ok := v.(int64); ok {
							serverConfig.FrontendPort = int(i)
						}
					}
					if v := st.Get("bind_address"); v != nil {
						if s, ok := v.(string); ok {
							serverConfig.BindAddress = s
						}
					}
					if v := st.Get("tls"); v != nil {
						if b, ok := v.(bool); ok {
							serverConfig.TLS = b
						}
					}
					if v := st.Get("certificate_path"); v != nil {
						if s, ok := v.(string); ok {
							serverConfig.CertificatePath = s
						}
					}
					if v := st.Get("key_file_path"); v != nil {
						if s, ok := v.(string); ok {
							serverConfig.KeyFilePath = s
						}
					}
					if v := st.Get("path"); v != nil {
						if s, ok := v.(string); ok {
							serverConfig.Path = s
						}
					}
					if v := st.Get("web_ui"); v != nil {
						if b, ok := v.(bool); ok {
							serverConfig.WebUIEnabled = b
						}
					}
				}
			}

			// Read storage config from file
			if storageTree := tree.Get("storage"); storageTree != nil {
				if st, ok := storageTree.(*toml.Tree); ok {
					if v := st.Get("type"); v != nil {
						if s, ok := v.(string); ok {
							storageConfig.Type = s
						}
					}
					if localTree := st.Get("local"); localTree != nil {
						if lt, ok := localTree.(*toml.Tree); ok {
							storageConfig.Local = &models.LocalStorageConfig{}
							if v := lt.Get("data_dir"); v != nil {
								if s, ok := v.(string); ok {
									storageConfig.Local.DataDir = s
								}
							}
						}
					}
					if s3Tree := st.Get("s3"); s3Tree != nil {
						if s3t, ok := s3Tree.(*toml.Tree); ok {
							storageConfig.S3 = &models.S3StorageConfig{}
							if v := s3t.Get("bucket"); v != nil {
								if s, ok := v.(string); ok {
									storageConfig.S3.Bucket = s
								}
							}
							if v := s3t.Get("region"); v != nil {
								if s, ok := v.(string); ok {
									storageConfig.S3.Region = s
								}
							}
							if v := s3t.Get("endpoint_url"); v != nil {
								if s, ok := v.(string); ok {
									storageConfig.S3.EndpointURL = s
								}
							}
							if v := s3t.Get("prefix"); v != nil {
								if s, ok := v.(string); ok {
									storageConfig.S3.Prefix = s
								}
							}
							if v := s3t.Get("access_key"); v != nil {
								if s, ok := v.(string); ok {
									storageConfig.S3.AccessKey = s
								}
							}
							if v := s3t.Get("secret_key"); v != nil {
								if s, ok := v.(string); ok {
									storageConfig.S3.SecretKey = s
								}
							}
						}
					}
				}
			}

			// Read downloader config from file
			if dlTree := tree.Get("downloader"); dlTree != nil {
				if dt, ok := dlTree.(*toml.Tree); ok {
					if v := dt.Get("self_update"); v != nil {
						if b, ok := v.(bool); ok {
							downloaderConfig.SelfUpdate = b
						}
					}
					if v := dt.Get("update_channel"); v != nil {
						if s, ok := v.(string); ok {
							downloaderConfig.UpdateChannel = s
						}
					}
					if v := dt.Get("update_version"); v != nil {
						if s, ok := v.(string); ok {
							downloaderConfig.UpdateVersion = s
						}
					}
					if v := dt.Get("timeout"); v != nil {
						if s, ok := v.(string); ok {
							downloaderConfig.Timeout = s
						}
					}
				}
			}

			// Read tokens from file
			if tokensTree := tree.Get("tokens"); tokensTree != nil {
				if tt, ok := tokensTree.(*toml.Tree); ok {
					if v := tt.Get("youtube"); v != nil {
						if arr, ok := v.([]interface{}); ok {
							for _, item := range arr {
								if s, ok := item.(string); ok {
									tokensConfig.YouTube = append(tokensConfig.YouTube, s)
								}
							}
						}
					}
					if v := tt.Get("vimeo"); v != nil {
						if arr, ok := v.([]interface{}); ok {
							for _, item := range arr {
								if s, ok := item.(string); ok {
									tokensConfig.Vimeo = append(tokensConfig.Vimeo, s)
								}
							}
						}
					}
				}
			}
		}
	}

	// Use in-memory values as fallback if file read failed
	if serverConfig.Hostname == "" {
		serverConfig = models.ServerConfig{
			Hostname:        h.server.Hostname,
			Port:            h.server.Port,
			FrontendPort:    h.server.FrontendPort,
			BindAddress:     h.server.BindAddress,
			TLS:             h.server.TLS,
			CertificatePath: h.server.CertificatePath,
			KeyFilePath:     h.server.KeyFilePath,
			Path:            h.server.Path,
			WebUIEnabled:    h.server.WebUIEnabled,
		}
	}
	if storageConfig.Type == "" {
		storageConfig = models.StorageConfig{
			Type: "local",
			Local: &models.LocalStorageConfig{
				DataDir: h.server.DataDir,
			},
		}
	}
	if downloaderConfig.Timeout == "" {
		downloaderConfig = models.DownloaderConfig{
			SelfUpdate: false,
			Timeout:    "30s",
		}
	}

	// Get yt-dlp version if downloader is available
	if h.downloader != nil {
		ctx := context.Background()
		if version, err := h.downloader.Version(ctx); err == nil {
			downloaderConfig.YtdlVersion = strings.TrimSpace(version)
		}
	}
	if len(tokensConfig.YouTube) == 0 && len(tokensConfig.Vimeo) == 0 {
		tokensConfig = models.TokensConfig{
			YouTube:    h.tokens["youtube"],
			Vimeo:      h.tokens["vimeo"],
			SoundCloud: h.tokens["soundcloud"],
			Twitch:     h.tokens["twitch"],
		}
	}

	// Build feeds configuration map from in-memory (this is complex to reload)
	feedsConfig := make(map[string]*models.FeedConfig)
	for id, cfg := range h.feeds {
		cleanupKeep := 0
		if cfg.Clean != nil {
			cleanupKeep = cfg.Clean.KeepLast
		}

		feedsConfig[id] = &models.FeedConfig{
			UpdatePeriod: cfg.UpdatePeriod.String(),
			CronSchedule: cfg.CronSchedule,
			Quality:      string(cfg.Quality),
			Format:       string(cfg.Format),
			PageSize:     cfg.PageSize,
			MaxHeight:    cfg.MaxHeight,
			CleanupKeep:  cleanupKeep,
			PlaylistSort: string(cfg.PlaylistSort),
			PrivateFeed:  cfg.PrivateFeed,
			OPML:         cfg.OPML,
			Filters: models.Filters{
				Title:          cfg.Filters.Title,
				NotTitle:       cfg.Filters.NotTitle,
				Description:    cfg.Filters.Description,
				NotDescription: cfg.Filters.NotDescription,
				MinDuration:    cfg.Filters.MinDuration,
				MaxDuration:    cfg.Filters.MaxDuration,
				MaxAge:         cfg.Filters.MaxAge,
				MinAge:         cfg.Filters.MinAge,
			},
			Custom: models.Custom{
				CoverArt:        cfg.Custom.CoverArt,
				CoverArtQuality: string(cfg.Custom.CoverArtQuality),
				Category:        cfg.Custom.Category,
				Subcategories:   cfg.Custom.Subcategories,
				Explicit:        cfg.Custom.Explicit,
				Language:        cfg.Custom.Language,
				Author:          cfg.Custom.Author,
				Title:           cfg.Custom.Title,
				Description:     cfg.Custom.Description,
				OwnerName:       cfg.Custom.OwnerName,
				OwnerEmail:      cfg.Custom.OwnerEmail,
				Link:            cfg.Custom.Link,
			},
		}
	}

	response := models.ConfigResponse{
		Server:     serverConfig,
		Storage:    storageConfig,
		Feeds:      feedsConfig,
		Database:   models.DatabaseConfig{Dir: "db"},
		Downloader: downloaderConfig,
		Tokens:     tokensConfig,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.WithError(err).Error("failed to encode config response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
