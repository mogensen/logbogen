# Logbogen

A logbook web application for tracking activities, certifications, and achievements. Built with [Go Fiber](https://gofiber.io/), SQLite, and server-rendered HTML templates.

## Build

Requires Go 1.24+.

```bash
go build -o logbogen .
./logbogen
```

The server starts at `http://127.0.0.1:3000` by default.

Copy `.env.example` to `.env` and adjust settings before running.

## Make targets

`make help` lists them: `build`, `test`, `test-e2e`, `run` and `lint`. `make lint` fails on unformatted files,
then runs `go vet` and `staticcheck`, which is installed into `.bin/` on first use. CI runs it too.

## Logging

The app logs structured lines, one JSON object per line by default (`LOG_FORMAT=text` gives key=value lines for
local use, and `LOG_LEVEL` sets the lowest level). Every request gets an id, sent back in the `X-Request-Id`
header and shown on error pages as the Fejl-id, and it is on every line logged for that request:

* `request`: one line per request with `method`, `path`, `route`, `status`, `latency_ms`, `bytes`, `ip` and
  `user_id`. The query string is left out, since it can hold OAuth codes.
* Events for what users do, such as `user_logged_in`, `user_logged_out`, `activity_created`, `activity_updated`,
  `activity_deleted`, `activity_cloned`, `certification_created`, `theme_changed`, with the ids involved.
* `request failed` and `panic recovered` for server errors, with the cause and the stack, which the page never shows.
* Database queries at debug level, and `slow query` and `query failed` above that.

## Browser tests

`make test-e2e` runs [Playwright](https://github.com/playwright-community/playwright-go) tests in a real Chrome
against the app, which catch what the Go tests cannot: content security policy violations, script errors and
the JavaScript behavior (logout, the category filter, the scoreboard and the theme toggle). They need Google
Chrome and network access, since the pages load their scripts, fonts, map tiles and avatars from public hosts.
The first run downloads the Playwright driver, and the tests are left out of `make test`.

## Docker

```bash
docker build -t logbogen .
docker run -p 3000:3000 -v logbogen-data:/app/data --env-file .env logbogen
```

The image listens on `0.0.0.0:3000` (`LISTEN_ADDR`), runs as a non-root user and keeps the
database and sessions in `/app/data`, which is the directory to mount as a volume. It runs
against real Auth0, so set the `AUTH0_*` variables, and put a TLS reverse proxy in front of it.

Pushing a git tag publishes the image to `ghcr.io/<owner>/<repo>` with the tag name and `latest`.

## Data

The SQLite database (including sessions) is stored in `./data/logbog.db`, or the path in `DB`.
The directory is created on startup, and is the one to mount as a volume when running in a container.

## Authentication

Logbogen authenticates users through **Auth0** (OIDC Authorization Code flow). The
app no longer stores passwords — users are identified by their Auth0 subject (`sub`)
and a user row is created on first login.

### Local development (no Auth0 needed)

Set `AUTH_DEV_MODE=true` in `.env` (the default in `.env.example`). The login page
then shows a small dev-login form: enter any name/email and you're signed in. No Auth0
tenant or network access is required, and the test suite runs in this mode too.

Dev mode refuses to start on a non loopback address, since it lets anyone sign in as any
user. Set `ALLOW_DEV_MODE=true` to override this on a trusted network.

### Running against real Auth0

1. Create a **Regular Web Application** in the Auth0 dashboard.
2. Register `http://localhost:3000/auth/callback` as an **Allowed Callback URL** and
   `http://localhost:3000/` as an **Allowed Logout URL** (`localhost` is exempt from
   Auth0's HTTPS requirement).
3. In `.env`, set `AUTH_DEV_MODE=false` and fill in `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`,
   `AUTH0_CLIENT_SECRET`, `AUTH0_CALLBACK_URL`, and `AUTH0_LOGOUT_RETURN_URL`.

> **Upgrading an existing database:** the old `users` table had a `NOT NULL password`
> column. The app drops it automatically on startup, so existing databases keep working
> (and keep their activities/certifications). Note that pre-Auth0 accounts are not linked
> to Auth0 identities — a fresh user row is created on first Auth0 login.

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
