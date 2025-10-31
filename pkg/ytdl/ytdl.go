package ytdl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daleiii/podsync-web/pkg/feed"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/daleiii/podsync-web/pkg/model"
)

const (
	DefaultDownloadTimeout = 10 * time.Minute
	UpdatePeriod           = 24 * time.Hour
)

type PlaylistMetadataThumbnail struct {
	Id         string `json:"id"`
	Url        string `json:"url"`
	Resolution string `json:"resolution"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

type PlaylistMetadata struct {
	Id          string                      `json:"id"`
	Title       string                      `json:"title"`
	Description string                      `json:"description"`
	Thumbnails  []PlaylistMetadataThumbnail `json:"thumbnails"`
	Channel     string                      `json:"channel"`
	ChannelId   string                      `json:"channel_id"`
	ChannelUrl  string                      `json:"channel_url"`
	WebpageUrl  string                      `json:"webpage_url"`
}

var (
	ErrTooManyRequests = errors.New(http.StatusText(http.StatusTooManyRequests))
)

// ProgressCallback is called during download to report progress
// stage: "downloading", "encoding", etc.
// percent: 0-100
// downloaded: bytes downloaded so far
// total: total size in bytes (may be 0 if unknown)
// speed: speed string like "1.2MiB/s"
type ProgressCallback func(stage string, percent float64, downloaded, total int64, speed string)

// Config is a youtube-dl related configuration
type Config struct {
	// SelfUpdate toggles self update every 24 hour
	SelfUpdate bool `toml:"self_update"`
	// UpdateChannel specifies which channel to use for updates: "stable", "nightly", or "master"
	UpdateChannel string `toml:"update_channel"`
	// UpdateVersion locks to a specific version (format: channel@tag or just tag)
	// Example: "stable@2023.07.06" or "2023.10.07"
	// Leave empty for latest version
	UpdateVersion string `toml:"update_version"`
	// Timeout in minutes for youtube-dl process to finish download
	Timeout int `toml:"timeout"`
	// CustomBinary is a custom path to youtube-dl, this allows using various youtube-dl forks.
	CustomBinary string `toml:"custom_binary"`
}

type YoutubeDl struct {
	path             string
	timeout          time.Duration
	updateChannel    string     // Update channel: stable, nightly, or master
	updateVersion    string     // Specific version to lock to (optional)
	updateLock       sync.Mutex // Don't call youtube-dl while self updating
	progressCallback ProgressCallback
}

func New(ctx context.Context, cfg Config) (*YoutubeDl, error) {
	var (
		path string
		err  error
	)

	if cfg.CustomBinary != "" {
		path = cfg.CustomBinary

		// Don't update custom youtube-dl binaries.
		log.Warnf("using custom youtube-dl binary, turning self updates off")
		cfg.SelfUpdate = false
	} else {
		path, err = exec.LookPath("youtube-dl")
		if err != nil {
			return nil, errors.Wrap(err, "youtube-dl binary not found")
		}

		log.Debugf("found youtube-dl binary at %q", path)
	}

	// Set default update channel if not specified
	if cfg.UpdateChannel == "" {
		cfg.UpdateChannel = "stable"
	}

	timeout := DefaultDownloadTimeout
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Minute
	}

	log.Debugf("download timeout: %d min(s)", int(timeout.Minutes()))

	ytdl := &YoutubeDl{
		path:          path,
		timeout:       timeout,
		updateChannel: cfg.UpdateChannel,
		updateVersion: cfg.UpdateVersion,
	}

	// Make sure youtube-dl exists
	version, err := ytdl.exec(ctx, "--version")
	if err != nil {
		return nil, errors.Wrap(err, "could not find youtube-dl")
	}

	log.Infof("using youtube-dl %s", version)

	if err := ytdl.ensureDependencies(ctx); err != nil {
		return nil, err
	}

	if cfg.SelfUpdate {
		// Do initial blocking update at launch
		if err := ytdl.Update(ctx); err != nil {
			log.WithError(err).Error("failed to update youtube-dl")
		}

		go func() {
			for {
				time.Sleep(UpdatePeriod)

				if err := ytdl.Update(context.Background()); err != nil {
					log.WithError(err).Error("update failed")
				}
			}
		}()
	}

	return ytdl, nil
}

func (dl *YoutubeDl) ensureDependencies(ctx context.Context) error {
	found := false

	if path, err := exec.LookPath("ffmpeg"); err == nil {
		found = true

		output, err := exec.CommandContext(ctx, path, "-version").CombinedOutput()
		if err != nil {
			return errors.Wrap(err, "could not get ffmpeg version")
		}

		log.Infof("found ffmpeg: %s", output)
	}

	if path, err := exec.LookPath("avconv"); err == nil {
		found = true

		output, err := exec.CommandContext(ctx, path, "-version").CombinedOutput()
		if err != nil {
			return errors.Wrap(err, "could not get avconv version")
		}

		log.Infof("found avconv: %s", output)
	}

	if !found {
		return errors.New("either ffmpeg or avconv required to run Podsync")
	}

	return nil
}

func (dl *YoutubeDl) Version(ctx context.Context) (string, error) {
	return dl.exec(ctx, "--version")
}

func (dl *YoutubeDl) Update(ctx context.Context) error {
	dl.updateLock.Lock()
	defer dl.updateLock.Unlock()

	// Build the update command based on channel and version settings
	var args []string

	if dl.updateVersion != "" {
		// Update to a specific version
		log.Infof("updating youtube-dl to version %s", dl.updateVersion)
		args = []string{"--update-to", dl.updateVersion, "--verbose"}
	} else if dl.updateChannel != "" && dl.updateChannel != "stable" {
		// Update to a specific channel (nightly or master)
		log.Infof("updating youtube-dl to %s channel", dl.updateChannel)
		args = []string{"--update-to", dl.updateChannel, "--verbose"}
	} else {
		// Default update (stable channel)
		log.Info("updating youtube-dl to latest stable version")
		args = []string{"--update", "--verbose"}
	}

	output, err := dl.exec(ctx, args...)
	if err != nil {
		log.WithError(err).Error(output)
		return errors.Wrap(err, "failed to self update youtube-dl")
	}

	log.Info(output)
	return nil
}

func (dl *YoutubeDl) PlaylistMetadata(ctx context.Context, url string) (metadata PlaylistMetadata, err error) {
	log.Info("getting playlist metadata for: ", url)
	args := []string{
		"--playlist-items", "0",
		"-J",            // JSON output
		"-q",            // quiet mode
		"--no-warnings", // suppress warnings
		url,
	}
	dl.updateLock.Lock()
	defer dl.updateLock.Unlock()
	output, err := dl.exec(ctx, args...)
	if err != nil {
		log.WithError(err).Errorf("youtube-dl error: %s", url)

		// YouTube might block host with HTTP Error 429: Too Many Requests
		if strings.Contains(output, "HTTP Error 429") {
			return PlaylistMetadata{}, ErrTooManyRequests
		}

		log.Error(output)
		return PlaylistMetadata{}, errors.New(output)
	}

	var playlistMetadata PlaylistMetadata
	json.Unmarshal([]byte(output), &playlistMetadata)
	return playlistMetadata, nil
}

// SetProgressCallback sets the callback to be called during downloads
func (dl *YoutubeDl) SetProgressCallback(callback ProgressCallback) {
	dl.progressCallback = callback
}

func (dl *YoutubeDl) Download(ctx context.Context, feedConfig *feed.Config, episode *model.Episode) (r io.ReadCloser, err error) {
	tmpDir, err := os.MkdirTemp("", "podsync-")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get temp dir for download")
	}

	defer func() {
		if err != nil {
			err1 := os.RemoveAll(tmpDir)
			if err1 != nil {
				log.Errorf("could not remove temp dir: %v", err1)
			}
		}
	}()

	// filePath with YoutubeDl template format
	filePath := filepath.Join(tmpDir, fmt.Sprintf("%s.%s", episode.ID, "%(ext)s"))

	args := buildArgs(feedConfig, episode, filePath)

	dl.updateLock.Lock()
	defer dl.updateLock.Unlock()

	output, err := dl.execWithProgress(ctx, args...)
	if err != nil {
		log.WithError(err).Errorf("youtube-dl error: %s", filePath)

		// YouTube might block host with HTTP Error 429: Too Many Requests
		if strings.Contains(output, "HTTP Error 429") {
			return nil, ErrTooManyRequests
		}

		log.Error(output)

		return nil, errors.New(output)
	}

	ext := "mp4"
	if feedConfig.Format == model.FormatAudio {
		ext = "mp3"
	}
	if feedConfig.Format == model.FormatCustom {
		ext = feedConfig.CustomFormat.Extension
	}

	// filePath now with the final extension
	filePath = filepath.Join(tmpDir, fmt.Sprintf("%s.%s", episode.ID, ext))
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open downloaded file")
	}

	return &tempFile{File: f, dir: tmpDir}, nil
}

func (dl *YoutubeDl) exec(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, dl.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, dl.path, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), errors.Wrap(err, "failed to execute youtube-dl")
	}

	return string(output), nil
}

// execWithProgress runs youtube-dl and parses progress output
func (dl *YoutubeDl) execWithProgress(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, dl.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, dl.path, args...)

	// Capture stderr for progress parsing
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", errors.Wrap(err, "failed to create stderr pipe")
	}

	// Capture stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", errors.Wrap(err, "failed to create stdout pipe")
	}

	if err := cmd.Start(); err != nil {
		return "", errors.Wrap(err, "failed to start youtube-dl")
	}

	// Parse progress from stderr in a goroutine
	var outputBuilder strings.Builder
	stderrDone := make(chan struct{})

	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuilder.WriteString(line)
			outputBuilder.WriteString("\n")

			// Parse progress if callback is set
			if dl.progressCallback != nil {
				dl.parseProgressLine(line)
			}
		}
	}()

	// Read stdout (mostly empty for downloads)
	stdoutScanner := bufio.NewScanner(stdout)
	for stdoutScanner.Scan() {
		line := stdoutScanner.Text()
		outputBuilder.WriteString(line)
		outputBuilder.WriteString("\n")
	}

	// Wait for command to finish
	err = cmd.Wait()
	<-stderrDone

	output := outputBuilder.String()
	if err != nil {
		return output, errors.Wrap(err, "failed to execute youtube-dl")
	}

	return output, nil
}

// parseProgressLine parses a single line of yt-dlp output for progress information
// Example lines:
// [download]   45.2% of 10.50MiB at 1.23MiB/s ETA 00:04
// [download] 100% of 10.50MiB in 00:08
// [ffmpeg] Destination: /tmp/file.mp3
func (dl *YoutubeDl) parseProgressLine(line string) {
	// Pattern for download progress: [download]   45.2% of 10.50MiB at 1.23MiB/s ETA 00:04
	downloadPattern := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)%\s+of\s+(\d+\.?\d*)(MiB|KiB|GiB|B)(?:\s+at\s+(\d+\.?\d*)(MiB|KiB|GiB|B)/s)?`)

	// Pattern for encoding: [ffmpeg] or [ExtractAudio]
	encodingPattern := regexp.MustCompile(`\[(ffmpeg|ExtractAudio|VideoConvertor)\]`)

	if matches := downloadPattern.FindStringSubmatch(line); matches != nil {
		// Parse percentage
		percent, _ := strconv.ParseFloat(matches[1], 64)

		// Parse total size
		totalSize, _ := strconv.ParseFloat(matches[2], 64)
		totalUnit := matches[3]
		totalBytes := convertToBytes(totalSize, totalUnit)

		// Calculate downloaded bytes
		downloadedBytes := int64(float64(totalBytes) * percent / 100.0)

		// Parse speed
		speed := ""
		if len(matches) >= 5 && matches[4] != "" {
			speedValue := matches[4]
			speedUnit := matches[5]
			speed = speedValue + speedUnit + "/s"
		}

		dl.progressCallback("downloading", percent, downloadedBytes, totalBytes, speed)
	} else if encodingPattern.MatchString(line) {
		// Encoding/post-processing stage - report as 100% downloading, now encoding
		dl.progressCallback("encoding", 100, 0, 0, "")
	}
}

