package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/fs"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/mxpv/podsync/pkg/ytdl"
	"github.com/mxpv/podsync/services/web"
)

type Config struct {
	// Server is the web server configuration
	Server web.Config `toml:"server"`
	// S3 is the optional configuration for S3-compatible storage provider
	Storage fs.Config `toml:"storage"`
	// Log is the optional logging configuration
	Log Log `toml:"log"`
	// Database configuration
	Database db.Config `toml:"database"`
	// Feeds is a list of feeds to host by this app.
	// ID will be used as feed ID in http://podsync.net/{FEED_ID}.xml
	Feeds map[string]*feed.Config
	// Tokens is API keys to use to access YouTube/Vimeo APIs.
	Tokens map[model.Provider]StringSlice `toml:"tokens"`
	// Downloader (youtube-dl) configuration
	Downloader ytdl.Config `toml:"downloader"`
	// Global cleanup policy applied to feeds that don't specify their own cleanup policy
	Cleanup *feed.Cleanup `toml:"cleanup"`
	// History configuration for job tracking
	History HistoryConfig `toml:"history"`
}

// HistoryConfig contains configuration for job history tracking
type HistoryConfig struct {
	Enabled       bool `toml:"enabled"`
	RetentionDays int  `toml:"retention_days"`
	MaxEntries    int  `toml:"max_entries"`
}

type Log struct {
	// Filename to write the log to (instead of stdout)
	Filename string `toml:"filename"`
	// MaxSize is the maximum size of the log file in MB
	MaxSize int `toml:"max_size"`
	// MaxBackups is the maximum number of log file backups to keep after rotation
	MaxBackups int `toml:"max_backups"`
	// MaxAge is the maximum number of days to keep the logs for
	MaxAge int `toml:"max_age"`
	// Compress old backups
	Compress bool `toml:"compress"`
	// Debug mode
	Debug bool `toml:"debug"`
}

// LoadConfig loads TOML configuration from a file path
// If the file doesn't exist, returns a default configuration
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file exists, return default empty config
			log.Infof("config file not found at %s, using default configuration", path)
			config := &Config{
				Feeds: make(map[string]*feed.Config),
			}
			config.applyDefaults(path)
			config.applyEnv()
			if err := config.validate(); err != nil {
				return nil, err
			}
			return config, nil
		}
		return nil, errors.Wrapf(err, "failed to read config file: %s", path)
	}

	config := Config{}
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal toml")
	}

	for id, f := range config.Feeds {
		f.ID = id
	}

	config.applyDefaults(path)
	config.applyEnv()

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) validate() error {
	var result *multierror.Error

	if c.Server.DataDir != "" {
		log.Warnf(`server.data_dir is deprecated, and will be removed in a future release. Use the following config instead:

[storage]
  [storage.local]
  data_dir = "%s"

`, c.Server.DataDir)
		if c.Storage.Local.DataDir == "" {
			c.Storage.Local.DataDir = c.Server.DataDir
		}
	}

	if c.Server.Path != "" {
		var pathReg = regexp.MustCompile(model.PathRegex)
		if !pathReg.MatchString(c.Server.Path) {
			result = multierror.Append(result, errors.Errorf("Server handle path must be match %s or empty", model.PathRegex))
		}
	}

	switch c.Storage.Type {
	case "local":
		if c.Storage.Local.DataDir == "" {
			result = multierror.Append(result, errors.New("data directory is required for local storage"))
		}
	case "s3":
		if c.Storage.S3.EndpointURL == "" || c.Storage.S3.Region == "" || c.Storage.S3.Bucket == "" {
			result = multierror.Append(result, errors.New("S3 storage requires endpoint_url, region and bucket to be set"))
		}
	default:
		result = multierror.Append(result, errors.Errorf("unknown storage type: %s", c.Storage.Type))
	}

	// Allow zero feeds for initial setup via web UI
	// Users can add feeds later through the API

	for id, f := range c.Feeds {
		if f.URL == "" {
			result = multierror.Append(result, errors.Errorf("URL is required for %q", id))
		}
	}

	return result.ErrorOrNil()
}

