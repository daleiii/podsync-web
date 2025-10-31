# Podsync Fork Comparison

This document compares our enhanced fork of Podsync with the original [mxpv/podsync](https://github.com/mxpv/podsync) repository.

## Summary

This fork adds a modern React-based web UI, comprehensive API improvements, TLS support, enhanced configuration management, and various quality-of-life features while maintaining full backward compatibility with the original project.

## Major Enhancements

### 🌐 Modern Web UI

**Original Project:**
- CLI-only configuration via TOML files
- No graphical interface
- Manual editing of config files required
- No visual feedback for downloads

**Our Fork:**
- Full-featured React-based web UI
- Real-time download progress tracking with visual progress bars
- Interactive settings management
- Feed management through web interface
- Episode browsing and management
- Built with Vite, TypeScript, and TailwindCSS
- Responsive design for mobile and desktop

### 🔌 REST API Enhancements

**Original Project:**
- Basic feed serving via HTTP
- Limited API endpoints
- No configuration API

**Our Fork:**
- Comprehensive REST API for all operations
- Configuration management via API endpoints:
  - `GET/PUT /api/v1/config/*` - Server, storage, downloader, tokens, auth, history
  - `POST /api/v1/config/restart` - Server restart
  - `POST /api/v1/config/tls/upload` - TLS certificate upload
- Feed management:
  - `GET /api/v1/feeds` - List all feeds
  - `POST /api/v1/feeds` - Create new feed
  - `GET/PUT/DELETE /api/v1/feeds/{id}` - Manage specific feeds
  - `POST /api/v1/feeds/{id}/refresh` - Manual feed refresh
- Episode management:
  - `GET /api/v1/episodes` - List episodes
  - `DELETE /api/v1/episodes/{feed_id}/{episode_id}` - Delete episode
  - `POST /api/v1/episodes/{feed_id}/{episode_id}/retry` - Retry failed download
  - `POST /api/v1/episodes/{feed_id}/{episode_id}/block` - Block episode
- Real-time progress tracking:
  - `GET /api/v1/progress` - Current download progress
  - `GET /api/v1/progress/stream` - Server-Sent Events stream
- History tracking:
  - `GET /api/v1/history` - Job history
  - `GET /api/v1/history/stats` - Statistics
  - `POST /api/v1/history/cleanup` - Cleanup old entries
  - `DELETE /api/v1/history` - Clear all history

### 🔒 Security & Authentication

**Original Project:**
- No built-in authentication
- HTTP only

**Our Fork:**
- HTTP Basic Authentication support
- TLS/HTTPS support with certificate upload via web UI
- Password-protected settings
- Secure credential storage with show/hide toggles
- Optional authentication for web UI and API

### ⚙️ Configuration Management

**Original Project:**
- TOML file configuration only
- Manual editing required
- No runtime configuration changes
- Limited validation

**Our Fork:**
- Web UI-based configuration
- API-driven configuration updates
- Runtime configuration changes (some require restart)
- Environment variable support with precedence
- No-config startup mode (configure everything via UI)
- Configuration validation and feedback
- TLS certificate upload functionality

### 📊 Download Progress & History

**Original Project:**
- No progress tracking
- Logs only
- No download history

**Our Fork:**
- Real-time download progress tracking
- Visual progress bars in web UI
- Server-Sent Events for live updates
- Comprehensive job history tracking
- Download statistics and analytics
- Configurable history retention
- Failed download retry functionality
- Episode blocking capability

### 🎨 User Experience

**Original Project:**
- Command-line focused
- Config file editing
- Log file monitoring
- Manual management

**Our Fork:**
- Intuitive web interface
- Point-and-click configuration
- Visual feedback and notifications
- Dashboard with feed overview
- Episode browser with filtering
- Settings page with organized sections
- Help text and tooltips
- Responsive mobile-friendly design

### 🐳 Docker Improvements

**Original Project:**
- Basic Dockerfile
- Simple docker-compose example
- Limited documentation

**Our Fork:**
- Multi-stage Docker build
- Frontend build integrated in Docker
- Comprehensive docker-compose examples
- Named volumes for data persistence
- Healthcheck configuration
- Environment variable documentation
- Both config-file and config-less modes
- TZ support for proper timezone handling

### 📦 Storage Options

**Original Project:**
- Local filesystem storage
- Basic S3 support

**Our Fork:**
- Local filesystem storage (enhanced)
- Full S3-compatible storage support
- S3 credentials configurable via web UI
- MinIO, DigitalOcean Spaces, Wasabi support
- IAM role support
- Storage configuration without restart

### 🔧 Developer Experience

**Original Project:**
- Go-focused
- Basic makefile
- Limited documentation

**Our Fork:**
- Full-stack development setup
- Frontend development with hot-reload
- Comprehensive build system
- Extensive documentation
- TypeScript type safety
- Linting and formatting configured
- VSCode debug configuration
- GitHub Actions ready (workflow template included)

### 📝 Configuration Options

**New Configuration Sections:**

```toml
[server.basic_auth]
  enabled = true
  username = "admin"
  password = "secure-password"

[storage.s3]
  access_key = "AKIAIOSFODNN7EXAMPLE"
  secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

[history]
  enabled = true
  retention_days = 30
  max_entries = 1000
```

**New Environment Variables:**

- `PODSYNC_WEB_UI` - Enable/disable web UI
- `PODSYNC_HISTORY_ENABLED` - Enable history tracking
- `PODSYNC_HISTORY_RETENTION_DAYS` - History retention period
- `PODSYNC_HISTORY_MAX_ENTRIES` - Max history entries
- `PODSYNC_CONFIG_PATH` - Config file path override

### 🐛 Bug Fixes

**Issues Fixed in Our Fork:**

1. **History Tracking Configuration Bug**: Fixed issue where history tracking would always be enabled even when explicitly disabled with default values
2. **Download Timeout Input**: Fixed browser console error with timeout field accepting string values ("30s") instead of numbers
3. **Frontend Port Field**: Hidden in production builds, only shown during development
4. **Settings Page Toggle**: Replaced complex CSS toggle that caused page crashes with standard checkbox
5. **S3 Credential Handling**: Proper reading and writing of S3 credentials in configuration

## Feature Matrix

| Feature | Original | Our Fork |
|---------|----------|----------|
| **Web UI** | ✗ | ✓ |
| **REST API** | Basic | Comprehensive |
| **Real-time Progress** | ✗ | ✓ |
| **Download History** | ✗ | ✓ |
| **HTTP Basic Auth** | ✗ | ✓ |
| **TLS/HTTPS Support** | ✗ | ✓ |
| **No-Config Startup** | ✗ | ✓ |
| **Configuration via UI** | ✗ | ✓ |
| **Episode Management** | ✗ | ✓ |
| **Failed Download Retry** | ✗ | ✓ |
| **Episode Blocking** | ✗ | ✓ |
| **S3 UI Configuration** | ✗ | ✓ |
| **Certificate Upload** | ✗ | ✓ |
| **Server-Sent Events** | ✗ | ✓ |
| **TypeScript Frontend** | ✗ | ✓ |
| **Mobile Responsive** | ✗ | ✓ |
| **Docker Multi-arch** | ✓ | ✓ |
| **TOML Configuration** | ✓ | ✓ |
| **Environment Variables** | Limited | Extensive |
| **yt-dlp Auto-update** | ✓ | ✓ |
| **Feed Filtering** | ✓ | ✓ |
| **Custom Metadata** | ✓ | ✓ |
| **OPML Export** | ✓ | ✓ |
| **Cron Scheduling** | ✓ | ✓ |

## Backward Compatibility

**100% Backward Compatible:**

- All original TOML configuration options work unchanged
- Existing feeds continue to work
- RSS feed URLs remain the same
- Command-line arguments unchanged
- Environment variables extended (original ones still work)
- Docker compose files compatible
- Database format unchanged

**Optional Enhancements:**

All new features are optional and can be disabled:
- Web UI can be disabled via `PODSYNC_WEB_UI=false`
- History tracking can be disabled in config
- Authentication is optional
- Original CLI-only workflow still fully supported

## Migration Guide

**Upgrading from Original Podsync:**

1. **No Changes Required** - Just replace the Docker image or binary
2. **Optional**: Access web UI at `http://localhost:8080`
3. **Optional**: Enable authentication in config
4. **Optional**: Configure S3 credentials via UI
5. **Optional**: Set up history tracking

**Existing Configuration:**

```toml
# Your existing config.toml works as-is
[server]
  port = 8080
  hostname = "https://my.domain.com"

[feeds.ID1]
  url = "https://www.youtube.com/channel/..."
  format = "audio"
```

**Enhanced Configuration:**

```toml
# Add optional enhancements
[server]
  port = 8080
  hostname = "https://my.domain.com"
  tls = true
  certificate_path = "/path/to/cert.pem"
  key_file_path = "/path/to/key.pem"

  [server.basic_auth]
    enabled = true
    username = "admin"
    password = "secure-password"

[history]
  enabled = true
  retention_days = 30
  max_entries = 1000

[feeds.ID1]
  url = "https://www.youtube.com/channel/..."
  format = "audio"
```

## Performance Considerations

**Build Time:**
- Frontend build adds ~1-2 minutes to Docker build time
- Caching optimizes subsequent builds

**Runtime Performance:**
- Web UI adds minimal memory overhead (~50MB)
- API endpoints are lightweight
- Progress tracking uses efficient SSE
- History tracking has negligible impact

**Storage:**
- Docker image is ~100MB larger (includes frontend assets)
- History database is small (~1-10MB typically)
- No impact on media file storage

## Technology Stack

### Backend (Unchanged from Original)
- Go 1.25
- BadgerDB for metadata
- yt-dlp for downloads
- FFmpeg for conversion

### Frontend (New)
- React 18
- TypeScript
- Vite build tool
- TailwindCSS
- Axios for API calls
- Server-Sent Events for real-time updates

### DevOps
- Docker multi-stage builds
- GitHub Actions ready
- Make-based build system
- golangci-lint for Go quality
- ESLint for TypeScript quality

## Future Enhancements

Potential additions being considered:

- [ ] OAuth2 authentication
- [ ] Multi-user support
- [ ] Feed templates
- [ ] Bulk operations
- [ ] Advanced filtering UI
- [ ] Webhook notifications
- [ ] Metrics and monitoring dashboard
- [ ] Plugin system
- [ ] RSS feed preview
- [ ] Media player integration

## Contributing

This fork is actively maintained and accepts contributions. See the main README.md for development setup and contribution guidelines.

## Credits

- **Original Project**: [mxpv/podsync](https://github.com/mxpv/podsync) by @mxpv
- **Fork Enhancements**: Enhanced UI, API, and features by Dale Larson (@daleiii)
- **Community**: Thank you to all contributors and users!

## License

This project maintains the same MIT License as the original Podsync project.
