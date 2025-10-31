package fs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// LocalConfig is the storage configuration for local file system
type LocalConfig struct {
	DataDir string `toml:"data_dir"`
}

// Local implements local file storage
type Local struct {
	rootDir      string
	WebUIEnabled bool
}

func NewLocal(rootDir string, webUIEnabled bool) (*Local, error) {
	return &Local{rootDir: rootDir, WebUIEnabled: webUIEnabled}, nil
}

func (l *Local) Open(name string) (http.File, error) {
	// Serve Web UI assets from html/ directory
	if l.WebUIEnabled {
		// For root directory, open the html directory itself
		if name == "/" {
			return os.Open("./html")
		}
		// Try to serve from html/ directory first for UI assets
		htmlPath := filepath.Join("./html", name)
		if file, err := os.Open(htmlPath); err == nil {
			return file, nil
		}
	}
	// Fall back to serving from data directory (for feeds, episodes, etc.)
	path := filepath.Join(l.rootDir, name)
	return os.Open(path)
}

func (l *Local) Delete(_ctx context.Context, name string) error {
	path := filepath.Join(l.rootDir, name)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	return nil
}

func (l *Local) Create(_ctx context.Context, name string, reader io.Reader) (int64, error) {
	var (
		logger = log.WithField("name", name)
		path   = filepath.Join(l.rootDir, name)
	)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return 0, errors.Wrapf(err, "failed to mkdir: %s", path)
	}

	logger.Infof("creating file: %s", path)
	written, err := l.copyFile(reader, path)
	if err != nil {
		return 0, errors.Wrap(err, "failed to copy file")
	}

	logger.Debugf("written %d bytes", written)
	return written, nil
}

func (l *Local) copyFile(source io.Reader, destinationPath string) (int64, error) {
	dest, err := os.Create(destinationPath)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create destination file")
	}

	defer dest.Close()

	written, err := io.Copy(dest, source)
	if err != nil {
		return 0, errors.Wrap(err, "failed to copy data")
	}

	return written, nil
}

func (l *Local) Size(_ctx context.Context, name string) (int64, error) {
	file, err := l.Open(name)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}
