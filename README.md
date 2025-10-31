# Podsync Web

![Podsync](docs/img/logo.png)

[![GitHub release](https://img.shields.io/github/v/release/daleiii/podsync-web)](https://github.com/daleiii/podsync-web/releases)
[![Docker Hub](https://img.shields.io/docker/pulls/deltathreed/podsync-web)](https://hub.docker.com/r/deltathreed/podsync-web)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**Podsync Web** is an enhanced fork of [Podsync](https://github.com/mxpv/podsync) that adds a modern web UI, comprehensive REST API, and powerful management features. Convert YouTube, Vimeo, and other video/audio platforms into podcast feeds with an intuitive web interface.

> **Credits:** This project is built on the excellent foundation of [mxpv/podsync](https://github.com/mxpv/podsync) by [@mxpv](https://github.com/mxpv). We've added significant enhancements while maintaining 100% backward compatibility.

## 🎯 What's New in This Fork

### Major Enhancements

- **🌐 Modern React Web UI** - Full-featured web interface for managing feeds, browsing episodes, and monitoring downloads
- **📊 Real-time Progress Tracking** - Visual progress bars and live status updates via Server-Sent Events
- **🔌 Comprehensive REST API** - Complete API for feed management, episode control, and configuration
- **📝 No-Config Startup** - Start immediately without a config file, configure everything via the web UI
- **🔒 Security Features** - HTTP Basic Authentication and TLS/HTTPS support with certificate upload
- **📈 Download History** - Track all downloads with statistics, retry failed episodes, and block unwanted content
- **⚙️ Live Configuration** - Update settings through the web UI without manual file editing
- **🐳 Enhanced Docker Support** - Improved Docker setup with healthchecks and comprehensive examples

For a detailed comparison with the original project, see [COMPARISON.md](COMPARISON.md).

## ✨ Features

### Web Interface & User Experience
- Modern React-based web UI with responsive design
- Dashboard with feed overview and status
- Episode browser with search and filtering
- Real-time download progress with visual progress bars
- Interactive settings management
- Mobile-friendly responsive design

### API & Integration
- Full REST API for all operations
- Feed management (create, update, delete, refresh)
- Episode management (list, delete, retry, block)
- Configuration API with live updates
- Server-Sent Events for real-time progress
- History tracking with statistics

### Security & Authentication
- Optional HTTP Basic Authentication
- TLS/HTTPS support
- Certificate upload via web UI
- Secure credential storage
- Password-protected settings

### Core Functionality (Inherited from Original)
- Works with YouTube, Vimeo, and SoundCloud
- Supports feeds configuration: video/audio, quality settings, max height
- MP3 encoding for audio feeds
- Update scheduler with cron expressions
- Episode filtering (title, duration, automatic ignoring of shorts)
- Feed customization (artwork, category, language, metadata)
- OPML export
- Episode cleanup (keep last N episodes)
- API key rotation for rate limiting
- Runs on Windows, macOS, Linux, and Docker
- ARM support
- Automatic yt-dlp updates

## 🚀 Quick Start

### Docker (Recommended)

**Pull from Docker Hub:**

```bash
docker pull deltathreed/podsync-web:latest
```

**Run without config file (easiest):**

```bash
docker run -d \
  --name podsync \
  -p 8080:8080 \
  -v podsync_data:/app/data \
  -v podsync_db:/app/db \
  -e TZ=America/Los_Angeles \
  --restart unless-stopped \
  deltathreed/podsync-web:latest
```

**Access the web UI:**

Open your browser to `http://localhost:8080` and configure feeds through the interface.

### Docker Compose (Recommended for Production)

Create `docker-compose.yml`:

```yaml
services:
  podsync:
    container_name: podsync
    image: deltathreed/podsync-web:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - podsync_data:/app/data
      - podsync_db:/app/db
      # Optional: mount config file if you prefer file-based configuration
      # - ./config.toml:/app/config.toml
    environment:
      - TZ=America/Los_Angeles
      # Optional: Configure API keys via environment variables
      # - PODSYNC_YOUTUBE_API_KEY=${YOUTUBE_API_KEY}
      # - PODSYNC_VIMEO_API_KEY=${VIMEO_API_KEY}
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

volumes:
  podsync_data:
  podsync_db:
```

Start with:

```bash
docker compose up -d
```

### Build from Source

```bash
git clone https://github.com/daleiii/podsync-web
cd podsync-web
make
./bin/podsync
```

Visit `http://localhost:8080` to use the web UI.

## 📖 Documentation

### Getting Started
- [How to get YouTube API Key](./docs/how_to_get_youtube_api_key.md)
- [How to get Vimeo API Token](./docs/how_to_get_vimeo_token.md)
- [Schedule updates with cron](./docs/cron.md)
- [QNAP NAS Setup Guide](./docs/how_to_setup_podsync_on_qnap_nas.md)

### Advanced Topics
- [Complete Feature Comparison](COMPARISON.md) - Detailed comparison with the original Podsync
- [Development Guide](CLAUDE.md) - For contributors and developers

## 🌐 Web UI

Once Podsync is running, access the web interface at `http://localhost:8080`:

- **Dashboard** - Overview of all feeds and their status
- **Episodes** - Browse and manage downloaded episodes
- **Settings** - Configure feeds, API keys, storage, and download preferences
- **Real-time Progress** - Monitor active downloads with progress bars

## 🐳 Docker Reference

### Docker Images

**Docker Hub:**
```bash
docker pull deltathreed/podsync-web:latest
```

### Docker Run Examples

**Basic usage (no config):**
```bash
docker run -d \
  --name podsync \
  -p 8080:8080 \
  -v podsync_data:/app/data \
  -v podsync_db:/app/db \
  deltathreed/podsync-web:latest
```

**With environment variables:**
```bash
docker run -d \
  --name podsync \
  -p 8080:8080 \
  -v podsync_data:/app/data \
  -v podsync_db:/app/db \
  -e TZ=America/Los_Angeles \
  -e PODSYNC_YOUTUBE_API_KEY="your_api_key" \
  -e PODSYNC_WEB_UI=true \
  --restart unless-stopped \
  deltathreed/podsync-web:latest
```

**With config file:**
```bash
docker run -d \
  --name podsync \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/db:/app/db \
  -v $(pwd)/config.toml:/app/config.toml \
  -e TZ=America/Los_Angeles \
  --restart unless-stopped \
  deltathreed/podsync-web:latest
```

### Docker Compose Examples

**Minimal setup:**
```yaml
services:
  podsync:
    container_name: podsync
    image: deltathreed/podsync-web:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - podsync_data:/app/data
      - podsync_db:/app/db
    environment:
      - TZ=America/Los_Angeles

volumes:
  podsync_data:
  podsync_db:
```

**Full configuration with all options:**
```yaml
services:
  podsync:
    container_name: podsync
    image: deltathreed/podsync-web:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - podsync_data:/app/data
      - podsync_db:/app/db
      # Optional: mount config file
      # - ./config.toml:/app/config.toml
    environment:
      # Timezone (recommended)
      - TZ=America/Los_Angeles

      # API Keys (space-separated for rotation)
      - PODSYNC_YOUTUBE_API_KEY=${YOUTUBE_API_KEY}
      - PODSYNC_VIMEO_API_KEY=${VIMEO_API_KEY}
      # - PODSYNC_SOUNDCLOUD_API_KEY=${SOUNDCLOUD_KEY}
      # - PODSYNC_TWITCH_API_KEY=${TWITCH_CLIENT_ID}:${TWITCH_CLIENT_SECRET}

      # Web UI and features
      - PODSYNC_WEB_UI=true
      - PODSYNC_HISTORY_ENABLED=true
      - PODSYNC_HISTORY_RETENTION_DAYS=30
      - PODSYNC_HISTORY_MAX_ENTRIES=1000

      # Optional: Config file path override
      # - PODSYNC_CONFIG_PATH=/app/config.toml
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

volumes:
  podsync_data:
  podsync_db:
```

**Behind reverse proxy (Traefik example):**
```yaml
services:
  podsync:
    container_name: podsync
    image: deltathreed/podsync-web:latest
    restart: unless-stopped
    volumes:
      - podsync_data:/app/data
      - podsync_db:/app/db
    environment:
      - TZ=America/Los_Angeles
      - PODSYNC_YOUTUBE_API_KEY=${YOUTUBE_API_KEY}
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.podsync.rule=Host(`podsync.yourdomain.com`)"
      - "traefik.http.routers.podsync.entrypoints=websecure"
      - "traefik.http.routers.podsync.tls.certresolver=letsencrypt"
      - "traefik.http.services.podsync.loadbalancer.server.port=8080"
    networks:
      - traefik

volumes:
  podsync_data:
  podsync_db:

networks:
  traefik:
    external: true
```

### Building Docker Images Locally

```bash
# Build for local architecture
docker buildx build -t localhost/podsync-web:latest .

# Build for multiple architectures
docker buildx build --platform linux/amd64,linux/arm64 -t localhost/podsync-web:latest .

# Build without cache (clean build)
docker buildx build --no-cache -t localhost/podsync-web:latest .

# Using make (recommended)
make docker
```

## 🌍 Environment Variables

All environment variables are optional and override values in `config.toml`:

| Variable | Description | Example | Default |
|----------|-------------|---------|---------|
| `PODSYNC_CONFIG_PATH` | Path to config file | `/app/config.toml` | `config.toml` |
| `PODSYNC_YOUTUBE_API_KEY` | YouTube API key(s), space-separated for rotation | `key1 key2 key3` | - |
| `PODSYNC_VIMEO_API_KEY` | Vimeo API key(s), space-separated for rotation | `key1 key2` | - |
| `PODSYNC_SOUNDCLOUD_API_KEY` | SoundCloud API key(s), space-separated | `key1 key2` | - |
| `PODSYNC_TWITCH_API_KEY` | Twitch credentials `CLIENT_ID:CLIENT_SECRET`, space-separated for multiple | `id1:secret1 id2:secret2` | - |
| `PODSYNC_WEB_UI` | Enable/disable web UI | `true`, `false`, `1`, `0` | `true` |
| `PODSYNC_HISTORY_ENABLED` | Enable history tracking | `true`, `false`, `1`, `0` | `true` |
| `PODSYNC_HISTORY_RETENTION_DAYS` | Days to keep history | `30` | `30` |
| `PODSYNC_HISTORY_MAX_ENTRIES` | Max history entries | `1000` | `1000` |
| `TZ` | Timezone for logs and scheduling | `America/Los_Angeles`, `Europe/London` | `UTC` |

**API Key Rotation:**

Multiple API keys can be provided (space-separated) for automatic rotation:

```bash
# Single key
PODSYNC_YOUTUBE_API_KEY="AIzaSyXXXXXXXXXXXXXXXX"

# Multiple keys for rotation (useful for rate limiting)
PODSYNC_YOUTUBE_API_KEY="key1 key2 key3"
```

## ⚙️ Configuration

### Configuration Methods

Podsync supports three configuration methods (in order of precedence):

1. **Web UI** - Configure everything through the browser (recommended)
2. **Environment Variables** - Override specific settings
3. **config.toml file** - Traditional file-based configuration

### No-Config Startup (Recommended)

Start Podsync without any configuration:

```bash
./bin/podsync
# or
docker run -p 8080:8080 deltathreed/podsync-web:latest
```

Access `http://localhost:8080` to configure feeds, API keys, and settings through the web UI.

### TOML Configuration File

Create a `config.toml` file for advanced configuration:

#### Minimal Configuration

```toml
[server]
port = 8080

[storage]
  [storage.local]
  data_dir = "/app/data"

[tokens]
youtube = "YOUR_YOUTUBE_API_KEY"

[feeds]
  [feeds.tech_channel]
  url = "https://www.youtube.com/channel/UCxC5Ls6DwqV0e-CYcAKkExQ"
  format = "audio"
  quality = "high"
```

#### Complete Configuration Reference

```toml
# =============================================================================
# Server Configuration
# =============================================================================
[server]
  # Public hostname for RSS feed URLs (optional, useful behind reverse proxy)
  hostname = "https://podsync.yourdomain.com"

  # Port for API and web UI (internal port, map with -p in Docker)
  port = 8080

  # Frontend dev server port (only used during development)
  frontend_port = 5173

  # Bind address: "*" for all interfaces, "127.0.0.1" for localhost only
  bind_address = "*"

  # Base path for all routes (e.g., "/podsync" for https://domain.com/podsync/)
  path = ""

  # Enable TLS/HTTPS
  tls = false
  certificate_path = "/path/to/cert.pem"
  key_file_path = "/path/to/key.pem"

  # Enable web UI (can also be controlled via PODSYNC_WEB_UI env var)
  web_ui = true

  # HTTP Basic Authentication (optional)
  [server.basic_auth]
    enabled = false
    username = "admin"
    password = "secure-password"

# =============================================================================
# Storage Configuration
# =============================================================================
[storage]
  # Storage type: "local" or "s3"
  type = "local"

  # Local filesystem storage
  [storage.local]
    data_dir = "/app/data"

  # S3-compatible storage (MinIO, AWS S3, DigitalOcean Spaces, Wasabi, etc.)
  [storage.s3]
    bucket = "my-podsync-bucket"
    region = "us-west-2"
    endpoint_url = "https://s3.us-west-2.amazonaws.com"
    prefix = "podsync/"  # Optional path prefix
    access_key = "AKIAIOSFODNN7EXAMPLE"  # Optional, can use IAM roles
    secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"  # Optional

# =============================================================================
# Database Configuration
# =============================================================================
[database]
  # BadgerDB directory for metadata storage
  dir = "/app/db"

# =============================================================================
# Downloader Configuration
# =============================================================================
[downloader]
  # Auto-update yt-dlp every 24 hours
  self_update = true

  # Update channel: "stable", "nightly", or "master"
  update_channel = "stable"

  # Lock to specific version (optional)
  # update_version = "2023.07.06"

  # Download timeout per episode (supports: 30s, 5m, 1h)
  timeout = "30m"

# =============================================================================
# API Tokens
# =============================================================================
[tokens]
  # API keys for video platforms (supports rotation with multiple keys)
  youtube = ["YOUR_YOUTUBE_API_KEY"]
  vimeo = ["YOUR_VIMEO_API_KEY"]
  soundcloud = ["YOUR_SOUNDCLOUD_KEY"]
  # Twitch format: ["CLIENT_ID:CLIENT_SECRET"]
  twitch = []

# =============================================================================
# History Tracking
# =============================================================================
[history]
  # Enable job history tracking
  enabled = true

  # Days to retain history
  retention_days = 30

  # Maximum history entries to keep
  max_entries = 1000

# =============================================================================
# Global Cleanup Policy
# =============================================================================
[cleanup]
  # Keep last N episodes per feed (applies to all feeds unless overridden)
  keep_last = 10

# =============================================================================
# Feed Definitions
# =============================================================================
[feeds]
  # Feed ID (used in URLs: /feeds/tech_channel.xml)
  [feeds.tech_channel]
    # Source URL (YouTube channel, playlist, Vimeo user, etc.)
    url = "https://www.youtube.com/channel/UCxC5Ls6DwqV0e-CYcAKkExQ"

    # Update frequency (supports: 1h, 12h, 24h, or cron expression)
    update_period = "12h"
    # Alternative: use cron expression
    # cron_schedule = "0 */6 * * *"

    # Output format: "audio" or "video"
    format = "audio"

    # Quality: "high" or "low"
    quality = "high"

    # Max video height (e.g., 720, 1080)
    max_height = 720

    # Number of episodes to fetch per update
    page_size = 50

    # Playlist sorting: "asc" (oldest first) or "desc" (newest first)
    playlist_sort = "desc"

    # Make feed private (requires authentication)
    private_feed = false

    # Include in OPML export
    opml = true

    # Feed-specific cleanup (overrides global cleanup)
    [feeds.tech_channel.clean]
      keep_last = 5

    # Content filters
    [feeds.tech_channel.filters]
      # Include only if title matches this regex
      title = ""

      # Exclude if title matches this regex
      not_title = ""

      # Include only if description matches
      description = ""

      # Exclude if description matches
      not_description = ""

      # Minimum episode duration (e.g., "1m", "30s")
      min_duration = "1m"

      # Maximum episode duration
      max_duration = ""

      # Minimum episode age
      min_age = ""

      # Maximum episode age (e.g., "30d" for last 30 days)
      max_age = "30d"

    # Custom feed metadata
    [feeds.tech_channel.custom]
      cover_art = "https://example.com/cover.jpg"  # Custom artwork URL
      cover_art_quality = "high"  # "high" or "low"
      category = "Technology"
      subcategories = ["Tech News", "Gadgets"]
      explicit = false
      language = "en"
      author = "Channel Name"
      title = "Custom Feed Title"
      description = "Custom feed description"
      owner_name = "Your Name"
      owner_email = "your@email.com"
      link = "https://your-website.com"

  # Additional feeds can be added with different IDs
  [feeds.music_podcast]
    url = "https://www.youtube.com/playlist?list=PLxxxxxxxxxxxxxx"
    format = "audio"
    quality = "high"
    update_period = "24h"
```

### Configuration with Reverse Proxy

If running behind a reverse proxy (nginx, Traefik, Caddy), set the `hostname` to your public URL:

```toml
[server]
port = 8080
hostname = "https://podsync.yourdomain.com"
```

Server will be accessible internally from `http://localhost:8080`, but RSS feed URLs will point to `https://podsync.yourdomain.com/feeds/...`

## 🔌 REST API

Podsync provides a comprehensive REST API. All endpoints require basic authentication if configured.

### Key Endpoints

**Configuration Management:**
- `GET /api/v1/config/server` - Get server configuration
- `PUT /api/v1/config/server` - Update server configuration
- `GET /api/v1/config/storage` - Get storage configuration
- `PUT /api/v1/config/storage` - Update storage configuration
- `GET /api/v1/config/tokens` - Get API tokens
- `PUT /api/v1/config/tokens` - Update API tokens
- `POST /api/v1/config/restart` - Restart server
- `POST /api/v1/config/tls/upload` - Upload TLS certificate

**Feed Management:**
- `GET /api/v1/feeds` - List all feeds
- `POST /api/v1/feeds` - Create new feed
- `GET /api/v1/feeds/{id}` - Get specific feed
- `PUT /api/v1/feeds/{id}` - Update feed
- `DELETE /api/v1/feeds/{id}` - Delete feed
- `POST /api/v1/feeds/{id}/refresh` - Manually refresh feed

**Episode Management:**
- `GET /api/v1/episodes?feed_id={id}` - List episodes for feed
- `DELETE /api/v1/episodes/{feed_id}/{episode_id}` - Delete episode
- `POST /api/v1/episodes/{feed_id}/{episode_id}/retry` - Retry failed download
- `POST /api/v1/episodes/{feed_id}/{episode_id}/block` - Block episode

**Progress & History:**
- `GET /api/v1/progress` - Get current download progress
- `GET /api/v1/progress/stream` - Server-Sent Events stream for real-time updates
- `GET /api/v1/history` - Get job history
- `GET /api/v1/history/stats` - Get statistics
- `POST /api/v1/history/cleanup` - Cleanup old entries
- `DELETE /api/v1/history` - Clear all history

### Example API Usage

```bash
# List all feeds
curl http://localhost:8080/api/v1/feeds

# Get episodes for a specific feed
curl http://localhost:8080/api/v1/episodes?feed_id=tech_channel

# With authentication
curl -u admin:password http://localhost:8080/api/v1/feeds

# Create a new feed
curl -X POST http://localhost:8080/api/v1/feeds \
  -H "Content-Type: application/json" \
  -d '{
    "id": "new_feed",
    "url": "https://www.youtube.com/channel/UCxxxxxx",
    "format": "audio",
    "quality": "high"
  }'

# Manually refresh a feed
curl -X POST http://localhost:8080/api/v1/feeds/tech_channel/refresh

# Monitor download progress (Server-Sent Events)
curl -N http://localhost:8080/api/v1/progress/stream
```

## 📋 Dependencies

### Backend Dependencies

If running the CLI as binary (not via Docker), ensure these are installed:

**On macOS:**
```bash
brew install yt-dlp ffmpeg go
```

**On Ubuntu/Debian:**
```bash
sudo apt install yt-dlp ffmpeg golang-go
```

### Frontend Dependencies (for development)

```bash
cd frontend
npm install
```

## 🔧 Development

### Backend Development

```bash
# Build
make build

# Run with debug logging
./bin/podsync --debug

# Run with config file
./bin/podsync --config config.toml

# Run tests
make test

# Run linter
golangci-lint run
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Start dev server with hot-reload (http://localhost:5173)
npm run dev

# Build for production (outputs to html/)
npm run build

# Run linter
npm run lint
```

**Note:** When developing the frontend, run both the backend (port 8080) and Vite dev server (port 5173) simultaneously. The Vite dev server proxies API requests to the backend.

### Debugging

**Backend:** Use VS Code with the Go extension. Launch configuration is in `.vscode/launch.json`.

**Frontend:** Use browser DevTools with the Vite dev server.

See [CLAUDE.md](CLAUDE.md) for detailed development guidelines.

## 🤝 Contributing

Contributions are welcome! This fork is actively maintained.

**Before submitting:**

1. Run `go fmt ./...` to format Go code
2. Run `golangci-lint run` to check code quality
3. Run `make test` to ensure all tests pass
4. Run `npm run lint` in `frontend/` for TypeScript/ESLint checks

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Credits

- **Original Project:** [mxpv/podsync](https://github.com/mxpv/podsync) by [@mxpv](https://github.com/mxpv) - Thank you for creating the excellent foundation that made this project possible!
- **Fork Enhancements:** Web UI, REST API, and additional features by [@daleiii](https://github.com/daleiii)
- **Community:** Thank you to all contributors and users!

## 🔗 Links

- **GitHub Repository:** https://github.com/daleiii/podsync-web
- **Docker Hub:** https://hub.docker.com/r/deltathreed/podsync-web
- **Original Podsync:** https://github.com/mxpv/podsync
- **Issue Tracker:** https://github.com/daleiii/podsync-web/issues

## 🆘 Support

- **Issues:** Report bugs at https://github.com/daleiii/podsync-web/issues
- **Documentation:** See [docs/](./docs/) directory for guides
- **API Documentation:** See REST API section above
- **Comparison:** See [COMPARISON.md](COMPARISON.md) for feature comparison
