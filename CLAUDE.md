# GateCHA - Project Guide

## Overview

GateCHA is a self-hosted, open-source (MIT) alternative to ALTCHA Sentinel.
It wraps the ALTCHA proof-of-work CAPTCHA protocol with API key management,
multi-site support, and basic statistics in a single Docker container.

## Tech Stack

- **Backend**: Go (chi router, modernc.org/sqlite, altcha-lib-go, golang-jwt)
- **Frontend**: Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS v4 + Chart.js
- **Database**: SQLite (WAL mode, embedded)
- **Deployment**: Docker multi-stage (Alpine)

## Project Structure

```
cmd/gatecha/main.go           - Entry point
internal/config/               - Environment variable parsing
internal/database/             - SQLite connection + migrations
internal/models/               - Data models (apikey, challenge, stats)
internal/altcha/               - ALTCHA protocol wrapper
internal/api/                  - HTTP handlers + middleware + router
internal/auth/                 - JWT + bcrypt admin auth
internal/dashboard/            - go:embed for Vue SPA
web/                           - Vue.js source
```

## Build Commands

```bash
make frontend     # Build Vue SPA + copy to internal/dashboard/dist
make backend      # Build Go binary
make build        # Both
make dev          # Run Go backend in dev mode
```

## Key Conventions

- Go module: `github.com/Upellift99/GateCHA`
- API key prefix: `gk_` (e.g., `gk_a1b2c3d4e5f67890abcdef12`)
- Public API: `/api/v1/challenge`, `/api/v1/verify` (API key auth via `?apiKey=`)
- Admin API: `/api/admin/*` (JWT auth via `Authorization: Bearer`)
- All dates stored as UTC ISO8601/RFC3339
- Stats use SQLite UPSERT for atomic counter increments

## Future Roadmap

- i18n / multilanguage support (vue-i18n)
- Rate limiting per API key
- Adaptive difficulty
