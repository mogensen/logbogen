# Logbogen

A logbook web application for tracking activities, certifications, and achievements. Built with [Go Fiber](https://gofiber.io/), SQLite, and server-rendered HTML templates.

## Build

Requires Go 1.27+.

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
  `user_id`. The query string is left out, since it can hold tokens.
* Events for what users do, such as `user_logged_in`, `user_logged_out`, `activity_created`, `activity_updated`,
  `activity_deleted`, `activity_cloned`, `certification_created`, `theme_changed`, with the ids involved.
* `request failed` and `panic recovered` for server errors, with the cause and the stack, which the page never shows.
* Database queries at debug level, and `slow query` and `query failed` above that.

## Browser tests

`make test-e2e` runs [Playwright](https://github.com/playwright-community/playwright-go) tests in a real Chrome
against the app, which catch what the Go tests cannot: content security policy violations, script errors and
the JavaScript behavior (Firebase sign-in, logout, the category filter, the scoreboard and the theme toggle).
They need Google Chrome and network access, since the pages load their scripts, fonts, map tiles and avatars
from public hosts. The sign-in tests run against the Firebase Auth emulator, which they start with `npx`
(Node is needed) unless one already runs on `127.0.0.1:9099`. The first run downloads the Playwright driver
and the emulator, and the tests are left out of `make test`.

## Docker

```bash
docker build -t logbogen .
docker run -p 3000:3000 -v logbogen-data:/app/data --env-file .env logbogen
```

The image listens on `0.0.0.0:3000` (`LISTEN_ADDR`), runs as a non-root user and keeps the
database and sessions in `/app/data`, which is the directory to mount as a volume. It runs
against real Firebase Authentication, so set the `FIREBASE_*` variables, and put a TLS reverse proxy in front of it.

On a host with a very old Docker the container can abort at start with `runtime/cgo: pthread_create failed:
Operation not permitted`, since its seccomp profile blocks the `clone3` syscall. Upgrading Docker fixes it, and
`--security-opt seccomp=unconfined` (`security_opt: [seccomp=unconfined]` in compose) works around it.

Pushing a git tag publishes the image to `ghcr.io/<owner>/<repo>` with the tag name and `latest`.

## Data

The SQLite database (including sessions) is stored in `./data/logbog.db`, or the path in `DB`.
The directory is created on startup, and is the one to mount as a volume when running in a container.

## Authentication

Logbogen authenticates users with **Firebase Authentication**, with Google or email and
password. The browser signs in with the Firebase JS SDK and posts the ID token to
`/auth/session`, where the server verifies it against the project's public signing keys
and starts its own session. Email addresses must be verified. The app does not store
passwords — users are identified by their Firebase user ID and a user row is created on
first login.

### Local development (no Firebase needed)

Set `AUTH_DEV_MODE=true` in `.env` (the default in `.env.example`). The login page
then shows a small dev-login form: enter any name/email and you're signed in. No Firebase
project or network access is required, and the test suite runs in this mode too.

Dev mode refuses to start on a non loopback address, since it lets anyone sign in as any
user. Set `ALLOW_DEV_MODE=true` to override this on a trusted network.

### Running against real Firebase

1. In the Firebase console, enable the **Google** and **Email/Password** sign-in
   providers, and add your domains (`localhost` and the production host) under
   **Authentication > Settings > Authorized domains**.
2. Register a web app and copy its settings. They are public identifiers, not secrets.
3. In `.env`, set `AUTH_DEV_MODE=false` and fill in `FIREBASE_PROJECT_ID`,
   `FIREBASE_API_KEY` and `FIREBASE_AUTH_DOMAIN` (for example `my-project.firebaseapp.com`).

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
