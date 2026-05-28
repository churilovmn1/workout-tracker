# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build        # compile to bin/workout-tracker
make run          # go run ./cmd/
make test         # go test ./... -v
make lint         # golangci-lint run ./...
make migrate-up   # apply all pending migrations
make migrate-down # roll back last migration (one step)
```

Run a single test package:
```bash
go test ./internal/service/... -v -run TestName
```

Local dev setup:
```bash
cp .env.example .env  # set DATABASE_URL, JWT_SECRET, TELEGRAM_BOT_TOKEN
docker compose up -d db
make migrate-up
make run
```

Full stack via Docker:
```bash
docker compose up -d        # starts db + app
docker compose run --rm migrate  # apply migrations
```

## Architecture

The app is a workout-tracking service with three entry points sharing the same service layer:

```
cmd/main.go          — wires everything together, starts HTTP server + Telegram bot
config/              — reads env vars (DATABASE_URL required; PORT, JWT_SECRET, TELEGRAM_BOT_TOKEN optional)
internal/
  models/            — plain structs (User, Workout, WorkoutExercise, Exercise, Template, TemplateExercise)
  repository/        — pgx/v5 pool-based SQL; one file per entity; WorkoutRepository.Create uses a tx for workout+exercises atomically
  service/           — business logic; one file per domain; AuthService owns JWT (HS256, 24h expiry) and bcrypt hashing
  handler/           — chi router; one file per domain; AuthMiddleware + AdminOnly middleware; getUserID() pulls user_id from context
bot/                 — Telegram bot (long-polling); session state machine for multi-step /newworkout flow; auto-registers users on /start
migrations/          — numbered up/down SQL files; applied with golang-migrate CLI
web/                 — HTML templates + static assets served by WebHandler
```

### Key data flow

HTTP request → chi router → `AuthMiddleware` (validates Bearer JWT, injects `user_id`+`role` into context) → handler → service → repository → PostgreSQL.

Telegram bot bypasses HTTP: it calls service and repository directly. Users created via `/start` get a generated login (`tg_<id>`) and a placeholder password hash (`-`); they are linked by `telegram_id` column added in migration 007.

### Role model

Two roles: `user` (default) and `admin`. Exercise create/update/delete are admin-only (`AdminOnly` middleware). All other `/api` routes require only a valid JWT.

### Dependency injection

All dependencies flow top-down in `cmd/main.go`. No global state or DI framework — each layer receives only what it needs as constructor arguments.

## Environment variables

| Variable | Required | Default |
|---|---|---|
| `DATABASE_URL` | yes | — |
| `JWT_SECRET` | no | `default-secret-change-me` |
| `PORT` | no | `8080` |
| `TELEGRAM_BOT_TOKEN` | no | — (bot disabled if empty) |
