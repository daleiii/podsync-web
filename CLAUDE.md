# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Podsync is a Go-based service that converts YouTube, Vimeo, and SoundCloud channels into podcast feeds. It downloads video/audio content and generates RSS feeds that can be consumed by podcast clients.

**Key Features:**
- Modern React-based web UI for managing feeds and monitoring downloads
- REST API for programmatic access
- Real-time download progress tracking
- No-config startup support
- Optional basic authentication
- Episode filtering including automatic ignoring of short episodes (under 60 seconds)

## Key Architecture Components

### Main Application (`cmd/podsync/`)
- **main.go**: Entry point with CLI argument parsing, signal handling, and service orchestration
- **config.go**: TOML configuration loading and validation with defaults

### Core Packages (`pkg/`)
- **builder/**: Media downloaders for different platforms (YouTube, Vimeo, SoundCloud)
- **feed/**: RSS/podcast feed generation and management, OPML export
- **db/**: BadgerDB-based storage for metadata and state
- **fs/**: Storage abstraction supporting local filesystem and S3-compatible storage
- **model/**: Core data structures and domain models
- **ytdl/**: YouTube-dl wrapper for media downloading
- **progress/**: Real-time download progress tracking system

### Services (`services/`)
- **api/**: REST API server with handlers for feeds, episodes, config, and progress
  - **handlers/**: HTTP handlers for API endpoints (feeds, episodes, config, progress)
  - **middleware/**: Basic authentication and other middleware
  - **router.go**: API route definitions
- **update/**: Feed update orchestration and scheduling
- **web/**: HTTP server for serving podcast feeds, media files, and web UI

### Frontend (`frontend/`)
- **React-based web UI** built with Vite and TypeScript
- **src/pages/**: Dashboard, Episodes, and Settings pages
- **src/components/**: Reusable UI components including DownloadProgress
- **src/services/**: API client for backend communication
- **src/types/**: TypeScript type definitions
- Built files are placed in `html/` and served by the Go backend

### Key Dependencies

**Backend:**
- youtube-dl/yt-dlp for media downloading
- BadgerDB for local storage
- go-toml for configuration
- robfig/cron for scheduling
- AWS SDK for S3 storage
- gorilla/mux for HTTP routing

**Frontend:**
- React with TypeScript
- Vite for build tooling
- TailwindCSS for styling
- Axios for API requests

## Common Development Commands

### Building
```bash
make build          # Build binary to bin/podsync
make                # Build and run tests
```

### Testing
```bash
make test           # Run all unit tests
go test -v ./...    # Run tests with verbose output
go test ./pkg/...   # Test specific packages
```

### Linting and Formatting
```bash
golangci-lint run   # Run all configured linters and formatters
gofmt -s -w .       # Format all Go files
goimports -w .      # Organize imports and format
```

### Running
```bash
./bin/podsync                         # Run without config (web UI available at http://localhost:8080)
./bin/podsync --config config.toml    # Run with config file
./bin/podsync --debug                 # Run with debug logging
./bin/podsync --headless              # Run once and exit (no web server)
```

### Frontend Development
```bash
cd frontend
npm install                           # Install dependencies
npm run dev                           # Start Vite dev server on http://localhost:5173
npm run build                         # Build for production (outputs to html/)
npm run lint                          # Run ESLint
```

**Note:** When developing the frontend, run both the backend (on port 8080) and the Vite dev server (on port 5173) simultaneously. The Vite dev server proxies API requests to the backend.

### Docker
```bash
make docker                           # Build local Docker image
docker run -it --rm localhost/podsync:latest
```

### Development Debugging

#### Backend Debugging
Use VS Code with the Go extension. The repository includes `.vscode/launch.json` with a "Debug Podsync" configuration that runs with `config.toml`.

#### Frontend Debugging
Use browser DevTools with the Vite dev server running. Source maps are enabled in development mode.

## Configuration

The application uses TOML configuration files. See `config.toml.example` for all available options. Key sections:
- `[server]`: Web server settings (port, hostname, TLS, username/password for basic auth)
- `[storage]`: Local or S3 storage configuration
- `[tokens]`: API keys for YouTube/Vimeo
- `[feeds]`: Feed definitions with URLs and settings
- `[downloader]`: youtube-dl configuration

**No-config startup:** The application can now start without a config file. Feeds and settings can be configured through the web UI at `http://localhost:8080`.

## Development Guidelines

### Code Quality

**Backend (Go):**
- Write clean, idiomatic Go code following Go conventions and best practices
- Use structured logging with logrus for consistent log formatting
- Ensure proper error handling and meaningful error messages
- Follow the existing code style and patterns in the repository

**Frontend (React/TypeScript):**
- Write type-safe TypeScript code with proper type definitions
- Use React hooks and functional components
- Follow React best practices for component composition
- Use TailwindCSS utility classes for styling
- Keep components focused and reusable
- Handle loading and error states appropriately in API calls

### Testing and Quality Assurance

**Backend:**
- **CRITICAL**: Always run ALL of the following commands before making a commit or opening a PR:
  1. `go fmt ./...` - Format all Go files
  2. `golangci-lint run` - Run all configured linters and formatters
  3. `make test` - Run all unit tests
- Run tests first with `make test` to ensure functionality works correctly
- Run linter with `golangci-lint run` to ensure proper formatting and code quality
- Ensure ALL tests pass AND ALL linting checks pass before committing
- The project uses golangci-lint with strict formatting rules - code must pass ALL checks

**Frontend:**
- Run `npm run lint` in the `frontend/` directory to check for TypeScript/ESLint issues
- Test the UI manually in the browser after making changes
- Verify that the build completes successfully with `npm run build`
- Check console for errors and warnings during development

**General:**
- Review code carefully for spelling errors, typos, and grammatical mistakes
- Test changes locally with different configurations when applicable
- For full-stack changes, test both frontend and backend interactions

### Git Workflow
- Keep commit messages brief and to the point
- Use a short, descriptive commit title (50 characters or less)
- Include a brief commit body that summarizes changes in 1-3 sentences when needed (wrap at 120 characters)
- Do not include automated signatures or generation notices in commit messages or pull requests
- Don't add "Generated with Claude Code" to commit messages or pull request descriptions
- Don't add "Co-Authored-By: Claude noreply@anthropic.com" to commit messages or pull request descriptions
- Keep commits focused and atomic - one logical change per commit
- Ensure the build passes before pushing commits

### Pull Request Guidelines
- Keep PR descriptions concise and focused
- Include the brief commit body summary plus relevant examples if applicable
- Avoid verbose sections like "Changes Made", "Test Plan", or extensive bullet lists
- Focus on what the change does and why, not exhaustive implementation details
- Include code examples only when they help demonstrate usage or key functionality

## Key Conventions

**Backend:**
- Configuration validation happens at startup
- Graceful shutdown with context cancellation
- Storage abstraction allows switching between local/S3
- API key rotation support for rate limiting
- Cron-based scheduling for feed updates
- Episode filtering and cleanup capabilities
- Server-Sent Events (SSE) for real-time progress updates
- Optional basic authentication middleware for web UI and API

**Frontend:**
- API communication through centralized service layer (`src/services/api.ts`)
- Type-safe API responses with TypeScript interfaces (`src/types/api.ts`)
- Real-time updates via Server-Sent Events for download progress
- Responsive design with TailwindCSS
- Component-based architecture with clear separation of concerns

## Formatting and Linting Requirements

**Backend (Go):**

This project uses golangci-lint with strict formatting rules configured in `.golangci.yml`. Common formatting requirements include:

- Proper spacing around operators (`if condition {` not `if(condition){`)
- Correct struct field alignment and spacing
- Proper import ordering (standard library, third-party, local packages)
- No trailing whitespace
- Consistent spacing around assignment operators (`key: value` not `key:value`)
- Space after commas in function parameters and struct literals

**Always run `go fmt ./...`, `golangci-lint run`, AND `make test` after making ANY code changes to ensure both functionality and formatting are correct before committing.**

**Frontend (TypeScript/React):**

The frontend uses ESLint and TypeScript compiler for linting. Common requirements:

- Proper TypeScript types for all variables, function parameters, and return values
- No unused variables or imports
- Consistent code style following Airbnb/Standard conventions
- React hooks dependencies must be properly specified
- Prefer const over let where possible

**Always run `npm run lint` in the `frontend/` directory after making frontend changes.**