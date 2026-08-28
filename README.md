# GateCHA

[![Go](https://github.com/Upellift99/GateCHA/actions/workflows/test.yml/badge.svg)](https://github.com/Upellift99/GateCHA/actions/workflows/test.yml)
[![Docker](https://github.com/Upellift99/GateCHA/actions/workflows/docker.yml/badge.svg)](https://github.com/Upellift99/GateCHA/actions/workflows/docker.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Upellift99/GateCHA)](https://goreportcard.com/report/github.com/Upellift99/GateCHA)

**Self-hosted ALTCHA CAPTCHA management with API keys, multi-site support, and statistics.**

🌐 **[Website](https://gatecha.org)** · 📖 **[Documentation](https://gatecha.org/docs)** · 🧩 **[WordPress plugin](https://github.com/Upellift99/GateCHA-WordPress)**

GateCHA is an open-source alternative to [ALTCHA Sentinel](https://altcha.org/docs/v2/sentinel/). It wraps the [ALTCHA](https://altcha.org/) proof-of-work CAPTCHA protocol with a management layer: API key management, per-site configuration, replay protection, and a dashboard with statistics.

## Features

- **ALTCHA-compatible** - Works with the official [ALTCHA widget](https://altcha.org/docs/v2/widget-integration/) (MIT)
- **API Key Management** - Create keys per site with custom difficulty, TTL, and domain restrictions (multiple domains + `*.example.com` wildcards)
- **Replay Protection** - Consumed challenges are tracked and rejected on reuse
- **Statistics Dashboard** - Track challenges issued, verifications (success/fail), per key, per day
- **MCP Endpoint** - Manage API keys from an AI client (Cursor, Claude Code) over an authenticated MCP server; off by default
- **Single Binary** - Vue.js dashboard embedded in the Go binary via `go:embed`
- **Multi-Database** - SQLite by default; MySQL available in the `mysql` build variant
- **Docker Ready** - One container, zero external dependencies (SQLite mode)
- **Lightweight** - ~23.8MB Docker image, ~14.2MB binary (SQLite); ~24.2MB / ~14.5MB with MySQL support

## Quick Start

### Docker Compose (recommended)

```bash
mkdir -p /opt/docker/GateCHA && cd /opt/docker/GateCHA
wget https://raw.githubusercontent.com/Upellift99/GateCHA/refs/heads/main/docker-compose.yml
docker compose up -d
```

Open `http://localhost:8080` and log in with `admin` / `changeme`.

### Docker Run

```bash
docker run -d -p 8080:8080 \
  -v gatecha_data:/app/data \
  -e GATECHA_ADMIN_PASSWORD=your-password \
  ghcr.io/upellift99/gatecha:latest
```

### Image tags

| Tag | Points to | Use it when |
|---|---|---|
| `latest` | Newest published release | Default choice |
| `0.3.3`, `0.3` | That exact release / newest patch in the 0.3 line | You want reproducible upgrades |
| `main` | Latest commit on `main`, unreleased | Testing an unreleased fix |
| `sha-<commit>` | One specific build | Pinning or bisecting |

Pin `0.3` (or a full version) in production if you'd rather approve minor
upgrades yourself, since `latest` crosses minor versions as they ship.

### Prebuilt binary (no Docker)

GateCHA ships as a single self-contained binary with the web dashboard embedded,
so there are no extra files or runtime dependencies. Grab the archive for your
platform from the [latest release](https://github.com/Upellift99/GateCHA/releases/latest)
(`linux`/`darwin`/`windows`, `amd64`/`arm64`):

```bash
# Example: Linux x86_64. Check the releases page for the current version
VERSION=0.3.0
curl -fsSL -o gatecha.tar.gz \
  "https://github.com/Upellift99/GateCHA/releases/download/v${VERSION}/gatecha_${VERSION}_linux_amd64.tar.gz"
tar -xzf gatecha.tar.gz
GATECHA_ADMIN_PASSWORD=your-password ./gatecha
```

Open `http://localhost:8080`. See [Configuration](#configuration) for all options.

### From Source

```bash
# Prerequisites: Go 1.26+, Node.js 20+
git clone https://github.com/Upellift99/GateCHA.git
cd GateCHA
make build        # SQLite only (default)
make build-mysql  # with MySQL support
./gatecha
```

## Usage

### 1. Create an API Key

Log in to the dashboard at `http://localhost:8080`, go to **API Keys**, and create a new key.

### 2. Add the Widget to Your Site

```html
<script async defer src="https://cdn.jsdelivr.net/npm/altcha/dist/altcha.min.js" type="module"></script>

<form action="/your-endpoint" method="POST">
  <!-- your form fields -->
  <altcha-widget
    challenge="https://your-gatecha-host/api/v1/challenge?apiKey=gk_your_key_id"
  ></altcha-widget>
  <button type="submit">Submit</button>
</form>
```

### 3. Verify on Your Backend

```python
# Example: Python
import requests

altcha_payload = request.form.get('altcha')
resp = requests.post(
    'https://your-gatecha-host/api/v1/verify?apiKey=gk_your_key_id',
    json={'payload': altcha_payload}
)
if resp.json().get('ok'):
    # Valid submission
    pass
```

### 4. Optional: collect interaction signals (HIS)

Alongside the proof of work, GateCHA can score *how* a visitor interacted with the
page: pointer travel, scroll and touch counts, typing rhythm and timings. This is
the Human Interaction Signature (HIS). It records aggregates only, never
coordinates, timestamps, key contents or field values.

Load the collector your own instance serves:

```html
<script src="https://your-gatecha-host/api/public/his.js" defer></script>
```

It attaches to any form containing an ALTCHA widget and, on submit, fills a hidden
`gatecha_his_signals` field with a JSON object. Forward that value to `/verify`:

```python
import json

resp = requests.post(
    'https://your-gatecha-host/api/v1/verify?apiKey=gk_your_key_id',
    json={
        'payload': request.form.get('altcha'),
        'his_signals': json.loads(request.form.get('gatecha_his_signals') or 'null'),
    },
)
```

Calling `/verify` from the browser instead? Read the same object from
`window.gatechaHIS.signals()`.

HIS never blocks. Scores are recorded and surfaced on the dashboard and per key.
Enabling **HIS sampling** on a key additionally stores the raw aggregates so the
key detail page can show you the score distribution.

To build your own collector, `his_signals` is this object, all numeric:

| Field | Meaning |
|-------|---------|
| `duration_ms` | Length of the observed interaction window |
| `time_to_first_ms` | Delay until the first interaction event, `-1` if none |
| `pointer_events` | Sampled pointer/mouse move events |
| `pointer_distance` | Total pointer path length, CSS pixels |
| `scrolls` | Scroll events |
| `touches` | Touch events |
| `keydowns` | Key-down events |
| `key_interval_stdev_ms` | Standard deviation of inter-keydown intervals |

#### Reading the score back

When a request carries `his_signals`, `/verify` returns the Monitor score
alongside the verification outcome, so your backend can apply its own threshold
without waiting for server-side enforcement:

```json
{ "ok": true, "his_bot_score": 0.7, "his_bot_suspected": false }
```

**`his_bot_score` runs from 0 to 1 where higher means more bot-like.** This is the
reverse of reCAPTCHA's convention, where the score measures confidence that the
visitor is human. Copying a reCAPTCHA rule across ("reject below 0.6") would
reject your humans and pass the bots.

`his_bot_suspected` is that score judged against GateCHA's own threshold
(`>= 0.8`), the same rule the dashboard counters use, for when you would rather
not own a number.

Two things to know before picking a threshold:

- **Both fields are absent when the request carried no `his_signals`.** Absent is
  not 0. A visitor whose collector never ran is not thereby a human, so treat the
  missing fields as "no opinion" rather than as a clean score.
- **The score moves in steps of 0.1**, being a handful of additive penalties and
  credits, so 0.6 against 0.8 is a real choice while 0.65 against 0.7 is not.
  Read it as a few tiers, not as a probability. A submission with no motion at
  all sits at 0.7 on its own, deliberately below the suspect threshold, because
  keyboard-only and assistive-technology users look like that too.

The fields ride along with failed verifications as well, which are often the ones
worth inspecting.

## API Endpoints

### Public (API Key auth via `?apiKey=gk_xxx`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/challenge` | Generate a PoW challenge |
| `POST` | `/api/v1/verify` | Verify a solution |

### Public (no auth)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/public/his.js` | Client-side HIS collector, see [step 4](#4-optional-collect-interaction-signals-his) |

### Admin (JWT auth via `Authorization: Bearer`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/admin/login` | Authenticate |
| `GET` | `/api/admin/keys` | List API keys |
| `POST` | `/api/admin/keys` | Create API key |
| `GET/PUT/DELETE` | `/api/admin/keys/:id` | Manage API key |
| `POST` | `/api/admin/keys/:id/rotate-secret` | Rotate HMAC secret |
| `GET` | `/api/admin/stats/overview` | Global statistics |
| `GET` | `/api/admin/stats/keys/:id` | Per-key statistics |
| `GET` | `/healthz` | Health check |

### MCP (MCP token auth via `Authorization: Bearer gm_xxx`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/mcp` | MCP server for API key management. Off by default, answers `404` until enabled. See [MCP Endpoint](#mcp-endpoint) |

## MCP Endpoint

GateCHA can expose its API key management as an [MCP](https://modelcontextprotocol.io/)
server, so an AI client such as Cursor or Claude Code can list, create and update keys
without you opening the dashboard.

**The endpoint is off by default and has to be turned on deliberately.** It is a second
authentication path to full admin capability, one that bypasses the dashboard login. While
it is off, `/mcp` answers `404` and no token is accepted.

### 1. Turn it on and issue a token

In the dashboard, go to **Settings** and find the **MCP Access** panel:

1. Switch **MCP endpoint** on.
2. Create a token, naming it after the person or machine that will use it. Issue one token
   per person: they are revoked individually, so revoking one does not disturb anyone else.
3. Tick **Read only** for a token that must never change anything. A read-only token is
   given a server on which the write tools were never registered, so they are neither
   listed nor callable.
4. Copy the secret. It starts with `gm_`, it is shown **once**, and it cannot be retrieved
   afterwards. Only its first characters are kept, so the list can tell tokens apart.

The token list shows when each token was last used, which is what tells you a token is
dormant and can be revoked.

### 2. Point your client at it

The endpoint speaks streamable HTTP at `/mcp` and authenticates with a bearer token. Any
client that can send an `Authorization` header works, and no OAuth setup is involved.

**Cursor** (`~/.cursor/mcp.json`, or `.cursor/mcp.json` in a project):

```json
{
  "mcpServers": {
    "gatecha": {
      "url": "https://captcha.example.com/mcp",
      "headers": {
        "Authorization": "Bearer ${env:GATECHA_MCP_TOKEN}"
      }
    }
  }
}
```

**Claude Code** (`.mcp.json`):

```json
{
  "mcpServers": {
    "gatecha": {
      "type": "http",
      "url": "https://captcha.example.com/mcp",
      "headers": {
        "Authorization": "Bearer ${GATECHA_MCP_TOKEN}"
      }
    }
  }
}
```

Keep the token in an environment variable rather than in the file itself. The token is
**only** read from the `Authorization` header: it is deliberately not accepted in the query
string, because URLs end up in proxy logs and browser history.

### 3. Available tools

| Tool | Read-only token | Description |
|------|-----------------|-------------|
| `list_keys` | yes | List keys, with an optional case-insensitive search on name, domain and key ID |
| `get_key` | yes | Fetch one key by ID |
| `create_key` | no | Create a key. The HMAC secret is returned here and nowhere else |
| `update_key` | no | Change a key's settings. Omitted fields keep their current value |
| `enable_key` | no | Enable a key so it serves challenges again |
| `disable_key` | no | Disable a key. The site using it stops being able to issue or verify challenges |

Three deliberate limits are worth knowing about:

- **The HMAC secret is returned by `create_key` only.** The type the other tools return has
  no field for it, so no tool can leak it by forgetting to strip it.
- **Enabling and disabling are their own tools**, not a field on `update_key`. Taking a
  site's CAPTCHA down shows up under its own name in the consent prompt your client
  displays, instead of hiding inside a generic update, and `update_key` cannot do it at all.
- **There is no delete tool.** Removing a key breaks a live site with no undo, so it stays
  a dashboard action.

## MySQL Support

The default build is SQLite-only for a lightweight single-binary deployment. MySQL support is compiled in only when explicitly requested.

**Build locally with MySQL support:**
```bash
make build-mysql
```

**Docker image with MySQL support:**
```bash
docker build --build-arg BUILD_TAGS=mysql -t gatecha:mysql .
```

**Docker Compose with MySQL:**
```bash
docker compose -f docker-compose.mysql.yml up -d
```

> **Note for contributors:** When updating Go dependencies while working on MySQL support, run `go mod tidy -tags mysql` instead of plain `go mod tidy` to preserve the MySQL driver in `go.mod`.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GATECHA_LISTEN_ADDR` | `:8080` | Listen address |
| `GATECHA_DB_DRIVER` | `sqlite` | Database driver: `sqlite` always available; `mysql` requires the mysql build variant |
| `GATECHA_DB_DSN` | `./data/gatecha.db` | Database DSN: file path for SQLite, connection string for MySQL (e.g. `user:pass@tcp(host:3306)/dbname?parseTime=true`) |
| `GATECHA_SECRET_KEY` | *(auto-generated)* | JWT signing secret |
| `GATECHA_ADMIN_USERNAME` | `admin` | Admin username |
| `GATECHA_ADMIN_PASSWORD` | *(auto-generated)* | Admin password |
| `GATECHA_LOG_LEVEL` | `info` | Log level |
| `GATECHA_CLEANUP_INTERVAL` | `10` | Cleanup interval (minutes) |
| `GATECHA_HIS_SAMPLE_RETENTION_DAYS` | `30` | Retention (days) for opted-in raw HIS calibration samples |
| `GATECHA_CORS_ALLOW_ALL` | `false` | Allow CORS from any origin |
| `GATECHA_TRUST_PROXY` | `false` | Trust `X-Forwarded-For`/`X-Real-IP` for the client IP. **Set to `true` when behind a reverse proxy** (see note below) |
| `GATECHA_ENABLE_HSTS` | `false` | Send the `Strict-Transport-Security` header (enable only when always served over HTTPS) |
| `GATECHA_MAX_BODY_BYTES` | `1048576` | Maximum accepted request body size, in bytes |
| `GATECHA_RATE_LIMIT_ENABLED` | `true` | Enable per-IP rate limiting |
| `GATECHA_RATE_LIMIT_LOGIN` | `5` | Admin login requests per minute, per IP |
| `GATECHA_RATE_LIMIT_API` | `60` | Public API (`/api/v1/*`) requests per minute, per IP |

> ⚠️ **Behind a reverse proxy, set `GATECHA_TRUST_PROXY=true`.**
> Per-IP rate limiting keys off the connecting IP. With `GATECHA_TRUST_PROXY=false`
> behind a proxy, that IP is the **proxy itself**, so every visitor shares a single
> rate-limit bucket. It exhausts almost immediately, the public ALTCHA challenge
> endpoint starts returning `429`s, and the **login captcha breaks** with
> *"Expected application/json, received text/html"*. Enabling `TRUST_PROXY` makes
> the limiter use each visitor's real IP. Only enable it behind a **trusted** proxy
> that sets `X-Forwarded-For`/`X-Real-IP`, otherwise clients can spoof their IP.

## License

MIT - see [LICENSE](LICENSE).
