---
name: GateCHA
description: Self-hosted ALTCHA proof-of-work CAPTCHA; the admin is a plain operations console that says first whether any key is failing.
colors:
  band: "#10285A"
  band-ink: "#FFFFFF"
  band-muted: "#C6D0E4"
  band-rule: "#2A4479"
  action: "#10285A"
  action-hover: "#0B1E45"
  action-ink: "#FFFFFF"
  link: "#1F4FC1"
  link-hover: "#163B94"
  focus: "#1F4FC1"
  mark-yellow: "#FFCD00"
  ok: "#00754E"
  ok-soft: "#E3F4EC"
  ok-edge: "#A9DCC4"
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
  ground: "#F3F5F7"
  panel: "#FFFFFF"
  sunken: "#F7F8FA"
  ink: "#14213D"
  ink-2: "#3A4560"
  muted: "#5B6478"
  rule: "#DDE2E8"
  rule-soft: "#EBEEF2"
  rule-strong: "#B9C1CD"
  code: "#0E1A33"
  code-ink: "#D7E3FF"
  backdrop: "rgb(11 18 33 / 0.6)"
typography:
  display:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "32px"
    fontWeight: 800
    lineHeight: "36px"
    letterSpacing: "-0.025em"
  headline:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "24px"
    fontWeight: 800
    lineHeight: "32px"
    letterSpacing: "-0.025em"
  title:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "17px"
    fontWeight: 800
    lineHeight: "1.5"
  body:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: "20px"
  caption:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "12px"
    fontWeight: 400
    lineHeight: "16px"
  label:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "11px"
    fontWeight: 700
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

The world is a navy header band over a porcelain ground (night ground in dark mode), Overpass for every word and Overpass Mono tabular figures for every count. Structure is drawn with 1px hairlines; there are no drop shadows anywhere. Colour is reserved for state: green verified, red failures, amber failing, grey disabled. The density is that of an operations table: 14px body, 10px row padding, panels inset 20px, and the numbers lead.

The 2026-10-23 transit metaphor was retired by the owner on 2026-10-24 and must not return in any form, nor be replaced by another borrowed metaphor.

**Key Characteristics:**
- Attention first: failing keys open the page and sort first in every list.
- Navy band in both themes, carrying the mark, the navigation, the version and Logout.
- Hairline panels on a tinted ground; no shadow, no gradient, no decorative colour.
- Counts in Overpass Mono with tabular figures, formatted `en-US`.
- Light and dark both designed, following the system theme through the same token names.
- Self-hosted fonts only; nothing loads from a third-party host.

## Colors

A cool navy-and-porcelain neutral set with four state colours that appear only when they mean something.

