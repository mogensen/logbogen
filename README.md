# Logbogen

A sailing logbook web application for tracking activities, certifications, and achievements. Built with [Go Fiber](https://gofiber.io/), SQLite, and server-rendered HTML templates.

## Build

Requires Go 1.24+.

```bash
go build -o logbogen .
./logbogen
```

The server starts at `http://127.0.0.1:3000` by default.

Copy `.env.example` to `.env` (if present) and adjust settings before running.

## Live reload with Air

[Air](https://github.com/air-verse/air) watches for file changes and rebuilds automatically during development.

Install Air:

```bash
go install github.com/air-verse/air@latest
```

Run:

```bash
air
```

Air will rebuild and restart the server whenever a `.go` file changes.

## Testing

```bash
go test ./...
```

Tests spin up a real in-memory SQLite database and a local HTTP server — no mocks or external dependencies needed.
