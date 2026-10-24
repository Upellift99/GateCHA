---
name: GateCHA
description: Self-hosted ALTCHA proof-of-work CAPTCHA; the admin is a plain operations console that says first whether any key is failing.
colors:
  action: "#0F766E"
  action-hover: "#115E59"
  action-ink: "#FFFFFF"
  link: "#0F766E"
  link-hover: "#115E59"
  focus: "#0F766E"
  band: "#FFFFFF"
  band-ink: "#0F172A"
  band-muted: "#475569"
  band-rule: "#E2E8F0"
  ok: "#15803D"
  ok-soft: "#E7F8EC"
  ok-edge: "#A7E0B8"
  warn: "#E8A400"
  warn-ink: "#2A1E00"
  warn-text: "#8A5A00"
  warn-soft: "#FFF7E0"
  warn-edge: "#F1D88E"
  fail: "#C8001A"
  fail-soft: "#FDECEE"
  fail-edge: "#F4B6BE"
  danger: "#C8001A"
  danger-hover: "#A30015"
  disabled: "#ECEDEF"
  disabled-ink: "#5E6166"
  ground: "#F8FAFC"
  panel: "#FFFFFF"
  sunken: "#F1F5F9"
  ink: "#0F172A"
  ink-2: "#334155"
  muted: "#475569"
  rule: "#E2E8F0"
  rule-soft: "#EEF2F6"
  rule-strong: "#CBD5E1"
  code: "#1E293B"
  code-ink: "#E2E8F0"
  backdrop: "rgb(15 23 42 / 0.6)"
typography:
  display:
    fontFamily: "Inter, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "32px"
    fontWeight: 700
    lineHeight: "36px"
    letterSpacing: "-0.025em"
  headline:
    fontFamily: "Inter, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "24px"
    fontWeight: 700
    lineHeight: "32px"
    letterSpacing: "-0.025em"
  title:
    fontFamily: "Inter, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "17px"
    fontWeight: 600
    lineHeight: "1.5"
  body:
    fontFamily: "Inter, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: "20px"
  caption:
    fontFamily: "Inter, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "12px"
    fontWeight: 400
    lineHeight: "16px"
  label:
    fontFamily: "Inter, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "11px"
    fontWeight: 600
    letterSpacing: "0.09em"
  figure:
    fontFamily: "Overpass Mono, ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: "32px"
    fontWeight: 600
    lineHeight: "40px"
    letterSpacing: "-0.025em"
    fontFeature: "tnum"
  data:
    fontFamily: "Overpass Mono, ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: "20px"
    fontFeature: "tnum"
rounded:
  chip: "3px"
  sm: "4px"
  md: "6px"
  mark: "7px"
  full: "9999px"
spacing:
  row: "10px"
  panel-head: "12px"
  panel-inset: "20px"
  card: "24px"
  section: "24px"
  band: "56px"
  aside: "380px"
  container: "1280px"
components:
  button-primary:
    backgroundColor: "{colors.action}"
    textColor: "{colors.action-ink}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
    height: "40px"
  button-primary-hover:
    backgroundColor: "{colors.action-hover}"
    textColor: "{colors.action-ink}"
  button-secondary:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
  button-secondary-hover:
    backgroundColor: "{colors.sunken}"
  button-danger:
    backgroundColor: "{colors.danger}"
    textColor: "{colors.action-ink}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
  button-danger-hover:
    backgroundColor: "{colors.danger-hover}"
  input:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "8px 12px"
  panel:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    padding: "{spacing.panel-inset}"
  card:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    padding: "{spacing.card}"
  status-failing:
    backgroundColor: "{colors.warn}"
    textColor: "{colors.warn-ink}"
    typography: "{typography.caption}"
    rounded: "{rounded.chip}"
    padding: "2px 8px"
  status-ok:
    textColor: "{colors.ok}"
    rounded: "{rounded.chip}"
    padding: "2px 8px"
  status-idle:
    textColor: "{colors.muted}"
    rounded: "{rounded.chip}"
    padding: "2px 8px"
  status-disabled:
    backgroundColor: "{colors.disabled}"
    textColor: "{colors.disabled-ink}"
    rounded: "{rounded.chip}"
    padding: "2px 8px"
  attention-banner:
    backgroundColor: "{colors.warn-soft}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    padding: "16px 20px"
  all-clear-line:
    backgroundColor: "{colors.ok-soft}"
    textColor: "{colors.ok}"
    rounded: "{rounded.md}"
    padding: "12px 20px"
  nav-band:
    backgroundColor: "{colors.band}"
    textColor: "{colors.band-ink}"
    height: "{spacing.band}"
  nav-tab:
    textColor: "{colors.band-muted}"
    typography: "{typography.body}"
    padding: "0 16px"
  nav-tab-active:
    textColor: "{colors.band-ink}"
  code-block:
    backgroundColor: "{colors.code}"
    textColor: "{colors.code-ink}"
    rounded: "{rounded.md}"
    padding: "16px"