// convertToBytes converts size with unit to bytes
func convertToBytes(size float64, unit string) int64 {
	switch unit {
	case "B":
		return int64(size)
	case "KiB":
		return int64(size * 1024)
	case "MiB":
		return int64(size * 1024 * 1024)
	case "GiB":
		return int64(size * 1024 * 1024 * 1024)
	default:
		return int64(size)
	}
}

func buildArgs(feedConfig *feed.Config, episode *model.Episode, outputFilePath string) []string {
	var args []string

	switch feedConfig.Format {
	case model.FormatVideo:
		// Video, mp4, high by default

		format := "bestvideo[ext=mp4][vcodec^=avc1]+bestaudio[ext=m4a]/best[ext=mp4][vcodec^=avc1]/best[ext=mp4]/best"

		if feedConfig.Quality == model.QualityLow {
			format = "worstvideo[ext=mp4][vcodec^=avc1]+worstaudio[ext=m4a]/worst[ext=mp4][vcodec^=avc1]/worst[ext=mp4]/worst"
		} else if feedConfig.Quality == model.QualityHigh && feedConfig.MaxHeight > 0 {
			format = fmt.Sprintf("bestvideo[height<=%d][ext=mp4][vcodec^=avc1]+bestaudio[ext=m4a]/best[height<=%d][ext=mp4][vcodec^=avc1]/best[ext=mp4]/best", feedConfig.MaxHeight, feedConfig.MaxHeight)
		}

		args = append(args, "--format", format)

	case model.FormatAudio:
		// Audio, mp3, high by default
		format := "bestaudio"
		if feedConfig.Quality == model.QualityLow {
			format = "worstaudio"
		}

		args = append(args, "--extract-audio", "--audio-format", "mp3", "--format", format)

	default:
		args = append(args, "--audio-format", feedConfig.CustomFormat.Extension, "--format", feedConfig.CustomFormat.YouTubeDLFormat)
	}

	// Insert additional per-feed youtube-dl arguments
	args = append(args, feedConfig.YouTubeDLArgs...)

	// Enable progress output for parsing by the progress callback
	args = append(args, "--progress", "--newline")

	args = append(args, "--output", outputFilePath, episode.VideoURL)
	return args
}
