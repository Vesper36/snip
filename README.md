<p align="center">
  <img src="https://img.icons8.com/?size=120&id=11138&format=png&color=3b82f6" width="120" alt="Snip">
</p>

<h1 align="center">Snip</h1>
<p align="center">
  <em>Self-hosted code & file sharing, batteries included.</em>
</p>

<p align="center">
  <a href="https://snip.vps.vesper366.com"><img src="https://img.shields.io/badge/demo-live-brightgreen?style=flat-square" alt="Live Demo"></a>
  <a href="https://github.com/Vesper36/snip/stargazers"><img src="https://img.shields.io/github/stars/Vesper36/snip?style=flat-square" alt="Stars"></a>
  <a href="https://github.com/Vesper36/snip/blob/main/LICENSE"><img src="https://img.shields.io/github/license/Vesper36/snip?style=flat-square" alt="License"></a>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/db-SQLite-003B57?style=flat-square&logo=sqlite" alt="SQLite">
  <a href="https://github.com/Vesper36/snip/actions"><img src="https://img.shields.io/github/actions/workflow/status/Vesper36/snip/ci.yml?style=flat-square" alt="CI"></a>
</p>

<p align="center">
  <a href="#why-snip">Why Snip</a> &bull;
  <a href="#features">Features</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#api">API</a> &bull;
  <a href="#deployment">Deployment</a> &bull;
  <a href="#configuration">Config</a>
</p>

---

> **Live demo:** https://snip.vps.vesper366.com

Snip is a lightweight, self-hosted platform for sharing code snippets, text, and files. Single Go binary with embedded assets -- zero external dependencies, runs anywhere. Like GitHub Gist, but yours.

## Why Snip?

| Other tools | Snip |
|-------------|------|
| GitHub Gist requires account, no self-hosting | Self-hosted, anonymous by default |
| transfer.sh is unmaintained | Active development |
| PrivateBin is PHP + JS heavy | Single 13MB Go binary |
| hastebin is dead | REST API, modern UI |
| Private, not customizable | Full source, MIT licensed |

## Features

**Core**
- 40+ languages with auto-detection syntax highlighting
- Auto expiry: 10min, 1h, 1d, 1w, 1m
- Burn-after-reading self-destructing pastes
- bcrypt password protection
- Custom short URLs
- View limits
- Drag & drop file upload
- QR code for mobile sharing
- Paste fork / duplicate
- Full-text search

**Production**
- REST API with API tokens
- Prometheus metrics at `/metrics`
- Health check at `/healthz`
- i18n: English, Chinese, Japanese, French, German
- Dark & light themes
- Single binary, zero deps
- SQLite (WAL mode)
- Docker + docker-compose
- Rate limiting, CSP headers
- Systemd service integration
- Cross-platform: Linux, macOS, Windows

**Developer experience**
- Keyboard shortcuts (`Tab` indent, `Ctrl+Enter` submit, `/` search)
- Image paste from clipboard
- CLI tool for shell piping
- One-line install script
- Auto-updating with GitHub releases

## Quick Start

### Try the live demo

