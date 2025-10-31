package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

// ConfigUpdateHandler handles configuration update API endpoints
type ConfigUpdateHandler struct {
	configPath string
	writer     *config.Writer
}

// NewConfigUpdateHandler creates a new config update handler
func NewConfigUpdateHandler(configPath string) *ConfigUpdateHandler {
	return &ConfigUpdateHandler{
		configPath: configPath,
		writer:     config.NewWriter(configPath),
	}
}

// UpdateServer updates server configuration
func (h *ConfigUpdateHandler) UpdateServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("failed to decode server config update request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the [server] section in TOML
	err := h.writer.UpdatePartial(func(tree *toml.Tree) error {
		var serverTree *toml.Tree
		if tree.Get("server") != nil {
			serverTree = tree.Get("server").(*toml.Tree)
		}
		if serverTree == nil {
			serverTree, _ = toml.TreeFromMap(make(map[string]interface{}))
			tree.Set("server", serverTree)
		}

		// Update each field if provided
		if hostname, ok := req["hostname"]; ok {
			serverTree.Set("hostname", hostname)
		}
		if port, ok := req["port"]; ok {
			// Convert float64 to int64 for TOML compatibility
			if portFloat, ok := port.(float64); ok {
				serverTree.Set("port", int64(portFloat))
			} else if portInt, ok := port.(int); ok {
				serverTree.Set("port", int64(portInt))
			} else {
				serverTree.Set("port", port)
			}
		}
		if frontendPort, ok := req["frontend_port"]; ok {
			// Convert float64 to int64 for TOML compatibility
			if portFloat, ok := frontendPort.(float64); ok {
				serverTree.Set("frontend_port", int64(portFloat))
			} else if portInt, ok := frontendPort.(int); ok {
				serverTree.Set("frontend_port", int64(portInt))
			} else {
				serverTree.Set("frontend_port", frontendPort)
			}
		}
		if bindAddr, ok := req["bind_address"]; ok {
			serverTree.Set("bind_address", bindAddr)
		}
		if tls, ok := req["tls"]; ok {
			serverTree.Set("tls", tls)
		}
		if certPath, ok := req["certificate_path"]; ok {
			serverTree.Set("certificate_path", certPath)
		}
		if keyPath, ok := req["key_file_path"]; ok {
			serverTree.Set("key_file_path", keyPath)
		}
		if path, ok := req["path"]; ok {
			serverTree.Set("path", path)
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Error("failed to update server configuration")
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Server configuration updated successfully. Restart required for changes to take effect.",
	})
}

// UpdateStorage updates storage configuration
func (h *ConfigUpdateHandler) UpdateStorage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("failed to decode storage config update request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the [storage] section in TOML
	err := h.writer.UpdatePartial(func(tree *toml.Tree) error {
		var storageTree *toml.Tree
		if tree.Get("storage") != nil {
			storageTree = tree.Get("storage").(*toml.Tree)
		}
		if storageTree == nil {
			storageTree, _ = toml.TreeFromMap(make(map[string]interface{}))
			tree.Set("storage", storageTree)
		}

		// Update storage type if provided
		if storageType, ok := req["type"]; ok {
			storageTree.Set("type", storageType)
		}

		// Update local storage settings
		if local, ok := req["local"].(map[string]interface{}); ok {
			var localTree *toml.Tree
			if storageTree.Get("local") != nil {
				localTree = storageTree.Get("local").(*toml.Tree)
			}
			if localTree == nil {
				localTree, _ = toml.TreeFromMap(make(map[string]interface{}))
				storageTree.Set("local", localTree)
			}
			if dataDir, ok := local["data_dir"]; ok {
				localTree.Set("data_dir", dataDir)
			}
		}

		// Update S3 storage settings if provided
		if s3, ok := req["s3"].(map[string]interface{}); ok {
			var s3Tree *toml.Tree
			if storageTree.Get("s3") != nil {
				s3Tree = storageTree.Get("s3").(*toml.Tree)
			}
			if s3Tree == nil {
				s3Tree, _ = toml.TreeFromMap(make(map[string]interface{}))
				storageTree.Set("s3", s3Tree)
			}
			for key, value := range s3 {
				s3Tree.Set(key, value)
			}
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Error("failed to update storage configuration")
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Storage configuration updated successfully.",
	})
}

// UpdateDownloader updates downloader configuration
func (h *ConfigUpdateHandler) UpdateDownloader(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("failed to decode downloader config update request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the [downloader] section in TOML
	err := h.writer.UpdatePartial(func(tree *toml.Tree) error {
		var downloaderTree *toml.Tree
		if tree.Get("downloader") != nil {
			downloaderTree = tree.Get("downloader").(*toml.Tree)
		}
		if downloaderTree == nil {
			downloaderTree, _ = toml.TreeFromMap(make(map[string]interface{}))
			tree.Set("downloader", downloaderTree)
		}

		// Update self_update if provided
		if selfUpdate, ok := req["self_update"]; ok {
			downloaderTree.Set("self_update", selfUpdate)
		}

		// Update update_channel if provided
		if updateChannel, ok := req["update_channel"]; ok {
			downloaderTree.Set("update_channel", updateChannel)
		}

		// Update update_version if provided
		if updateVersion, ok := req["update_version"]; ok {
			downloaderTree.Set("update_version", updateVersion)
		}

		// Update timeout if provided
		if timeout, ok := req["timeout"]; ok {
			downloaderTree.Set("timeout", timeout)
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Error("failed to update downloader configuration")
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Downloader configuration updated successfully.",
	})
}

// UpdateAuth updates authentication configuration
func (h *ConfigUpdateHandler) UpdateAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("failed to decode auth config update request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the [server.basic_auth] section in TOML
	err := h.writer.UpdatePartial(func(tree *toml.Tree) error {
		var serverTree *toml.Tree
		if tree.Get("server") != nil {
			serverTree = tree.Get("server").(*toml.Tree)
		}
		if serverTree == nil {
			serverTree, _ = toml.TreeFromMap(make(map[string]interface{}))
			tree.Set("server", serverTree)
		}

		var authTree *toml.Tree
		if serverTree.Get("basic_auth") != nil {
			authTree = serverTree.Get("basic_auth").(*toml.Tree)
		}
		if authTree == nil {
			authTree, _ = toml.TreeFromMap(make(map[string]interface{}))
			serverTree.Set("basic_auth", authTree)
		}

		// Update auth fields if provided
		if enabled, ok := req["enabled"]; ok {
			authTree.Set("enabled", enabled)
		}
		if username, ok := req["username"]; ok {
			authTree.Set("username", username)
		}
		if password, ok := req["password"]; ok {
			authTree.Set("password", password)
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Error("failed to update auth configuration")
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Authentication configuration updated successfully. Restart required for changes to take effect.",
	})
}

// UpdateTokens updates API tokens configuration
func (h *ConfigUpdateHandler) UpdateTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("failed to decode tokens config update request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the [tokens] section in TOML
	err := h.writer.UpdatePartial(func(tree *toml.Tree) error {
		var tokensTree *toml.Tree
		if tree.Get("tokens") != nil {
			tokensTree = tree.Get("tokens").(*toml.Tree)
		}
		if tokensTree == nil {
			tokensTree, _ = toml.TreeFromMap(make(map[string]interface{}))
			tree.Set("tokens", tokensTree)
		}

		// Parse comma-separated tokens and update each provider
		for provider, value := range req {
			if strValue, ok := value.(string); ok && strValue != "" {
				// Split by comma and trim whitespace
				tokens := []string{}
				for _, token := range splitAndTrim(strValue, ",") {
					if token != "" {
						tokens = append(tokens, token)
					}
				}
				if len(tokens) > 0 {
					tokensTree.Set(provider, tokens)
				}
			}
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Error("failed to update tokens configuration")
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "API tokens updated successfully. Restart required for changes to take effect.",
	})
}

// Helper function to split and trim strings
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	current := ""
	for _, c := range s {
		if string(c) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// UpdateHistory updates history configuration
func (h *ConfigUpdateHandler) UpdateHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("failed to decode history config update request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the [history] section in TOML
	err := h.writer.UpdatePartial(func(tree *toml.Tree) error {
		var historyTree *toml.Tree
		if tree.Get("history") != nil {
			historyTree = tree.Get("history").(*toml.Tree)
		}
		if historyTree == nil {
			historyTree, _ = toml.TreeFromMap(make(map[string]interface{}))
			tree.Set("history", historyTree)
		}

		// Update each field if provided
		if enabled, ok := req["enabled"]; ok {
			historyTree.Set("enabled", enabled)
		}
		if retentionDays, ok := req["retention_days"]; ok {
			// Convert float64 to int64 for TOML compatibility
			if daysFloat, ok := retentionDays.(float64); ok {
				historyTree.Set("retention_days", int64(daysFloat))
			} else if daysInt, ok := retentionDays.(int); ok {
				historyTree.Set("retention_days", int64(daysInt))
			} else {
				historyTree.Set("retention_days", retentionDays)
			}
		}
		if maxEntries, ok := req["max_entries"]; ok {
			// Convert float64 to int64 for TOML compatibility
			if entriesFloat, ok := maxEntries.(float64); ok {
				historyTree.Set("max_entries", int64(entriesFloat))
			} else if entriesInt, ok := maxEntries.(int); ok {
				historyTree.Set("max_entries", int64(entriesInt))
			} else {
				historyTree.Set("max_entries", maxEntries)
			}
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Error("failed to update history configuration")
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "History configuration updated successfully. Restart required for changes to take effect.",
	})
}

// ReloadConfig triggers a configuration reload
// Note: This only reloads the config file representation, not runtime components
// For full reload, the application should be restarted
func (h *ConfigUpdateHandler) ReloadConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Configuration file updated. Note: A full application restart is recommended for all changes to take effect.",
	})
}

// RestartServer triggers a graceful application restart
func (h *ConfigUpdateHandler) RestartServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Info("restart requested via API")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Server restart initiated. The application will restart shortly.",
	})

	// Send restart signal after response is sent
	go func() {
		// Give the response time to be sent
		time.Sleep(500 * time.Millisecond)
		log.Info("sending restart signal")
		// Send SIGTERM to trigger graceful shutdown
		// The process manager (systemd, docker, etc.) should restart the process
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			log.WithError(err).Error("failed to find process")
			return
		}
		if err := p.Signal(syscall.SIGTERM); err != nil {
			log.WithError(err).Error("failed to send SIGTERM")
		}
	}()
}
