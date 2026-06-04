<p align="center">
  <img src="https://img.icons8.com/?size=96&id=11138&format=png&color=58a6ff" alt="Snip Logo" width="96">
</p>

<h1 align="center">Snip</h1>

<p align="center">
  <strong>Self-hosted code & file sharing platform</strong>
</p>

<p align="center">
  Like GitHub Gist, but self-hosted, faster, and private.
</p>

<p align="center">
  <a href="#features">Features</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#api">API</a> &bull;
  <a href="#docker">Docker</a> &bull;
  <a href="#configuration">Configuration</a>
</p>

---

Snip is a lightweight, self-hosted platform for sharing code snippets, text, and files. Built with Go as a single binary with embedded assets -- no external dependencies, no JavaScript build step, no Node.js required.

## Features

- **Syntax Highlighting** -- 40+ languages with automatic detection
- **Auto Expiry** -- Pastes expire after 10min, 1h, 1d, 1w, or 1m
- **Burn After Reading** -- Self-destructing pastes deleted after first view
- **Password Protection** -- Protect sensitive content with bcrypt-hashed passwords
- **Custom URLs** -- Choose your own short URL slug
- **View Limits** -- Set maximum number of views
- **REST API** -- Full API with token-based authentication
- **Search** -- Full-text search across all pastes
- **Dark Mode** -- Beautiful dark theme out of the box
- **Single Binary** -- One file, zero dependencies, embed all assets
- **SQLite Storage** -- No external database required
- **Docker Ready** -- One-command deployment

## Quick Start

### From Source

```bash
git clone https://github.com/vesper/snip.git
cd snip
make build
make run
```

Open http://localhost:53524

### Docker

```bash
docker-compose up -d
```

### Pre-built Binary

Download from [Releases](https://github.com/vesper/snip/releases) and run:

```bash
./snip
```

## API

Snip provides a REST API for programmatic access.

### Create a Paste

```bash
curl -X POST http://localhost:53524/api/v1/pastes \
  -H "Content-Type: application/json" \
  -d '{
    "content": "fmt.Println(\"Hello, World!\")",
    "language": "go",
    "title": "Hello World",
    "expires_in": "1d"
  }'
```

### Get a Paste

```bash
curl http://localhost:53524/api/v1/pastes/{slug}
```

### Search

```bash
curl http://localhost:53524/api/v1/search?q=hello
```

### Authentication

Generate an API token from the Settings page, then include it in requests:

```bash
curl -H "Authorization: snip_xxxxx" http://localhost:53524/api/v1/pastes
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/pastes` | Create a paste |
| `GET` | `/api/v1/pastes` | List pastes |
| `GET` | `/api/v1/pastes/{slug}` | Get a paste |
| `DELETE` | `/api/v1/pastes/{slug}` | Delete a paste |
| `GET` | `/api/v1/pastes/{slug}/raw` | Raw content |
| `GET` | `/api/v1/search?q=` | Search pastes |
| `GET` | `/api/v1/stats` | Get statistics |
| `POST` | `/api/v1/tokens` | Create API token |
| `DELETE` | `/api/v1/tokens/{id}` | Revoke API token |

## Docker

### Docker Compose (Recommended)

```yaml
version: "3.8"
services:
  snip:
    build: .
    ports:
      - "3000:3000"
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
  -e SNIP_BASE_URL=https://snip.yourdomain.com \
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
| `SNIP_MAX_SIZE` | `10485760` | Max paste size (bytes) |

## Architecture

```
snip/
  cmd/server/          -- Entry point
  internal/
    config/            -- Configuration loading
    models/            -- Data models
    store/             -- SQLite database layer
    auth/              -- JWT + bcrypt authentication
    middleware/         -- HTTP middleware (auth, rate-limit, security)
    handlers/          -- HTTP handlers + embedded templates/static
      templates/       -- Go HTML templates
      static/          -- CSS, JS assets
```

**Tech Stack:**
- Go + Chi Router
- SQLite (WAL mode)
- Go html/template
- Highlight.js (syntax highlighting)
- HTMX (interactivity)

## License

MIT
