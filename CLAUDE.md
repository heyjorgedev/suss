# CLAUDE.md

This file provides guidance for AI assistants working with the suss codebase.

## Project Overview

**suss** is a fast, simple, and privacy-friendly URL shortener built with Go. It provides a web interface for shortening URLs, generating QR codes, and managing shortened links with secret keys for owner access.

## Tech Stack

- **Language**: Go 1.25+
- **HTTP Router**: [chi](https://github.com/go-chi/chi) v5
- **Database**: SQLite with [go-sqlite3](https://github.com/mattn/go-sqlite3)
- **Templating**: [templ](https://github.com/a-h/templ) - type-safe HTML templates
- **Styling**: TailwindCSS v4
- **QR Codes**: [go-qrcode](https://github.com/skip2/go-qrcode)
- **Asset Hashing**: [hashfs](https://github.com/benbjohnson/hashfs) for cache-busting
- **Hot Reload**: [Air](https://github.com/cosmtrek/air) for development

## Project Structure

```
suss/
├── cmd/suss/main.go      # Application entry point, config, and Program struct
├── suss.go               # Package-level version variables
├── shorturl.go           # Domain types: ShortURL struct and ShortURLService interface
├── error.go              # Application error types and codes
├── link.go               # Link type (placeholder)
├── http/                 # HTTP layer
│   ├── server.go         # HTTP server, router setup, middleware chain
│   ├── shorturl.go       # HTTP handlers for URL shortening endpoints
│   ├── middleware.go     # Custom middleware (host context)
│   ├── error.go          # HTTP error handling
│   ├── html/             # Templ templates
│   │   ├── base.templ    # HTML document structure, head, body
│   │   ├── layout.templ  # Header, footer components
│   │   ├── homepage.templ
│   │   ├── manage.templ
│   │   ├── preview.templ
│   │   ├── error.templ
│   │   └── svg.templ     # Logo SVG component
│   ├── assets/           # Source assets
│   │   └── tailwind.css  # TailwindCSS entry point
│   └── dist/             # Built/static assets
│       ├── dist.go       # Embedded filesystem with hashfs
│       ├── css/          # Built CSS (gitignored)
│       ├── favicon/      # Favicon assets
│       └── img/          # Images (og-image.png)
├── sqlite/               # SQLite data layer
│   ├── sqlite.go         # DB connection, configuration, migrations
│   ├── shorturl.go       # ShortURLService implementation
│   ├── tx.go             # Transaction wrapper
│   ├── time.go           # NullTime type for RFC3339 time handling
│   └── migration/        # SQL migration files
│       └── 00000000.sql  # Initial schema (short_urls table)
├── .air.toml             # Air hot reload configuration
├── Dockerfile            # Multi-stage Docker build
├── go.mod / go.sum       # Go dependencies
└── package.json          # Node dependencies for TailwindCSS
```

## Architecture Patterns

### Domain-Driven Design
- **Domain types** defined in root package (`suss/`)
- **Interfaces** defined alongside domain types (e.g., `ShortURLService`)
- **Implementations** in separate packages (`sqlite/`, `http/`)

### Dependency Injection
The `Program` struct in `cmd/suss/main.go` wires dependencies:
```go
type Program struct {
    Config          *Config
    DB              *sqlite.DB
    HTTPServer      *http.Server
    ShortURLService suss.ShortURLService
}
```

### Error Handling
Application errors use typed errors with codes:
- `ECONFLICT`, `EINTERNAL`, `EINVALID`, `ENOTFOUND`, `EUNAUTHORIZED`
- Use `suss.Errorf(code, format, args...)` to create errors
- Use `suss.ErrorCode(err)` and `suss.ErrorMessage(err)` to extract info

## Development Commands

### Prerequisites
```bash
# Install Go 1.25+
# Install Node.js (for TailwindCSS)
npm install
```

### Build Commands

```bash
# Generate templ files (creates *_templ.go from *.templ)
go tool templ generate

# Build TailwindCSS
npm run build

# Build the application
go build -o ./tmp/main cmd/suss/main.go

# Run with SQLite file
DB_DSN=./tmp/database.sqlite ./tmp/main
```

### Development with Hot Reload

```bash
# Install Air (if not installed)
go install github.com/cosmtrek/air@latest

# Run with hot reload (uses .air.toml config)
air
```

Air automatically:
1. Runs `go tool templ generate` on file changes
2. Runs `npm run build` for CSS changes
3. Rebuilds and restarts the Go binary

### Docker Build

```bash
docker build -t suss .
docker run -p 8080:8080 suss
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_DSN` | `:memory:` | SQLite database path (use file path for persistence) |
| `HOSTNAME` | `0.0.0.0` | HTTP server bind address |
| `PORT` | `8080` | HTTP server port |
| `HASH_KEY` | `` | Hash key for future use |
| `ENCRYPTION_KEY` | `` | Encryption key for future use |

## Key Files to Understand

### Adding New Features

1. **New Domain Type**: Add to root package, define interface
2. **Database Layer**: Add to `sqlite/` package, implement interface
3. **HTTP Handlers**: Add to `http/shorturl.go` or new file
4. **Templates**: Add `.templ` files to `http/html/`
5. **Routes**: Register in `http/server.go` NewServer()

### Database Migrations

Add new SQL files to `sqlite/migration/` with naming convention `NNNNNNNN.sql` (8-digit prefix for ordering). Migrations run automatically on startup.

## HTTP Routes

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/` | `handlerHomepage` | Homepage with URL shortening form |
| POST | `/shorten` | `handlerShortUrlCreate` | Create short URL (rate-limited: 5/min) |
| GET | `/preview/{slug}` | `handlerShortUrlPreview` | Preview page before redirect |
| GET | `/{slug}+` | `handlerShortUrlPreview` | Alternative preview URL |
| GET | `/{slug}` | `handlerShortUrlVisit` | Redirect to long URL |
| GET | `/manage/{slug}` | `handlerShortUrlManage` | Management page (requires secret) |
| GET | `/qrcode/{slug}.png` | `handlerShortUrlQrCode` | Generate QR code PNG |
| GET | `/assets/*` | hashfs FileServer | Static assets with cache-busting |

## Templ Templates

Templates use the [templ](https://templ.guide/) library for type-safe HTML:

```go
// Define template with parameters
templ ManagePage(props ManagePageProps) {
    @html() {
        @head() { ... }
        @body() { ... }
    }
}

// Render in handler
html.ManagePage(props).Render(r.Context(), w)
```

After modifying `.templ` files, regenerate with:
```bash
go tool templ generate
```

## Code Conventions

### Go Style
- Standard Go formatting (`gofmt`)
- Errors wrapped with context: `fmt.Errorf("cannot open db: %w", err)`
- Interfaces defined in domain package, implementations in infrastructure packages

### HTTP Handlers
- Return `http.HandlerFunc` from factory functions
- Use chi URL params: `chi.URLParam(r, "slug")`
- Render errors via `s.Error(w, r, err)`

### Database
- All queries use parameterized statements
- Transactions via `db.BeginTx(ctx, nil)`
- Times stored as RFC3339 strings

### Templates
- Composable components: `@component()` syntax
- Children via `{ children... }`
- Props as struct types for complex data

## Testing

No test files currently exist. When adding tests:
- Place `*_test.go` files alongside source files
- Air excludes `*_test.go` from hot reload triggers

## Common Tasks

### Add a new page
1. Create `http/html/newpage.templ`
2. Run `go tool templ generate`
3. Add handler in `http/shorturl.go`
4. Register route in `http/server.go`

### Add a database field
1. Create new migration file in `sqlite/migration/`
2. Update `suss.ShortURL` struct
3. Update `sqlite/shorturl.go` queries

### Modify styles
1. Edit templates or `http/assets/tailwind.css`
2. Run `npm run build` (or let Air handle it)
