# GateCHA

[![Go](https://github.com/Upellift99/GateCHA/actions/workflows/test.yml/badge.svg)](https://github.com/Upellift99/GateCHA/actions/workflows/test.yml)
[![Docker](https://github.com/Upellift99/GateCHA/actions/workflows/docker.yml/badge.svg)](https://github.com/Upellift99/GateCHA/actions/workflows/docker.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Upellift99/GateCHA)](https://goreportcard.com/report/github.com/Upellift99/GateCHA)

**Self-hosted ALTCHA CAPTCHA management with API keys, multi-site support, and statistics.**

GateCHA is an open-source alternative to [ALTCHA Sentinel](https://altcha.org/docs/v2/sentinel/). It wraps the [ALTCHA](https://altcha.org/) proof-of-work CAPTCHA protocol with a management layer: API key management, per-site configuration, replay protection, and a dashboard with statistics.

## Features

- **ALTCHA-compatible** - Works with the official [ALTCHA widget](https://altcha.org/docs/v2/widget-integration/) (MIT)
- **API Key Management** - Create keys per site with custom difficulty, TTL, and domain restrictions
- **Replay Protection** - Consumed challenges are tracked and rejected on reuse
- **Statistics Dashboard** - Track challenges issued, verifications (success/fail), per key, per day
- **Single Binary** - Vue.js dashboard embedded in the Go binary via `go:embed`
- **Docker Ready** - One container, SQLite embedded, zero external dependencies
- **Lightweight** - ~15MB Docker image, ~3MB binary

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
  ghcr.io/upellift99/gatecha:main
```

### From Source

```bash
# Prerequisites: Go 1.26+, Node.js 20+
git clone https://github.com/Upellift99/GateCHA.git
cd GateCHA
make build
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
    challengeurl="https://your-gatecha-host/api/v1/challenge?apiKey=gk_your_key_id"
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

## API Endpoints

### Public (API Key auth via `?apiKey=gk_xxx`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/challenge` | Generate a PoW challenge |
| `POST` | `/api/v1/verify` | Verify a solution |

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

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GATECHA_LISTEN_ADDR` | `:8080` | Listen address |
| `GATECHA_DB_PATH` | `./data/gatecha.db` | SQLite database path |
| `GATECHA_SECRET_KEY` | *(auto-generated)* | JWT signing secret |
| `GATECHA_ADMIN_USERNAME` | `admin` | Admin username |
| `GATECHA_ADMIN_PASSWORD` | *(auto-generated)* | Admin password |
| `GATECHA_LOG_LEVEL` | `info` | Log level |
| `GATECHA_CLEANUP_INTERVAL` | `10` | Cleanup interval (minutes) |
| `GATECHA_CORS_ALLOW_ALL` | `false` | Allow CORS from any origin |

## License

MIT - see [LICENSE](LICENSE).
