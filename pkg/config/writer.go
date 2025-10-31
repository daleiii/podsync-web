package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Writer handles writing configuration to TOML files
type Writer struct {
	configPath string
}

// NewWriter creates a new config writer
func NewWriter(configPath string) *Writer {
	return &Writer{
		configPath: configPath,
	}
}

// WriteConfig writes the entire configuration to the TOML file
// It creates a backup of the existing file before writing
func (w *Writer) WriteConfig(cfg interface{}) error {
	// Create backup of existing config
	if err := w.backupConfig(); err != nil {
		log.WithError(err).Warn("failed to create config backup")
	}

	// Marshal config to TOML
	data, err := toml.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal config to TOML")
	}

	// Write to temporary file first for atomic write
	tmpFile := w.configPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return errors.Wrap(err, "failed to write temporary config file")
	}

	// Rename temporary file to actual config file (atomic operation)
	if err := os.Rename(tmpFile, w.configPath); err != nil {
		return errors.Wrap(err, "failed to rename temporary config file")
	}

	log.WithField("path", w.configPath).Info("configuration file updated")
	return nil
}

// backupConfig creates a backup of the current config file
func (w *Writer) backupConfig() error {
	// Check if config file exists
	if _, err := os.Stat(w.configPath); os.IsNotExist(err) {
		return nil // No file to backup
	}

	// Read current config
	data, err := os.ReadFile(w.configPath)
	if err != nil {
		return errors.Wrap(err, "failed to read config for backup")
	}

	// Write backup with timestamp
	backupPath := w.configPath + ".backup"
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return errors.Wrap(err, "failed to write config backup")
	}

	log.WithField("path", backupPath).Debug("created config backup")
	return nil
}

// UpdatePartial updates a specific section of the config file
// This reads the current config, updates the specified section, and writes it back
// If the config file doesn't exist, it creates a new one
func (w *Writer) UpdatePartial(updateFn func(tree *toml.Tree) error) error {
	var tree *toml.Tree
	var err error

	// Read current config as TOML tree, or create empty tree if file doesn't exist
	data, err := os.ReadFile(w.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty tree if file doesn't exist
			log.WithField("path", w.configPath).Info("config file doesn't exist, creating new one")
			tree, err = toml.TreeFromMap(make(map[string]interface{}))
			if err != nil {
				return errors.Wrap(err, "failed to create empty config tree")
			}
		} else {
			return errors.Wrap(err, "failed to read config file")
		}
	} else {
		tree, err = toml.LoadBytes(data)
		if err != nil {
			return errors.Wrap(err, "failed to parse config TOML")
		}
	}

	// Apply the update function
	if err := updateFn(tree); err != nil {
		return errors.Wrap(err, "failed to update config")
	}

	// Ensure config directory exists
	configDir := filepath.Dir(w.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create config directory")
	}

	// Create backup
	if err := w.backupConfig(); err != nil {
		log.WithError(err).Warn("failed to create config backup")
	}

	// Marshal back to bytes
	var buf []byte
	buf, err = tree.Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to marshal updated config")
	}

	// Write to temporary file
	tmpFile := w.configPath + ".tmp"
	if err := os.WriteFile(tmpFile, buf, 0644); err != nil {
		return errors.Wrap(err, "failed to write temporary config file")
	}

	// Rename to actual file (atomic)
	if err := os.Rename(tmpFile, w.configPath); err != nil {
		return errors.Wrap(err, "failed to rename temporary config file")
	}

	log.WithField("path", w.configPath).Info("configuration partially updated")
	return nil
}

// GetConfigDir returns the directory containing the config file
func (w *Writer) GetConfigDir() string {
	return filepath.Dir(w.configPath)
}