---

# Design System: GateCHA

## Overview

**Creative North Star: "The Plain Console"**

The admin answers one question before anything else: does any of my sites have a problem right now? Every surface is built around that order. Failing keys are named first with their counts, then today's verification funnel, then every key grouped by what the admin has to do about it, then the 30-day history. When nothing fails the page says so in one line, and it says nothing about keys working until a verification has actually happened today.

The world is GateCHA's own, shared with gatecha.org: a light header band (white with a bottom hairline; panel colour in dark mode) carrying the teal shield mark and the two-colour wordmark, over the slate-grey ground the site uses (a night ground in dark mode). Inter sets every word and Overpass Mono tabular figures set every count. Structure is drawn with 1px hairlines; there are no drop shadows anywhere. The GateCHA teal is the one brand colour and carries actions, links and focus; green, amber, red and grey are reserved for state. The density is that of an operations table: 14px body, 10px row padding, panels inset 20px, and the numbers lead.

Brand history: on 2026-10-24 the owner rejected the transit metro metaphor, then a navy band with a yellow "G" mark, and kept GateCHA's teal and shield. Neither rejected identity is to be re-proposed, nor replaced by another borrowed metaphor.

**Key Characteristics:**
- Attention first: failing keys open the page and sort first in every list.
- One brand identity with gatecha.org: teal shield mark, "Gate" + teal "CHA" wordmark, slate neutrals, Inter.
- Hairline panels on a tinted ground; no shadow, no gradient, no decorative colour.
- Counts in Overpass Mono with tabular figures, formatted `en-US`.
- Light and dark both designed, following the system theme through the same token names.
- Self-hosted fonts only; nothing loads from a third-party host.

## Colors

The slate neutrals of gatecha.org, one teal for everything the admin can act on, and four state colours that appear only when they mean something.

### Primary
- **GateCHA Teal** (action, link, focus): primary buttons, text links, focus rings, the active tab underline, checkboxes and the caret, the "CHA" of the wordmark, the mark's tile, the 7-day sparklines of healthy keys. Hover deepens to action-hover / link-hover. In dark mode the button fill stays teal (hover #0D9488) while links and focus lift to a light teal (#5EEAD4, hover #99F6E4) so they read on the night ground.

