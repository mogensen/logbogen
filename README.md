# Logbogen

A sailing logbook web application for tracking activities, certifications, and achievements. Built with [Go Fiber](https://gofiber.io/), SQLite, and server-rendered HTML templates.

## Build

Requires Go 1.24+.

```bash
go build -o logbogen .
./logbogen
```

The server starts at `http://127.0.0.1:3000` by default.

Copy `.env.example` to `.env` and adjust settings before running.

## Make targets

`make help` lists them: `build`, `test`, `run` and `lint`. `make lint` fails on unformatted files,
then runs `go vet` and `staticcheck`, which is installed into `.bin/` on first use. CI runs it too.

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
An existing `fiber.db` is not migrated automatically: stop the app and move it with
`mkdir -p data && mv fiber.db data/logbog.db`.

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
