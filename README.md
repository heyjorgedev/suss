# suss

A fast, simple, and privacy-friendly URL shortener built with Go.

## Features

- **URL Shortening** - Convert long URLs to short, memorable 6-character slugs
- **QR Code Generation** - Automatically generate QR codes for each shortened URL
- **Preview Pages** - View the target URL before redirecting (append `+` to any short URL)
- **Link Management** - Access a management dashboard using your secret key
- **Rate Limiting** - Built-in protection against abuse (5 requests/minute per IP)
- **Privacy-Friendly** - No tracking, no analytics, minimal data collection
- **Self-Hostable** - Easy deployment with Docker or as a standalone binary

## Quick Start

### Using Docker

```bash
docker build -t suss .
docker run -p 8080:8080 -v suss-data:/data -e DB_DSN=/data/suss.db suss
```

Visit `http://localhost:8080` to start shortening URLs.

### From Source

Prerequisites:
- Go 1.25+
- Node.js (for TailwindCSS)
- SQLite development libraries

```bash
# Install Node dependencies
npm install

# Generate template files
go tool templ generate

# Build CSS
npm run build

# Build and run
go build -o suss cmd/suss/main.go
DB_DSN=./database.sqlite ./suss
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_DSN` | `:memory:` | SQLite database path (use a file path for persistence) |
| `HOSTNAME` | `0.0.0.0` | HTTP server bind address |
| `PORT` | `8080` | HTTP server port |

## Development

Install [Air](https://github.com/cosmtrek/air) for hot reload during development:

```bash
go install github.com/cosmtrek/air@latest
air
```

Air automatically regenerates templates and rebuilds CSS on file changes.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Homepage with URL shortening form |
| `POST` | `/shorten` | Create a new short URL |
| `GET` | `/{slug}` | Redirect to the original long URL |
| `GET` | `/{slug}+` | Preview page before redirecting |
| `GET` | `/preview/{slug}` | Preview page (alternate path) |
| `GET` | `/manage/{slug}?secret=KEY` | Management page (requires secret key) |
| `GET` | `/qrcode/{slug}.png` | Generate QR code as PNG |

## Tech Stack

- **Go** with [chi](https://github.com/go-chi/chi) router
- **SQLite** for simple, file-based persistence
- **[templ](https://templ.guide/)** for type-safe HTML templates
- **TailwindCSS v4** for styling
- **[go-qrcode](https://github.com/skip2/go-qrcode)** for QR code generation

## Project Structure

```
suss/
├── cmd/suss/main.go      # Application entry point
├── shorturl.go           # Domain types and interfaces
├── error.go              # Application error handling
├── http/
│   ├── server.go         # HTTP server and routing
│   ├── shorturl.go       # URL shortening handlers
│   └── html/             # Templ templates
├── sqlite/
│   ├── sqlite.go         # Database connection and migrations
│   ├── shorturl.go       # ShortURLService implementation
│   └── migration/        # SQL migration files
└── Dockerfile            # Multi-stage Docker build
```

## License

MIT
