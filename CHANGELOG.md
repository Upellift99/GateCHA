# Changelog

## [0.3.5](https://github.com/Upellift99/GateCHA/compare/v0.3.4...v0.3.5) (2026-08-17)

**Upgrading is recommended.** Unlike 0.3.4, whose advisories never reached a
running instance, this one rebuilds on a Go standard library that fixes three
vulnerabilities reachable from the request-handling path: a post-handshake
message flood in `crypto/tls` (GO-2026-6090), a missing `ReadHeaderTimeout` on
the unencrypted HTTP/2 check in `net/http` (GO-2026-6089), and unbounded
recursion in `encoding/asn1` (GO-2026-5972). All three are denial of service —
no data disclosure, no remote code execution — but two sit directly on the
listener GateCHA exposes, so an instance reachable from the internet can be made
to stop answering challenges. Both the 0.3.4 binaries and the 0.3.4 container
image are affected: both were built on go1.26.5, and the fixes landed in
go1.26.6.

This finishes what 0.3.2 (#108) started. That release stopped pinning the Go
toolchain from `go.mod` so CI would follow the newest patch, and the
`govulncheck` job was added to prove it. The pin was gone, but the follow-up
never happened: `setup-go` consults the runner's tool cache before the version
manifest, and the cached go1.26.5 satisfied a bare `1.26`, so every job — the
scanner included — kept building against a stdlib a week out of date while
reporting success. Setting `check-latest: true` is what forces the lookup
(#128). Operators need do nothing here beyond upgrading; the change is entirely
in CI.

Also included: `golang.org/x/crypto` 0.55.0 (#125), which ships no behaviour
change for GateCHA, a `pinia` patch in the dashboard (#126), and dev-only
tooling bumps that never enter the image.

### Bug Fixes

* **ci:** make setup-go actually install the newest Go patch ([#128](https://github.com/Upellift99/GateCHA/issues/128)) ([8c17d37](https://github.com/Upellift99/GateCHA/commit/8c17d37f072f7a7bfc8aa99507cc8ab12ff0183d))

## [0.3.4](https://github.com/Upellift99/GateCHA/compare/v0.3.3...v0.3.4) (2026-08-10)

This is the first release where `ghcr.io/upellift99/gatecha:latest` means the
newest release. Until now it tracked `main`, so anyone following the README or
the bundled `docker-compose.yml` was running unreleased code without being told.
Pulling `:latest` after this release moves you from a `main` build onto 0.3.4.
If you deliberately wanted the development build, pin `:main` — it still exists.
To approve minor upgrades by hand instead, pin `:0.3`.

Also refreshes the embedded IP geolocation database (`phuslu/iploc`, #119),
which improves country attribution in the traffic-by-country panel. It ships no
code change, and would otherwise have sat unreleased on `main` indefinitely,
since dependency bumps do not trigger a release on their own.

Two npm advisories were closed along the way (`brace-expansion`, `nanoid`, #122).
Neither reaches a running instance: both are build-tooling packages, and the
container ships only the Go binary with the dashboard embedded. **This is not a
security fix for operators and does not warrant an out-of-band upgrade.**

### Bug Fixes

* **ci:** point the :latest image at the newest release, not main ([#123](https://github.com/Upellift99/GateCHA/issues/123)) ([3933de8](https://github.com/Upellift99/GateCHA/commit/3933de80265af6583240e1d425a31ee4d522a7f8))

## [0.3.3](https://github.com/Upellift99/GateCHA/compare/v0.3.2...v0.3.3) (2026-08-03)


### Bug Fixes

* **dashboard:** show the build version in the footer after logging in ([#114](https://github.com/Upellift99/GateCHA/issues/114)) ([4c3ba67](https://github.com/Upellift99/GateCHA/commit/4c3ba67983ed27da076315f7c88838c526dcf87f))

## [0.3.2](https://github.com/Upellift99/GateCHA/compare/v0.3.1...v0.3.2) (2026-07-27)


### Bug Fixes

* **security:** build releases on a patched Go toolchain, move off EOL Alpine ([#108](https://github.com/Upellift99/GateCHA/issues/108)) ([ed85dd2](https://github.com/Upellift99/GateCHA/commit/ed85dd2e807742b878831f43e853fb473280f55c))

## [0.3.1](https://github.com/Upellift99/GateCHA/compare/v0.3.0...v0.3.1) (2026-06-21)


### Bug Fixes

* **ui:** footer links + stack "Traffic by country" header ([#65](https://github.com/Upellift99/GateCHA/issues/65)) ([59bde01](https://github.com/Upellift99/GateCHA/commit/59bde018c597450c60650d0abe4f2d087594161b))

## [0.3.0](https://github.com/Upellift99/GateCHA/compare/v0.2.1...v0.3.0) (2026-06-21)


### Features

* **ui:** link the footer version to its GitHub release ([#63](https://github.com/Upellift99/GateCHA/issues/63)) ([d296743](https://github.com/Upellift99/GateCHA/commit/d2967434060ba8d296e8e149dffae2728fbe2054))

## [0.2.1](https://github.com/Upellift99/GateCHA/compare/v0.2.0...v0.2.1) (2026-06-21)


### Bug Fixes

* **sonar:** use &lt;output&gt; instead of role=status in CopyButton (S6819) ([#61](https://github.com/Upellift99/GateCHA/issues/61)) ([df69a18](https://github.com/Upellift99/GateCHA/commit/df69a182a4a7399e5bcf2b14e3b1b19619f2ecae))

## [0.2.0](https://github.com/Upellift99/GateCHA/compare/v0.1.0...v0.2.0) (2026-06-21)


### Features

* **ui:** icon copy buttons with tooltip + colored On/Off badges ([#59](https://github.com/Upellift99/GateCHA/issues/59)) ([b7dc04b](https://github.com/Upellift99/GateCHA/commit/b7dc04b221c3233ca716e4bb412e2b7e04dc4788))

## 0.1.0 (2026-06-21)


### Features

* **altcha:** adaptive proof-of-work difficulty per key (opt-in) ([#39](https://github.com/Upellift99/GateCHA/issues/39)) ([453b69f](https://github.com/Upellift99/GateCHA/commit/453b69f20888048432b355c005311cdeb2be7f10))
* **geo:** traffic-by-country panel with flags (privacy-first, no IP stored) ([#42](https://github.com/Upellift99/GateCHA/issues/42)) ([b0fe377](https://github.com/Upellift99/GateCHA/commit/b0fe37740b526e557702a24fefcdfabe3aac67f9))
* **his:** Human Interaction Signature — Monitor mode (DIY HIS) ([#40](https://github.com/Upellift99/GateCHA/issues/40)) ([a7d49a1](https://github.com/Upellift99/GateCHA/commit/a7d49a1f0a5f4a14216f1c45913a19414a40d8a2))
* **his:** raw-signal sampling + calibration (enforcement foundation) ([#46](https://github.com/Upellift99/GateCHA/issues/46)) ([31d4739](https://github.com/Upellift99/GateCHA/commit/31d473939abf8d9a4144e4a07b42daafa92b8005))
* multi-domain + wildcard keys, login footer, brand favicon ([#50](https://github.com/Upellift99/GateCHA/issues/50)) ([19ed436](https://github.com/Upellift99/GateCHA/commit/19ed436f13d556d3e12a62d4b268c19465fd8a65))
* **ratelimit:** per-key rate limit (config + enforcement + UI) ([#37](https://github.com/Upellift99/GateCHA/issues/37)) ([d27b10e](https://github.com/Upellift99/GateCHA/commit/d27b10e7ec0254abba26a47c4bcf97ea52e7e53d))
* **security:** rate limiting, security headers & body size limits ([#26](https://github.com/Upellift99/GateCHA/issues/26)) ([771a88f](https://github.com/Upellift99/GateCHA/commit/771a88f1d8b01d9deedff2cd1d3bf3f31e4e9b4f))
* **ui:** copy buttons for key/instance URL + inject version in Docker CI ([#55](https://github.com/Upellift99/GateCHA/issues/55)) ([20f9dfb](https://github.com/Upellift99/GateCHA/commit/20f9dfbc0d74ea3cf6493bcda2742868dbd8325a))
* **ui:** dashboard footer with gatecha.org link + build version ([#47](https://github.com/Upellift99/GateCHA/issues/47)) ([c2a0124](https://github.com/Upellift99/GateCHA/commit/c2a0124165dbf86011e20f9522e75ec65ada6f33))
* **ui:** professional dashboard redesign aligned with the brand identity ([#43](https://github.com/Upellift99/GateCHA/issues/43)) ([9b02b69](https://github.com/Upellift99/GateCHA/commit/9b02b6903a6be1366ebf9adac60e80293b959a10))


### Bug Fixes

* **api:** send Cache-Control: no-store on all JSON responses ([#35](https://github.com/Upellift99/GateCHA/issues/35)) ([b9980a2](https://github.com/Upellift99/GateCHA/commit/b9980a2b85e781cf4526e50174a6b2c018924b7a))
* **captcha:** use altcha v3 `challenge` attribute, not v2 `challengeurl` ([#36](https://github.com/Upellift99/GateCHA/issues/36)) ([8c4c515](https://github.com/Upellift99/GateCHA/commit/8c4c51513ca6ea064b1ad13afd56a49318927564))
* **config:** use GATECHA_DB_DSN env var name in compose and Makefile ([#28](https://github.com/Upellift99/GateCHA/issues/28)) ([d03936a](https://github.com/Upellift99/GateCHA/commit/d03936a0f85d0e2a03f477913f2d494636296c01))
* **db:** prevent daily_stats wipe on SQLite migration (ON DELETE CASCADE) ([#29](https://github.com/Upellift99/GateCHA/issues/29)) ([1a1732e](https://github.com/Upellift99/GateCHA/commit/1a1732e71678c96e3b0ec6d3626c5c84fcba63c1))
* **ratelimit:** warn when TRUST_PROXY is off behind a proxy (login captcha breaks) ([#34](https://github.com/Upellift99/GateCHA/issues/34)) ([f9507e8](https://github.com/Upellift99/GateCHA/commit/f9507e856c124f7e7ef7f2b08ce35b6e7dc14dae))
* **sonar:** clear 2 issues on main (S1192, S7758) ([#45](https://github.com/Upellift99/GateCHA/issues/45)) ([788b4ad](https://github.com/Upellift99/GateCHA/commit/788b4ade15fe75226eb62bcf9102fe69d5f45879))
* **sonar:** hoist fetchHISCalibration to module scope (S7721) ([#48](https://github.com/Upellift99/GateCHA/issues/48)) ([d12f458](https://github.com/Upellift99/GateCHA/commit/d12f458e2da6f65cb21de15a2ab2fb7b88a2b458))
* **sonar:** split config.Load to cut cognitive complexity (S3776) ([#27](https://github.com/Upellift99/GateCHA/issues/27)) ([e0aa653](https://github.com/Upellift99/GateCHA/commit/e0aa653a6e2424e137f969f2b1d1c7dd3f5bc646))
* **sonar:** use Number.NaN over global NaN in HIS collector (S7773) ([#41](https://github.com/Upellift99/GateCHA/issues/41)) ([edb404b](https://github.com/Upellift99/GateCHA/commit/edb404b9647cc1ca7cc728a21d2c6ff50d4fe96e))


### Continuous Integration

* bump release-please-action to v5 (Node 24) ([#58](https://github.com/Upellift99/GateCHA/issues/58)) ([edba115](https://github.com/Upellift99/GateCHA/commit/edba1154d5fcfc35aa2af95698e2ff297d45763f))
