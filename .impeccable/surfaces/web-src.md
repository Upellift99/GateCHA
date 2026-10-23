---
version: 1
slug: "web-src"
primary_target: "web/src"
related_targets: []
---

# Surface brief: GateCHA admin dashboard (web/src)

Mode: Operate. Scope: the whole Vue admin (login, dashboard, API keys list/detail/form, settings), light and dark, both designed, following the system theme. The website (gatecha.org) adopts this world later; not in scope here.

Audience and job: self-hosting admins who drop in to create a key for a site, copy it, and check whether any key is failing. Proof is their own traffic; no invented claims. Constraints: no external CDN or font host (self-hosted fonts only), WCAG AA in both themes, existing routes, API and behaviour preserved, unit and e2e tests kept green.

Unresolved: line colours cycle past the palette size; the number stays the unique identity.

## Direction contract

THESIS: Every API key is a numbered line of the network; the dashboard opens on the traffic-status board that says, per line, normal service, disturbed or closed. Refuses the generic SaaS card grid with a teal accent.

OWN-WORLD: Signage navy band (#10285A) carrying the navigation; porcelain ground (#F3F5F7) / night ground (#0B1221); round line bullets in metro colours with the key's number; Overpass for UI and headings, Overpass Mono tabular figures for every count; structure by 1px hairlines, no drop-shadow cards; state colours only for state (amber disturbed, red failures, green normal, grey closed).

STORY: The admin sees at once which line is disturbed and why (failures today), follows that line by its number through list, detail and chart, and acts (open, edit, disable) with standard controls.

FIRST VIEWPORT: Navy band (mark, Dashboard / API Keys / Settings, version, Logout). Title + "Create key" right. KPI strip of four hairline-separated figures. Left two thirds: "Network status" board, one row per key: bullet, name + domain, 7-day spark, today's challenges and failures, status chip; disturbed rows first-noticed. Right third: traffic by country, HIS summary. Below: 30-day chart whose legend reuses the bullets; selecting a line dims the others.

FORM: "Info trafic", metro/RER signage, position 3 of 7 on the ordered list (postal franking, toll, metro signage, IGN map, datasheet, ledger, weather bulletin). Raises: tabular counts on strict columns (from Ikeda field), hairline grid only (from design annual), focus dims the rest (from streaming wall), one number per key everywhere (from Factory catalogue). seed:2d3285152c74180d

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
