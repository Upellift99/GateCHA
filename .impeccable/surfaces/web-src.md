---
version: 2
slug: "web-src"
primary_target: "web/src"
related_targets: []
---

# Surface brief: GateCHA admin dashboard (web/src)

Mode: Operate. Scope: the whole Vue admin (login, dashboard, API keys list/detail/form, settings), light and dark, both designed, following the system theme.

Audience and job: self-hosting admins who drop in to create a key for a site, copy it, and check whether any key is failing. Proof is their own traffic; no invented claims. Constraints: no external CDN or font host (self-hosted fonts only), WCAG AA in both themes, existing routes, API and behaviour preserved, unit and e2e tests kept green.

History: v1 ("Info trafic", 2026-10-23) dressed every key as a numbered metro line. The owner rejected it on 2026-10-24: the metaphor got in the way of the one question the page answers. Do not bring back line numbers, coloured key bullets, transit vocabulary or any borrowed metaphor. Mockups of the three replacement options live in `.impeccable/mocks/dash/` (option 2 chosen).

## Direction contract

THESIS: The dashboard answers "does any of my sites have a problem right now?" before anything else. Failing keys are named first, with their counts; when nothing fails it says so in one line, and says nothing about keys working before any verification has happened.

OWN-WORLD: Navy header band (#10285A) carrying the navigation and the mark (a "G" whose bar turns into a yellow check); porcelain ground (#F3F5F7) / night ground (#0B1221); Overpass for UI and headings, Overpass Mono tabular figures for every count; structure by 1px hairlines, no drop-shadow cards; state colours only for state (amber failing, red failures, green verified, grey disabled).

STORY: The admin reads the banner (which key, how many failures, what to check), sees today's funnel (challenges issued, verified, failed), finds the key in its group (Needs attention, Working, No traffic today, Disabled), and opens it.

FIRST VIEWPORT: Band (mark, Dashboard / API Keys / Settings, version, Logout). Title with today's UTC date, "Create key" right. Attention banner. Left: "Today's verifications" funnel, then the keys grouped by state with a 7-day trend and today's counts. Right: 30-day verified/failed bars with challenges as a line, countries, bot-suspected share.

FORM: Plain operations console, attention first. Statuses read OK, No traffic, Failing, Disabled.

FINISH: unreviewed and undocumented is unfinished; this build ends with the review, DESIGN.md, and every shipping raster carrying its provenance
