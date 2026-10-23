# Product

<!-- impeccable:product-schema 1 -->

<!-- Canonical copy. GateCHA-website and GateCHA-WordPress carry copies of this
     file; edit it here first, then propagate. -->

## Platform

web

## Users

- **Primary: self-hosters.** Developers and sysadmins who run their own
  infrastructure and want a CAPTCHA they control instead of a SaaS (ALTCHA
  Sentinel, reCAPTCHA, hCaptcha). They deploy the server with Docker or a
  binary, create one API key per site, wire the widget into their forms and
  call `/api/v1/verify` from their backend. Comfortable with a terminal,
  environment variables and a reverse proxy.
- **Secondary: WordPress site owners.** Less technical; they reach GateCHA
  through the official plugin (Settings → GateCHA: instance URL + API key,
  then tick the forms to protect). They still need a running GateCHA
  instance, usually set up by themselves or by whoever hosts their site.
  The plugin is a distribution channel for the server, not a standalone
  product.

## Product Purpose

GateCHA is a self-hosted, open-source (MIT) management server for the ALTCHA
proof-of-work CAPTCHA protocol: API keys per site, replay protection,
statistics, and a web dashboard, in a single container or binary. It exists
so that CAPTCHA protection does not require sending visitors' data to a third
party or paying for a cloud service. Success: a self-hoster goes from
`docker compose up` to a protected form in minutes, and then forgets it is
running.

## Positioning

Same open ALTCHA protocol and official widget as ALTCHA Sentinel, but free,
self-hosted and fully owned: no cloud, no tracking, no external dependencies
in SQLite mode. GateCHA is not affiliated with ALTCHA, and every surface that
compares them says so.

## Operating Context

Three repositories, one product:

- **GateCHA** (this repo): Go server + embedded Vue 3 admin dashboard
  (login, statistics, API keys list/detail/form, settings). Used by the
  operator occasionally, mostly to create keys and check stats.
- **GateCHA-website** (`gatecha.org`): static landing page, `/docs` and
  `/wordpress` pages. The evaluation surface for both audiences.
- **GateCHA-WordPress**: the `gatecha-captcha` plugin published on
  WordPress.org, with its settings page inside wp-admin and the widget
  rendered inside arbitrary third-party themes. Translated (de_DE, fr_FR).

## Capabilities and Constraints

- ALTCHA-compatible challenge/verify API; works with the official widget.
- API keys (`gk_` prefix) with per-key difficulty, TTL, domain restrictions
  (multiple domains, `*.example.com` wildcards), per-key rate limiting
  (`internal/api/ratelimit.go`) and adaptive difficulty
  (`internal/api/adaptive.go`). The last two are documented on gatecha.org/docs
  but not yet in the README.
- Replay protection on consumed challenges.
- Statistics per key and per day (issued, verified, failed).
- Human Interaction Signature (HIS): optional aggregate interaction signals
  scored 0–1; Monitor mode by default, per-key enforcement opt-in. The
  `his_signals` payload carries aggregates only: never typed content, field
  values, pointer coordinates, timestamps or IP addresses. Like any request,
  `/verify` still sees the source IP, used in memory for rate limiting,
  adaptive difficulty and country stats; no IP is written to the database,
  and request logs carry none unless an operator sets
  `GATECHA_LOG_CLIENT_IP=true` to debug.
- MCP endpoint (off by default) to manage API keys from an AI client.
- SQLite by default; MySQL in the `-mysql` build variant. Images for
  linux/amd64 and linux/arm64; binaries for linux/darwin/windows.
- WordPress plugin (source of truth: GateCHA-WordPress `readme.txt`): 14
  protected form types (core, WooCommerce, CF7, WPForms, Gravity Forms,
  Elementor Pro, Forminator, Formidable, HTML Forms) plus a `[gatecha]`
  shortcode; fail-open / fail-closed modes; bundled widget (no CDN); all
  verification goes through the GateCHA server.
- **Sources of truth:** figures (sizes, requirements) come from the GateCHA
  `README.md`; features come from the README, or from the code where the
  README lags. The website and the plugin readme should align with them,
  never the reverse. Drift found on 2026-10-23 (site claiming a ~3 MB binary,
  ~15 MB image and Go 1.25+, and omitting MySQL, MCP, HIS, rate limiting and
  adaptive difficulty) is corrected in GateCHA-website#8.
- Roadmap, not shipped: dashboard i18n (vue-i18n); content signals on
  `/verify` (#170). Do not present them as available.

## Brand Commitments

- Name is written **GateCHA** (capital G, then "CHA" as in CAPTCHA/ALTCHA).
- Factual, technical, no-hype voice: concrete numbers, commands people can
  copy, privacy claims stated precisely (what does and does not leave the
  browser).
- Always state non-affiliation with ALTCHA when comparing to Sentinel.

## Evidence on Hand

- Real dashboard screenshot: `GateCHA-website/assets/dashboard.png`.
- Social card: `GateCHA-website/assets/social-card.png` (source
  `social-card.html`); favicon `GateCHA-website/assets/favicon.svg`.
- WordPress.org listing assets under `GateCHA-WordPress/.wordpress-org/`.
- Measured image/binary sizes in the GateCHA README; releases on GitHub
  (latest v0.9.1).
- **None:** no testimonials, customer logos, install counts, benchmarks or
  security audits. Do not fabricate them.

## Product Principles

1. **Your server, your data.** Never introduce a third-party call, CDN,
   tracker or cookie on any surface; privacy claims must stay literally true.
2. **Minutes to protected.** The path from install to first verified form is
   the product; every surface shortens it rather than decorating it.
3. **Tell the truth in numbers.** Sizes, requirements and features come from
   the README and releases; stale figures are bugs.
4. **Self-hosters first.** Write for the operator who reads code and configs;
   WordPress owners get a plain-language path without diluting that.
5. **Boring to run.** Defaults that are safe (Monitor mode, MCP off, fail
   modes explicit), so operators can forget it is there.

## Accessibility & Inclusion

The CAPTCHA itself is proof-of-work with no visual puzzle, which is the
product's accessibility argument; any surface that renders the widget must
keep it operable by keyboard and screen readers. The WordPress plugin is
translated (de_DE, fr_FR) and its admin follows wp-admin conventions.