[https://snip.vps.vesper366.com](https://snip.vps.vesper366.com)

### One-line install

```bash
curl -L https://github.com/Vesper36/snip/releases/latest/download/install.sh | bash
```

### Docker

```bash
docker run -d -p 53524:53524 --name snip vesper/snip
```

### From source

```bash
git clone https://github.com/Vesper36/snip.git
cd snip
go build -o snip ./cmd/server
./snip
```

Open http://localhost:53524

## CLI Usage

Pipe from stdin or file:

```bash
# From stdin
echo "console.log('hello')" | ./scripts/snip.sh --lang javascript --expire 1h

# From file
./scripts/snip.sh main.go --expire 1d --title "My Go file"
```

## API

### Create a paste

```bash
curl -X POST http://localhost:53524/api/v1/pastes \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package main\nfunc main() { println(\"hi\") }",
    "language": "go",
    "title": "Hello world",
    "expires_in": "1d"
  }'
```

Response:

```json
{
  "slug": "abc12345",
  "url": "http://localhost:53524/abc12345",
  "raw_url": "http://localhost:53524/abc12345/raw",
  "language": "go",
  "views": 0
}
```

### Get a paste

```bash
curl http://localhost:53524/api/v1/pastes/abc12345
```

### Search

```bash
curl http://localhost:53524/api/v1/search?q=hello
```

### Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/pastes` | - | Create a paste |
| `GET` | `/api/v1/pastes` | - | List pastes |
| `GET` | `/api/v1/pastes/{slug}` | - | Get a paste |
| `DELETE` | `/api/v1/pastes/{slug}` | Token | Delete a paste |
| `GET` | `/api/v1/pastes/{slug}/raw` | - | Raw content |
| `GET` | `/api/v1/search?q=` | - | Search pastes |
| `GET` | `/api/v1/stats` | - | Get statistics |
| `GET` | `/healthz` | - | Health check |
| `GET` | `/metrics` | - | Prometheus metrics |
| `POST` | `/api/v1/tokens` | Token | Create API token |
| `DELETE` | `/api/v1/tokens/{id}` | Token | Revoke API token |

### Authentication

Generate an API token from the Settings page, then use it:

```bash
curl -H "Authorization: snip_xxxxx" http://localhost:53524/api/v1/pastes
```

### File upload

```bash
curl -X POST http://localhost:53524/api/v1/pastes \
  -H "Authorization: snip_xxxxx" \
  -F "file=@main.go" \
  -F "language=go" \
  -F "title=My Go file"
```

## Deployment

### Docker Compose (recommended)

```yaml
version: "3.8"
services:
  snip:
    image: vesper/snip:latest
    container_name: snip
    ports:
      - "53524:53524"
    volumes:
      - snip-data:/app/data
    environment:
      - SNIP_BASE_URL=https://snip.yourdomain.com
      - SNIP_JWT_SECRET=replace-with-random-32-bytes
    restart: unless-stopped

volumes:
  snip-data:
```

### Systemd (Linux)

The install script auto-creates `/etc/systemd/system/snip.service`:

```ini
[Unit]
Description=Snip - code & file sharing
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/snip
Environment=SNIP_PORT=53524
Environment=SNIP_DB_PATH=/var/lib/snip/snip.db
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

```bash
systemctl status snip
journalctl -u snip -f
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: snip
spec:
  replicas: 1
  selector:
    matchLabels: { app: snip }
  template:
    metadata:
      labels: { app: snip }
    spec:
      containers:
      - name: snip
        image: vesper/snip:latest
        ports: [{ containerPort: 53524 }]
        livenessProbe:
          httpGet: { path: /healthz, port: 53524 }
        readinessProbe:
          httpGet: { path: /healthz, port: 53524 }
        volumeMounts:
        - name: data
          mountPath: /app/data
      volumes:
      - name: data
        persistentVolumeClaim: { claimName: snip-data }
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIP_HOST` | `0.0.0.0` | Listen host |
| `SNIP_PORT` | `53524` | Listen port |
| `SNIP_BASE_URL` | `http://localhost:53524` | Public URL (used in share links) |
| `SNIP_DB_PATH` | `./data/snip.db` | SQLite database path |
| `SNIP_ADMIN_PASSWORD` | (empty) | Admin password (optional) |
| `SNIP_JWT_SECRET` | (auto) | JWT signing key (set in production!) |
| `SNIP_MAX_SIZE` | `10485760` | Max paste size in bytes (10MB) |
| `SNIP_ALLOW_ANONYMOUS` | `true` | Allow anonymous paste creation |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Ctrl`/`Cmd` + `Enter` | Submit paste form |
| `Tab` | Indent 4 spaces in editor |
| `/` | Focus search box |
| `Esc` | Blur input |

## Architecture

```
snip/
  cmd/server/          Entry point, route registration
  internal/
    config/            Environment-based configuration
    models/            Data structures
    store/             SQLite layer (WAL, pure-Go driver)
    auth/              JWT, bcrypt, API tokens
    i18n/              Embedded EN/ZH translations
    middleware/         Rate limit, auth, i18n, security
    handlers/          HTTP handlers + embedded templates/static
      templates/       Go html/template (per-page instances)
      static/          CSS, JS, manifest, robots, sitemap
  scripts/             CLI tool + install script
  .github/workflows/   CI + release automation
```

**Stack:** Go 1.21+, Chi router, modernc.org/sqlite (pure Go), html/template, HTMX, Highlight.js

## Contributing

PRs welcome! Standard Go workflow:

```bash
go test ./...
go vet ./...
```

## License

MIT © 2024 Vesper