func (c *Config) applyDefaults(configPath string) {
	// Set default port if not specified
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}

	if c.Server.Hostname == "" {
		if c.Server.Port != 0 && c.Server.Port != 80 {
			c.Server.Hostname = fmt.Sprintf("http://localhost:%d", c.Server.Port)
		} else {
			c.Server.Hostname = "http://localhost"
		}
	}

	if c.Storage.Type == "" {
		c.Storage.Type = "local"
	}

	// Set default data directory if not specified
	if c.Storage.Type == "local" && c.Storage.Local.DataDir == "" {
		c.Storage.Local.DataDir = filepath.Join(filepath.Dir(configPath), "data")
	}

	if c.Log.Filename != "" {
		if c.Log.MaxSize == 0 {
			c.Log.MaxSize = model.DefaultLogMaxSize
		}
		if c.Log.MaxAge == 0 {
			c.Log.MaxAge = model.DefaultLogMaxAge
		}
		if c.Log.MaxBackups == 0 {
			c.Log.MaxBackups = model.DefaultLogMaxBackups
		}
	}

	if c.Database.Dir == "" {
		c.Database.Dir = filepath.Join(filepath.Dir(configPath), "db")
	}

	// Set default history configuration if not specified
	if c.History.RetentionDays == 0 {
		c.History.RetentionDays = 30 // Default 30 days retention
	}
	if c.History.MaxEntries == 0 {
		c.History.MaxEntries = 1000 // Default max 1000 entries
	}
	// History is enabled by default only if no config was loaded
	// If a config file exists with [history] section, respect the enabled value
	data, err := os.ReadFile(configPath)
	if err != nil || !strings.Contains(string(data), "[history]") {
		// No config file or no [history] section - enable by default
		c.History.Enabled = true
	}

	// Web UI is enabled by default (can be disabled in config or via environment variable)
	// Only set to true if it wasn't explicitly set to false in the config
	if !c.Server.WebUIEnabled {
		// Check if this is truly the default (not explicitly set to false)
		// We'll allow the environment variable to override this in applyEnv()
		c.Server.WebUIEnabled = true
	}

	for _, _feed := range c.Feeds {
		if _feed.UpdatePeriod == 0 {
			_feed.UpdatePeriod = model.DefaultUpdatePeriod
		}

		if _feed.Quality == "" {
			_feed.Quality = model.DefaultQuality
		}

		if _feed.Custom.CoverArtQuality == "" {
			_feed.Custom.CoverArtQuality = model.DefaultQuality
		}

		if _feed.Format == "" {
			_feed.Format = model.DefaultFormat
		}

		if _feed.PageSize == 0 {
			_feed.PageSize = model.DefaultPageSize
		}

		if _feed.PlaylistSort == "" {
			_feed.PlaylistSort = model.SortingAsc
		}

		// Apply global cleanup policy if feed doesn't have its own
		if _feed.Clean == nil && c.Cleanup != nil {
			_feed.Clean = c.Cleanup
		}
	}
}

func (c *Config) applyEnv() {
	envVars := map[model.Provider]string{
		model.ProviderYoutube:    "PODSYNC_YOUTUBE_API_KEY",
		model.ProviderVimeo:      "PODSYNC_VIMEO_API_KEY",
		model.ProviderSoundcloud: "PODSYNC_SOUNDCLOUD_API_KEY",
		model.ProviderTwitch:     "PODSYNC_TWITCH_API_KEY",
	}

	// Replace API keys from config with environment variables
	for provider, envVar := range envVars {
		val, ok := os.LookupEnv(envVar)
		if ok {
			log.Infof("Found %s environment variable, replacing config token with it", envVar)
			// If no tokens are provided in the config.toml, we need to create a new map
			if c.Tokens == nil {
				c.Tokens = make(map[model.Provider]StringSlice)
			}
			// Support multiple keys separated by spaces for API key rotation
			keys := strings.Fields(val)
			c.Tokens[provider] = keys
		}
	}

	// Apply history configuration from environment variables
	if val, ok := os.LookupEnv("PODSYNC_HISTORY_ENABLED"); ok {
		c.History.Enabled = val == "true" || val == "1"
	}
	if val, ok := os.LookupEnv("PODSYNC_HISTORY_RETENTION_DAYS"); ok {
		if days, err := fmt.Sscanf(val, "%d", &c.History.RetentionDays); err == nil && days == 1 {
			log.Infof("Found PODSYNC_HISTORY_RETENTION_DAYS environment variable: %d days", c.History.RetentionDays)
		}
	}
	if val, ok := os.LookupEnv("PODSYNC_HISTORY_MAX_ENTRIES"); ok {
		if entries, err := fmt.Sscanf(val, "%d", &c.History.MaxEntries); err == nil && entries == 1 {
			log.Infof("Found PODSYNC_HISTORY_MAX_ENTRIES environment variable: %d entries", c.History.MaxEntries)
		}
	}

	// Apply Web UI configuration from environment variable
	if val, ok := os.LookupEnv("PODSYNC_WEB_UI"); ok {
		c.Server.WebUIEnabled = val == "true" || val == "1"
		log.Infof("Found PODSYNC_WEB_UI environment variable: %v", c.Server.WebUIEnabled)
	}
}

// StringSlice is a toml extension that lets you to specify either a string
// value (a slice with just one element) or a string slice.
type StringSlice []string

func (s *StringSlice) UnmarshalTOML(v interface{}) error {
	if str, ok := v.(string); ok {
		*s = []string{str}
		return nil
	}

	return errors.New("failed to decode string slice field")
}
