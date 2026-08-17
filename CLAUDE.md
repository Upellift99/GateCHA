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

## Known Constraints

- `web/package.json` forces `overrides: { "js-beautify": "^2.0.3" }`. `@vue/test-utils`
  pins `js-beautify: ^1.14.9`, whose `editorconfig`/`minimatch@9` subtree carries
  `brace-expansion@2.x` and an unfixable OOM advisory (GHSA-mh99-v99m-4gvg). Override the
  **parent**, not `brace-expansion` itself: 5.x exports an object rather than a callable,
  so `minimatch@9`'s `const expand = require('brace-expansion')` would throw on any brace
  pattern. Nothing in the test suite reaches that path, so it would fail in production use
  rather than in CI — which is why the parent override is the safe form. Drop it once
  `@vue/test-utils` widens its range; check `npm view @vue/test-utils@latest dependencies`.
  See #107.
- Release binaries and the Docker image must build on the same Go minor. Workflows use a
  bare `go-version: "1.26"` so setup-go resolves the latest patch; never switch back to
  `go-version-file: go.mod`, whose pinned `go 1.25.0` shipped binaries with 20 reachable
  stdlib vulnerabilities. The `govulncheck` CI job guards this. See #108.
- The bare `go-version` above is necessary but **not sufficient**: setup-go looks in the
  runner's tool cache before the version manifest, and any cached `1.26.x` satisfies the
  spec. Every `setup-go` step therefore also sets `check-latest: true` — without it CI
  stayed on go1.26.5 for a week after 1.26.6 fixed three reachable stdlib vulns. Add it to
  any new `setup-go` step. The Dockerfile is unaffected: `golang:1.26-alpine` floats.

## Future Roadmap

- i18n / multilanguage support (vue-i18n)
- Rate limiting per API key
- Adaptive difficulty
