<p align="center">
  <img src="https://img.icons8.com/?size=96&id=11138&format=png&color=3b82f6" width="96" alt="Snip Logo">
</p>

<h1 align="center">Snip</h1>
<p align="center">
  <strong>Self-hosted code & file sharing platform</strong><br>
  <em>Like GitHub Gist, but self-hosted, faster, and private.</em>
</p>

<p align="center">
  <a href="https://github.com/Vesper36/snip/stargazers"><img src="https://img.shields.io/github/stars/Vesper36/snip?style=flat-square" alt="Stars"></a>
  <a href="https://github.com/Vesper36/snip/blob/main/LICENSE"><img src="https://img.shields.io/github/license/Vesper36/snip?style=flat-square" alt="License"></a>
  <img src="https://img.shields.io/badge/language-Go-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/db-SQLite-003B57?style=flat-square&logo=sqlite" alt="SQLite">
</p>

<p align="center">
  <a href="#features">Features</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#api">API</a> &bull;
  <a href="#docker">Docker</a> &bull;
  <a href="#configuration">Config</a>
</p>

---

Snip is a lightweight, self-hosted platform for sharing code snippets, text, and files. Single Go binary with embedded assets -- zero dependencies, zero config, just run.

## Features

| Feature | Description |
|---------|-------------|
| Syntax Highlighting | 40+ languages with auto-detection |
| Auto Expiry | Pastes expire after 10min, 1h, 1d, 1w, or 1m |
| Burn After Reading | Self-destructing pastes deleted after first view |
| Password Protection | bcrypt-hashed password protection |
| Custom URLs | Choose your own short URL slug |
| View Limits | Set maximum number of views |
| File Upload | Drag & drop files or use the file picker |
| QR Code | Generate QR code for mobile sharing |
| i18n | English and Chinese interface |
| REST API | Full API with token-based authentication |
| Dark Mode | Beautiful dark theme, responsive design |
| Single Binary | 13MB, embed all assets, no dependencies |
| Docker Ready | One-command deployment |

## Quick Start

### Pre-built Binary

```bash
# Download from releases
curl -L https://github.com/Vesper36/snip/releases/latest/download/snip-linux-amd64 -o snip
chmod +x snip
./snip
```

Open http://localhost:53524

### From Source

```bash
git clone https://github.com/Vesper36/snip.git
cd snip
make build
make run
```

### Docker

```bash
docker-compose up -d
```

## API

Create a paste:

```bash
curl -X POST http://localhost:53524/api/v1/pastes \
  -H "Content-Type: application/json" \
  -d '{"content": "fmt.Println(\"Hello!\")", "language": "go"}'
```

Get a paste:

```bash
curl http://localhost:53524/api/v1/pastes/{slug}
```

Search:

```bash
curl http://localhost:53524/api/v1/search?q=hello
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/pastes` | Create a paste |
| `GET` | `/api/v1/pastes` | List pastes |
| `GET` | `/api/v1/pastes/{slug}` | Get a paste |
| `DELETE` | `/api/v1/pastes/{slug}` | Delete a paste (requires auth) |
| `GET` | `/api/v1/pastes/{slug}/raw` | Raw content |
| `GET` | `/api/v1/search?q=` | Search pastes |
| `GET` | `/api/v1/stats` | Get statistics |
| `POST` | `/api/v1/tokens` | Create API token (requires auth) |
| `DELETE` | `/api/v1/tokens/{id}` | Revoke API token (requires auth) |

### Authentication

Generate an API token from the Settings page, then include it in requests:

```bash
curl -H "Authorization: snip_xxxxx" http://localhost:53524/api/v1/pastes
```

### File Upload via API

```bash
curl -X POST http://localhost:53524/api/v1/pastes \
  -H "Authorization: snip_xxxxx" \
  -F "file=@main.go" \
  -F "language=go"
```

## Docker

### Docker Compose

```yaml
version: "3.8"
services:
  snip:
    build: .
    ports:
      - "53524:53524"
    volumes:
      - snip-data:/app/data
    environment:
      - SNIP_BASE_URL=https://snip.yourdomain.com
      - SNIP_JWT_SECRET=your-secret-here
    restart: unless-stopped

volumes:
  snip-data:
```

### Docker Run

```bash
docker run -d \
  --name snip \
  -p 53524:53524 \
  -v snip-data:/app/data \
  vesper/snip
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIP_HOST` | `0.0.0.0` | Listen host |
| `SNIP_PORT` | `53524` | Listen port |
| `SNIP_BASE_URL` | `http://localhost:53524` | Public URL |
| `SNIP_DB_PATH` | `./data/snip.db` | Database path |
| `SNIP_ADMIN_PASSWORD` | (empty) | Admin password |
| `SNIP_JWT_SECRET` | (auto) | JWT signing key |
| `SNIP_MAX_SIZE` | `10485760` | Max paste size (10MB) |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Ctrl` + `Enter` | Submit paste |
| `Tab` | Insert 4 spaces |
| `/` | Focus search |
| `Esc` | Blur input |

## Architecture

```
snip/
  cmd/server/          Entry point
  internal/
    config/            Env-based configuration
    models/            Data models
    store/             SQLite layer (WAL mode)
    auth/              JWT + bcrypt
    i18n/              Embedded translations
    middleware/         Rate limit, auth, i18n
    handlers/          HTTP handlers + templates + static
```

**Stack:** Go, Chi, SQLite (modernc pure-go), html/template, HTMX, Highlight.js

## License

MIT