### State
- **Verified Green** (ok, ok-soft, ok-edge): verified counts, the OK status, the all-clear line, the verified segment of the funnel and country bars. A true green, distinct in hue from the teal, so state never reads as brand. The 30-day chart fills verified bars with ok-edge so the red failed segment stays readable on top. Dark: #4ADE80.
- **Failing Amber** (warn, warn-ink, warn-soft, warn-edge, warn-text): the Failing status chip, the attention banner, and the row tint of keys that need attention.
- **Failure Red** (fail, fail-soft, fail-edge): failure counts and failed segments. Danger buttons use the separate danger pair, identical to fail in light mode and darker in dark mode (#C92A3E) so white text holds AA.
- **Disabled Grey** (disabled, disabled-ink): the Disabled status chip.

### Neutral
- **Slate Ground** (ground): page background. Dark: #0B1120.
- **Panel White** (panel): panels, cards, inputs, the footer. Dark: #131C2E.
- **Sunken** (sunken): group header rows, table heads, inline code, meter tracks, secondary-button and Logout hover. Dark: #0F1726.
- **Ink / Ink 2 / Muted** (ink, ink-2, muted): headings, names and figures; field labels and secondary text; captions, column labels, empty states. Dark: #F1F5F9 / #CBD5E1 / #94A3B8.
- **Hairlines** (rule, rule-soft, rule-strong): panel borders; row separators and chart gridlines; input and secondary-button borders, disabled sparklines.
- **Band** (band, band-ink, band-muted, band-rule): the header band takes the panel surface, ink, muted and rule values; they are separate names so the header can be retuned alone.
- **Code Slate** (code, code-ink): integration snippets, dark in both themes.

### Named Rules
**The One Teal Rule.** Teal means "you can act here or this is GateCHA": actions, links, focus, the mark and the wordmark. It is never used for state, and green is never used for an action.

**The State-Only Rule.** Green, amber, red and grey mean verified, failing, failed and disabled, and nothing else. They never decorate. The one standing exception is text selection, which borrows the amber.

**The Words-First Rule.** Colour backs a word, it never replaces one. Every status chip spells OK, No traffic, Failing or Disabled; every sparkline carries an aria-label; a disabled key's trend is dashed as well as grey.

**The Same Names Rule.** Dark mode redefines the same custom properties under `prefers-color-scheme: dark`. Components use semantic tokens only and never branch on the theme. Text pairs are meant to hold 4.5:1 on ground, panel and sunken in both themes; strokes and marks 3:1.

## Typography

**Display Font:** Inter (with ui-sans-serif, system-ui)
**Body Font:** Inter
**Label/Mono Font:** Overpass Mono (with ui-monospace, SFMono-Regular, Menlo)

**Character:** Inter carries the UI because the owner found Overpass at bold and extrabold hard to read, and because gatecha.org already sets in Inter: legibility and continuity with the site. Overpass Mono stays for counts and key ids, so numbers align in columns. Both ship as self-hosted variable woff2, Latin subset (`web/src/assets/fonts/inter-latin.woff2`). The detector's overused-font ignore for Inter in `.impeccable/config.json` is a sanctioned exception for that reason.

### Hierarchy
- **Display** (700, 32px, 36px): the Dashboard page title only.
- **Headline** (700, 24px, 32px): page titles on API Keys, key detail, key form, Settings, Sign in.
- **Title** (600, 17px): every panel and dialog heading, the attention banner heading.
- **Body** (400, 14px, 20px): table cells and form text. Buttons, tabs, links, key names and country names are 14px 600; field labels are 14px 500 in ink-2. Help text under fields caps at 52ch.
- **Caption** (400, 12px to 13px): panel sub-lines, domains under key names, chart legends.
- **Label** (600, 11px, 0.09em, uppercase): column headers, group header rows and the funnel's figure labels. It names the data directly beneath it and is never placed above a heading.
- **Figure** (Overpass Mono 600, 32px/40px, 26px/32px below 640px): the funnel's three counts.
- **Data** (Overpass Mono, 14px or 12px): counts in tables, key IDs, secrets, domains, version, chart axes.

### Named Rules
**The Light Weights Rule.** Bold (700) is for page titles and the wordmark only; panel headings, names, labels, chips and buttons are semibold (600). No extrabold or black anywhere.

**The Tabular Figures Rule.** Every count is set in Overpass Mono with `tabular-nums` and formatted with `toLocaleString('en-US')`. No bare `toLocaleString()`, no proportional digits in a column.

**The UTC Rule.** "Today" is the current UTC day and every date is formatted in UTC (`en-US`, short month and day). The dashboard title says "(UTC)".

## Layout

One centred container (max 1280px) with 16px side padding, 24px from 640px, 32px from 1024px, and 32px of vertical padding. Pages stack sections 24px apart.

The dashboard is a two-column grid from 1024px: a fluid main column (funnel, then keys) and a fixed 380px aside (30-day chart, countries, bot-suspected). Below 1024px everything stacks in reading order, the attention banner always above both columns. The keys table hides its 7-day column below 640px; the API Keys table drops columns at 640px and 768px and moves the status chip under the key name on mobile.

Dashboard panels use a header strip (20px inset, 12px tall padding, hairline below) over a 20px body. Detail, form and settings pages use cards with 24px padding; CountryTraffic switches between the two insets with its `roomy` prop. Rows are 10px tall padding with rule-soft separators. Forms cap at 576px, login at 384px. Settings splits into a 5-column grid from 1024px.

### Named Rules
**The Attention-First Rule.** Failing keys come first everywhere: named in the banner, first group in the keys panel, rank 0 in sorting (failing, OK, no traffic, disabled). A key is failing when today it has at least 5 failures and those are at least 20% of its verifications.

**The Honest Quiet Rule.** "All keys are working today." appears only when today has verifications. With active keys but no verification yet, the line reads "No verification yet today." in muted text on a plain panel.

## Elevation & Depth

Flat. No element casts a shadow. Depth comes from tonal layering (ground under panel, sunken inside panel) and 1px hairlines; the header band is separated from the page by its bottom hairline, not by colour. The only lifted layer is the modal dialog, which sits on a translucent backdrop (`backdrop` token) with a small blur, framed by a rule-strong border, not a shadow.

### Named Rules
**The Hairline Rule.** Structure is drawn with 1px rules: panel borders and the band edge in rule, row separators and chart gridlines in rule-soft, controls in rule-strong. No shadow cards.

## Shapes

Small, consistent corners: 6px on panels, cards, buttons, inputs, banners and code blocks; 4px on inline code, the Logout button and small tags; 3px on status chips, the funnel meter and country code tags; full round on status dots, the banner's alert disc, the toggle switch and the thin country meters. The attention banner adds a 4px amber left edge to its hairline. The mark is a teal rounded tile (7px on a 32px grid, drawn at 28px in the header) holding a white shield outline with a white check.

## Components

### Buttons
Plain and direct; colour tells weight, not decoration.
- **Shape:** gently rounded (6px).
- **Primary:** teal with white text, 14px semibold, 8px by 16px; page-level primaries ("Create key") are 40px tall with a leading plus icon.
- **Hover / Focus:** hover deepens to action-hover; focus shows a 2px focus-colour ring (offset 2px on page-level primaries). Disabled is 50% opacity.
- **Secondary:** panel background, rule-strong border, ink text; hover fills sunken. Used for Cancel, Disable, Edit, Open key, pagination.
- **Danger:** danger red with white text, for Delete and destructive dialog confirms only.
- **In the header:** Logout is a 32px outlined button in band-rule with band-ink text and a sunken hover.

### Status chip
The one status vocabulary: OK, No traffic, Failing, Disabled, from `lib/status.ts`.
- **Style:** 12px semibold, 3px corners, 2px by 8px. OK is green text with a 7px dot; No traffic is muted text; Failing is amber fill with warn-ink text, a dot, and today's failure count in red mono on a panel-white inset, followed by "failures today" from 640px; Disabled is grey fill.
- **Compact:** drops the words "failures today" for tables without room; the count stays.

### Cards / Containers
- **Corner Style:** 6px.
- **Background:** panel on ground.
- **Shadow Strategy:** none; see Elevation & Depth.
- **Border:** 1px rule.
- **Internal Padding:** 20px for dashboard panels with a header strip, 24px for detail, form, settings and login cards.

### Inputs / Fields
- **Style:** panel background, 1px rule-strong border, 6px corners, 8px by 12px; placeholders in muted. Checkboxes and the caret take the action teal.
- **Focus:** border turns focus teal with a focus-colour ring.
- **Toggle switch:** a 44 by 24 SVG track (action when on, rule-strong when off) with a 16px white knob that slides 20px; focus ring on the track.

### Navigation
The header band, 56px tall and sticky: white with a bottom band-rule hairline in light mode, panel colour in dark. The mark and the wordmark lead in 18px bold; tabs (Dashboard, API Keys, Settings) are 14px semibold in band-muted with a transparent 3px bottom edge, turning band-ink with a teal bottom edge when active. Version in mono and Logout sit at the right. Below 640px the tabs drop to a second row of the band, 44px tall. Signed out, the band carries only the mark and wordmark.

### Brand mark and wordmark
The shield with a check, white on a teal tile, is the GateCHA mark; `BrandMark.vue` and `public/favicon.svg` are the same drawing as gatecha.org and the WordPress.org plugin icon and banners, and change together. The wordmark is "Gate" in ink and "CHA" in teal (`text-action`), set wherever the product names itself: the header, the signed-out header and the footer.

### Attention banner
The first element under the page title. With failing keys: amber-soft fill, amber-edge hairline, 4px amber left edge, an amber disc, a 17px semibold heading ("1 key needs attention" / "N keys need attention"), the key name with failed of verified counts and percentage, and an "Open key" secondary button for a single key. With none failing it collapses to one line: ok-soft with a green dot when verifications happened today, plain muted panel otherwise.

### Today's verifications funnel
A panel with three figures (Challenges, Verified in green, Failed in red when non-zero) and, once anything was checked, a 14px verified/failed meter with mono counts under each end.

### Keys today
Every key in one table grouped Needs attention, Working, No traffic today, Disabled, each group headed by a sunken label row with its count. Rows show the name (semibold, link on hover), first domain with "+N more", a 7-day challenges sparkline (88 by 22, link teal; fail red for failing keys; dashed rule-strong for disabled), today's challenges and failures. Failing rows are tinted warn-soft and their failure count is semibold red.

### Charts
The dashboard's 30-day chart is inline SVG: stacked bars (verified in ok-edge, failed in fail) with challenges issued as a 1.75px ink line, two rule-soft gridlines, mono date ends, a legend and a full-text aria-label. Countries are listed with an ISO code tag (never a flag emoji), the count, the failure share (red from 50%) and a thin verified/failed meter.

### Code block
Code Slate background in both themes, code-ink text, 14px mono, 16px padding with room for a copy button in the corner.

## Do's and Don'ts

### Do:
- **Do** name failing keys first, with failed-of-verified counts, before any chart.
- **Do** use the four statuses OK, No traffic, Failing, Disabled from `lib/status.ts`, always as words, with colour behind them.
- **Do** set every count in Overpass Mono with `tabular-nums` and `toLocaleString('en-US')`.
- **Do** set UI text in Inter at 400 to 700: bold for page titles and the wordmark, semibold for headings, names, labels and controls.
- **Do** draw structure with 1px hairlines (rule, rule-soft, rule-strong) on panel over ground.
- **Do** use semantic tokens only, so dark mode follows by redefining the same names.
- **Do** keep the shield mark one drawing across `BrandMark.vue`, `public/favicon.svg`, gatecha.org and the WordPress.org assets.
- **Do** say "today" in UTC and format dates in UTC.

### Don't:
- **Don't** bring back the metro metaphor, the navy band or the yellow "G" mark, nor any other borrowed metaphor: no numbering keys, no per-key colours, no themed status vocabulary.
- **Don't** use teal for state or green for an action.
- **Don't** use green, amber, red or grey for anything but state.
- **Don't** use extrabold or black weights.
- **Don't** add drop shadows or shadowed cards.
- **Don't** say keys are working before any verification has happened today.
- **Don't** use flag emoji for countries; use the ISO code tag.
- **Don't** load fonts, scripts or images from a third-party host.
