# Changelog

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
