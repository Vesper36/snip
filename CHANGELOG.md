# Changelog

All notable changes to Snip will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-06-05

### Added
- Core paste sharing functionality with syntax highlighting
- Auto-expiry support (10 min, 1 hour, 1 day, 1 week, 1 month)
- Burn-after-reading mode for self-destructing pastes
- Password protection with bcrypt encryption
- Custom short URLs and view limits
- Drag & drop file upload
- QR code generation for mobile sharing
- Paste fork/duplicate functionality
- Full-text search with pagination
- REST API with API tokens
- Prometheus metrics at `/metrics`
- Health check at `/healthz`
- i18n support (English, Chinese, Japanese, French, German)
- Dark & light theme toggle with localStorage persistence
- Single binary deployment with embedded assets
- SQLite database with WAL mode
- Docker and docker-compose support
- Rate limiting (120 req/min per IP)
- Security headers (CSP, X-Frame-Options, etc.)
- Keyboard shortcuts (Tab indent, Ctrl+Enter submit, / search)
- Image paste from clipboard
- CLI tool for shell piping
- One-line install script
- Systemd service integration
- Cross-platform builds (Linux, macOS, Windows)
- GitHub Actions CI/CD
- Comprehensive README with deployment guides
- Security policy and contributing guidelines
- GitHub issue and PR templates

### Security
- Password bypass prevention on /raw and /download endpoints
- Fork handler requires password for protected pastes
- Content-Disposition header injection protection
- SQL LIKE wildcard escaping in search queries
- Client-side QR code generation (no external API dependency)
- Versioned schema migration system
- Database connection pooling (MaxOpenConns=5)
- Request logging with proper status codes
- CORS support for API routes

### Fixed
- Error template i18n rendering ({{T}} → {{call .T}})
- Light theme nav bar color (hardcoded dark → CSS variable)
- Go module version mismatch (1.25.0 → 1.22)
- Graceful shutdown with proper cleanup goroutine termination

## [Unreleased]

### Planned
- Admin panel with JWT authentication
- Database backup endpoint
- WebSocket for real-time updates
- Multi-user support with team workspaces
- Public/private paste visibility settings
- Paste templates for common snippets
- Email notifications for paste expiry
- Webhook integration for paste events
- Advanced analytics dashboard
- Mobile app (iOS/Android)
