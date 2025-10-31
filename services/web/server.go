package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/daleiii/podsync-web/pkg/db"
	"github.com/daleiii/podsync-web/pkg/model"
)

type Server struct {
	http.Server
	db     db.Storage
	apiMux http.Handler
}

type Config struct {
	// Hostname to use for download links
	Hostname string `toml:"hostname"`
	// Port is a server port to listen to
	Port int `toml:"port"`
	// FrontendPort is the port for the frontend development server
	FrontendPort int `toml:"frontend_port"`
	// Bind a specific IP addresses for server
	// "*": bind all IP addresses which is default option
	// localhost or 127.0.0.1  bind a single IPv4 address
	BindAddress string `toml:"bind_address"`
	// Flag indicating if the server will use TLS
	TLS bool `toml:"tls"`
	// Path to a certificate file for TLS connections
	CertificatePath string `toml:"certificate_path"`
	// Path to a private key file for TLS connections
	KeyFilePath string `toml:"key_file_path"`
	// Specify path for reverse proxy and only [A-Za-z0-9]
	Path string `toml:"path"`
	// DataDir is a path to a directory to keep XML feeds and downloaded episodes,
	// that will be available to user via web server for download.
	DataDir string `toml:"data_dir"`
	// WebUIEnabled is a flag indicating if web UI is enabled
	WebUIEnabled bool `toml:"web_ui"`
	// BasicAuth configuration for HTTP basic authentication
	BasicAuth *BasicAuthConfig `toml:"basic_auth"`
}

type BasicAuthConfig struct {
	// Enabled indicates if basic auth is enabled
	Enabled bool `toml:"enabled"`
	// Username for basic auth
	Username string `toml:"username"`
	// Password for basic auth
	Password string `toml:"password"`
}

func New(cfg Config, storage http.FileSystem, database db.Storage) *Server {
	return NewWithAPI(cfg, storage, database, nil)
}

func NewWithAPI(cfg Config, storage http.FileSystem, database db.Storage, apiHandler http.Handler) *Server {
	port := cfg.Port
	if port == 0 {
		port = 8080
	}

	bindAddress := cfg.BindAddress
	if bindAddress == "*" {
		bindAddress = ""
	}

	srv := Server{
		db:     database,
		apiMux: apiHandler,
	}

	srv.Addr = fmt.Sprintf("%s:%d", bindAddress, port)
	log.Debugf("using address: %s:%s", bindAddress, srv.Addr)

	fileServer := http.FileServer(storage)

	// If WebUI is enabled, wrap the file server to handle SPA routing
	var handler = fileServer
	if cfg.WebUIEnabled {
		handler = spaHandler{fileServer: fileServer, storage: storage}
	}

	log.Debugf("handle path: /%s", cfg.Path)
	http.Handle(fmt.Sprintf("/%s", cfg.Path), handler)

	// Add health check endpoint
	http.HandleFunc("/health", srv.healthCheckHandler)

	// Add API routes if provided
	if apiHandler != nil {
		http.Handle("/api/", apiHandler)
	}

	return &srv
}

type HealthStatus struct {
	Status         string    `json:"status"`
	Timestamp      time.Time `json:"timestamp"`
	FailedEpisodes int       `json:"failed_episodes,omitempty"`
	Message        string    `json:"message,omitempty"`
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check for recent download failures within the last 24 hours
	failedCount := 0
	cutoffTime := time.Now().Add(-24 * time.Hour)

	// Walk through all feeds to count recent failures
	err := s.db.WalkFeeds(ctx, func(feed *model.Feed) error {
		return s.db.WalkEpisodes(ctx, feed.ID, func(episode *model.Episode) error {
			if episode.Status == model.EpisodeError && episode.PubDate.After(cutoffTime) {
				failedCount++
			}
			return nil
		})
	})

	w.Header().Set("Content-Type", "application/json")

	status := HealthStatus{
		Timestamp: time.Now(),
	}

	if err != nil {
		log.WithError(err).Error("health check database error")
		status.Status = "unhealthy"
		status.Message = "database error during health check"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if failedCount > 0 {
		status.Status = "unhealthy"
		status.FailedEpisodes = failedCount
		status.Message = fmt.Sprintf("found %d failed downloads in the last 24 hours", failedCount)
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		status.Status = "healthy"
		status.Message = "no recent download failures detected"
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(status)
}

// spaHandler wraps a file server to properly handle SPA routing
type spaHandler struct {
	fileServer http.Handler
	storage    http.FileSystem
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Try to open the requested path
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Check if the file exists
	f, err := h.storage.Open(path)
	if err == nil {
		f.Close()
		// File exists, serve it
		h.fileServer.ServeHTTP(w, r)
		return
	}

	// File doesn't exist, check if it's an API or feed request
	// API requests should get 404, but UI routes should get index.html
	if len(path) > 4 && (path[:5] == "/api/" || path[len(path)-4:] == ".xml" || path[len(path)-4:] == ".mp3" || path[len(path)-4:] == ".mp4") {
		http.NotFound(w, r)
		return
	}

	// For all other routes, serve index.html (SPA routing)
	indexFile, err := h.storage.Open("/index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer indexFile.Close()

	stat, err := indexFile.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile)
}
