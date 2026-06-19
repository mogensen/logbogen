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

## Authentication

Logbogen authenticates users through **Auth0** (OIDC Authorization Code flow). The
app no longer stores passwords — users are identified by their Auth0 subject (`sub`)
and a user row is created on first login.

### Local development (no Auth0 needed)

Set `AUTH_DEV_MODE=true` in `.env` (the default in `.env.example`). The login page
then shows a small dev-login form: enter any name/email and you're signed in. No Auth0
tenant or network access is required, and the test suite runs in this mode too.

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