### Primary
- **Header Navy** (band, action): the header band in both themes, the brand tile, and the primary button in light mode. In dark mode the action lifts to a brighter blue (#3366E0, hover #2B59C3) so it holds contrast on the night ground; the band stays navy.
- **Link Blue** (link, focus): text links, focus rings, the challenges sparkline and per-key chart line. Dark mode: #8AB0FF.

### Secondary
- **Check Yellow** (mark-yellow): the check stroke in the mark only. It is not a UI token and is not used anywhere else.

### State
- **Verified Green** (ok, ok-soft, ok-edge): verified counts, the OK status, the all-clear line, the verified segment of the funnel and country bars. The 30-day chart fills verified bars with ok-edge so the red failed segment stays readable on top.
- **Failing Amber** (warn, warn-ink, warn-soft, warn-edge, warn-text): the Failing status chip, the attention banner, and the row tint of keys that need attention.
- **Failure Red** (fail, fail-soft, fail-edge): failure counts and failed segments. Danger buttons use the separate danger pair, identical to fail in light mode (#C8001A) and darker in dark mode (#C92A3E) so white text holds AA.
- **Disabled Grey** (disabled, disabled-ink): the Disabled status chip.

### Neutral
- **Porcelain Ground** (ground): page background. Dark: #0B1221.
- **Panel White** (panel): panels, cards, inputs, the footer. Dark: #111A2E.
- **Sunken** (sunken): group header rows, inline code, meter tracks, secondary-button hover. Dark: #0E1627.
- **Ink / Ink 2 / Muted** (ink, ink-2, muted): headings and figures; field labels and secondary text; captions, column labels, empty states. Dark: #E7EBF3 / #C3CAD8 / #9AA4B8.
- **Hairlines** (rule, rule-soft, rule-strong): panel borders; row separators and chart gridlines; input and secondary-button borders, disabled sparklines.
- **Code Night** (code, code-ink): integration snippets, dark in both themes.

### Named Rules
**The State-Only Rule.** Green, amber, red and grey mean verified, failing, failed and disabled, and nothing else. They never decorate. The one standing exception is text selection, which borrows the amber.

**The Words-First Rule.** Colour backs a word, it never replaces one. Every status chip spells OK, No traffic, Failing or Disabled; every sparkline carries an aria-label; a disabled key's trend is dashed as well as grey.

**The Same Names Rule.** Dark mode redefines the same custom properties under `prefers-color-scheme: dark`. Components use semantic tokens only and never branch on the theme. Text pairs hold 4.5:1 on ground, panel and sunken in both themes; strokes and marks hold 3:1.

## Typography

**Display Font:** Overpass (with ui-sans-serif, system-ui)
**Body Font:** Overpass
**Label/Mono Font:** Overpass Mono (with ui-monospace, SFMono-Regular, Menlo)

**Character:** One sturdy humanist sans in heavy weights for headings and plain weights for text, paired with its own mono for figures, so numbers align in columns without changing voice. Both ship as self-hosted variable woff2, Latin subset.

### Hierarchy
- **Display** (800, 32px, 36px): the Dashboard page title only.
- **Headline** (800, 24px, 32px): page titles on API Keys, key detail, key form, Settings, Sign in.
- **Title** (800, 17px): every panel and dialog heading, the attention banner heading.
- **Body** (400 to 600, 14px, 20px): table cells, buttons, links, form text. Field labels are 14px 500 in ink-2. Help text under fields caps at 52ch.
- **Caption** (400, 12px to 13px): panel sub-lines, domains under key names, chart legends.
- **Label** (700, 11px, 0.09em, uppercase): column headers, group header rows and the funnel's figure labels. It names the data directly beneath it and is never placed above a heading.
- **Figure** (Overpass Mono 600, 32px/40px, 26px/32px below 640px): the funnel's three counts.
- **Data** (Overpass Mono, 14px or 12px): counts in tables, key IDs, secrets, domains, version, chart axes.

### Named Rules
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

Flat. No element casts a shadow. Depth comes from tonal layering (ground under panel, sunken inside panel) and 1px hairlines. The only lifted layer is the modal dialog, which sits on a translucent backdrop (`backdrop` token) with a small blur, framed by a rule-strong border, not a shadow.

### Named Rules
**The Hairline Rule.** Structure is drawn with 1px rules: panel borders in rule, row separators and chart gridlines in rule-soft, controls in rule-strong. No shadow cards.

## Shapes

Small, consistent corners: 6px on panels, cards, buttons, inputs, banners and code blocks; 4px on inline code, the Logout button and small tags; 3px on status chips, the funnel meter and country code tags; full round on status dots, the banner's alert disc, the toggle switch and the thin country meters. The attention banner adds a 4px amber left edge to its hairline. The mark is a navy rounded tile (7px on 32px) with a 1.5px band-rule inner stroke.

## Components

### Buttons
Plain and direct; colour tells weight, not decoration.
- **Shape:** gently rounded (6px).
- **Primary:** action navy with white text, 14px semibold, 8px by 16px; page-level primaries ("Create key") are 40px tall with a leading plus icon.
- **Hover / Focus:** hover darkens to action-hover; focus shows a 2px focus-colour ring (offset 2px on page-level primaries). Disabled is 50% opacity.
- **Secondary:** panel background, rule-strong border, ink text; hover fills sunken. Used for Cancel, Disable, Edit, Open key, pagination.
- **Danger:** danger red with white text, for Delete and destructive dialog confirms only.
- **On the band:** Logout is a 32px outlined button in band-rule with a white 10% hover.

### Status chip
The one status vocabulary: OK, No traffic, Failing, Disabled, from `lib/status.ts`.
- **Style:** 12px bold, 3px corners, 2px by 8px. OK is green text with a 7px dot; No traffic is muted text; Failing is amber fill with warn-ink text, a dot, and today's failure count in red mono on a panel-white inset, followed by "failures today" from 640px; Disabled is grey fill.
- **Compact:** drops the words "failures today" for tables without room; the count stays.

### Cards / Containers
- **Corner Style:** 6px.
- **Background:** panel on ground.
- **Shadow Strategy:** none; see Elevation & Depth.
- **Border:** 1px rule.
- **Internal Padding:** 20px for dashboard panels with a header strip, 24px for detail, form, settings and login cards.

### Inputs / Fields
- **Style:** panel background, 1px rule-strong border, 6px corners, 8px by 12px; placeholders in muted. Checkboxes and the caret take the action colour.
- **Focus:** border turns focus blue with a focus-colour ring.
- **Toggle switch:** a 44 by 24 SVG track (action when on, rule-strong when off) with a 16px white knob that slides 20px; focus ring on the track.

### Navigation
The navy band, 56px tall, sticky, identical in both themes. The mark and "GateCHA" in 18px extrabold lead; tabs (Dashboard, API Keys, Settings) are 14px semibold in band-muted with a transparent 3px bottom edge, turning band-ink with a band-ink bottom edge when active. Version in mono and Logout sit at the right. Below 640px the tabs drop to a second row of the band, 44px tall. Signed out, the band carries only the mark.

### Attention banner
The first element under the page title. With failing keys: amber-soft fill, amber-edge hairline, 4px amber left edge, an amber disc, a 17px heading ("1 key needs attention" / "N keys need attention"), the key name with failed of verified counts and percentage, and an "Open key" secondary button for a single key. With none failing it collapses to one line: ok-soft with a green dot when verifications happened today, plain muted panel otherwise.

### Today's verifications funnel
A panel with three figures (Challenges, Verified in green, Failed in red when non-zero) and, once anything was checked, a 14px verified/failed meter with mono counts under each end.

### Keys today
Every key in one table grouped Needs attention, Working, No traffic today, Disabled, each group headed by a sunken label row with its count. Rows show the name (bold, link on hover), first domain with "+N more", a 7-day challenges sparkline (88 by 22, link blue; fail red for failing keys; dashed rule-strong for disabled), today's challenges and failures. Failing rows are tinted warn-soft and their failure count is bold red.

### Charts
The dashboard's 30-day chart is inline SVG: stacked bars (verified in ok-edge, failed in fail) with challenges issued as a 1.75px ink line, two rule-soft gridlines, mono date ends, a legend and a full-text aria-label. Countries are listed with an ISO code tag (never a flag emoji), the count, the failure share (bold red from 50%) and a thin verified/failed meter.

### Code block
Code Night background in both themes, code-ink text, 14px mono, 16px padding with room for a copy button in the corner.

## Do's and Don'ts

### Do:
- **Do** name failing keys first, with failed-of-verified counts, before any chart.
- **Do** use the four statuses OK, No traffic, Failing, Disabled from `lib/status.ts`, always as words, with colour behind them.
- **Do** set every count in Overpass Mono with `tabular-nums` and `toLocaleString('en-US')`.
- **Do** draw structure with 1px hairlines (rule, rule-soft, rule-strong) on panel over ground.
- **Do** use semantic tokens only, so dark mode follows by redefining the same names.
- **Do** keep the navy band and the mark identical in both themes; the mark in `BrandMark.vue` and `public/favicon.svg` is one drawing.
- **Do** say "today" in UTC and format dates in UTC.

### Don't:
- **Don't** bring back the retired metaphor or any other borrowed metaphor: no numbering keys, no per-key colours, no themed status vocabulary.
- **Don't** use green, amber, red or grey for anything but state.
- **Don't** add drop shadows or shadowed cards.
- **Don't** say keys are working before any verification has happened today.
- **Don't** use flag emoji for countries; use the ISO code tag.
- **Don't** load fonts, scripts or images from a third-party host.
