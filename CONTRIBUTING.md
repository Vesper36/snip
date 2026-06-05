# Contributing to Snip

Thank you for your interest in contributing to Snip! Here's how to get started.

## Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Vesper36/snip.git
   cd snip
   ```

2. **Install Go 1.22+**
   ```bash
   # macOS
   brew install go
   # Linux (Ubuntu/Debian)
   sudo apt-get install golang-go
   ```

3. **Build and run**
   ```bash
   make build
   make run
   ```

4. **Run tests**
   ```bash
   make test
   ```

## Project Structure

```
snip/
  cmd/server/          # Entry point, route registration
  internal/
    config/            # Environment-based configuration
    models/            # Data structures
    store/             # SQLite layer (WAL, pure-Go driver)
    auth/              # JWT, bcrypt, API tokens
    i18n/              # Embedded translations (EN/ZH/JA/FR/DE)
    middleware/         # Rate limit, auth, i18n, security
    handlers/          # HTTP handlers + embedded templates/static
      templates/       # Go html/template (per-page instances)
      static/          # CSS, JS, manifest, robots, sitemap
  scripts/             # CLI tool + install script
```

## Making Changes

1. **Fork and create a branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes** following the code style

3. **Run tests and lint**
   ```bash
   make test
   make lint
   ```

4. **Commit with a descriptive message**
   ```
   feat: add new feature
   fix: resolve issue with X
   docs: update README
   ```

5. **Push and create a Pull Request**

## Code Style

- Follow Go conventions and idioms
- Use `go vet` for static analysis
- Keep functions focused and small
- Write tests for new features
- Add i18n keys for all user-facing strings

## Adding a New Language

1. Create `internal/i18n/XX.json` with translations
2. Update `internal/i18n/i18n.go` to embed the new file
3. Add the language label to the `Labels` map

## Reporting Issues

- Use GitHub Issues
- Include steps to reproduce
- Include Go version and OS
- Include relevant logs

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
