---
name: GateCHA
description: Self-hosted ALTCHA proof-of-work CAPTCHA; the admin reads as a transit network status board, "Info trafic".
colors:
  signage-navy: "#10285A"
  signage-navy-deep: "#0B1E45"
  band-ink: "#FFFFFF"
  band-muted: "#C6D0E4"
  band-rule: "#2A4479"
  route-blue: "#1F4FC1"
  route-blue-deep: "#163B94"
  porcelain-ground: "#F3F5F7"
  panel-white: "#FFFFFF"
  sunken-grey: "#F7F8FA"
  ink-navy: "#14213D"
  ink-secondary: "#3A4560"
  muted-slate: "#5B6478"
  hairline: "#DDE2E8"
  hairline-soft: "#EBEEF2"
  hairline-strong: "#B9C1CD"
  normal-green: "#00754E"
  normal-green-soft: "#E3F4EC"
  normal-green-edge: "#A9DCC4"
  disturbed-amber: "#E8A400"
  disturbed-amber-ink: "#2A1E00"
  disturbed-amber-text: "#8A5A00"
  disturbed-amber-soft: "#FFF7E0"
  disturbed-amber-edge: "#F1D88E"
  failure-red: "#C8001A"
  failure-red-soft: "#FDECEE"
  failure-red-edge: "#F4B6BE"
  danger-red-deep: "#A30015"
  closed-grey: "#ECEDEF"
  closed-grey-ink: "#5E6166"
  code-night: "#0E1A33"
  code-night-ink: "#D7E3FF"
  line-1: "#FFCD00"
  line-1-stroke: "#B38F00"
  line-2: "#00A88F"
  line-2-stroke: "#00A88F"
  line-3: "#6E4C9F"
  line-3-stroke: "#6E4C9F"
  line-4: "#F28E1C"
  line-4-stroke: "#C86A00"
  line-5: "#00AEEF"
  line-5-stroke: "#0089C2"
  line-6: "#84B817"
  line-6-stroke: "#5E8A00"
  line-7: "#A0006E"
  line-7-stroke: "#A0006E"
  line-8: "#8D8D8D"
  line-8-stroke: "#767676"
typography:
  display:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif"
    fontSize: "32px"
    fontWeight: 800
    lineHeight: "36px"
    letterSpacing: "-0.025em"
  headline:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif"
    fontSize: "24px"
    fontWeight: 800
    lineHeight: "32px"
    letterSpacing: "-0.025em"
  title:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif"
    fontSize: "17px"
    fontWeight: 800
    lineHeight: "1.5"
  body:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: "20px"
  caption:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif"
    fontSize: "12px"
    fontWeight: 400
    lineHeight: "16px"
  label:
    fontFamily: "Overpass, ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif"
    fontSize: "11px"
    fontWeight: 700
    lineHeight: "16px"
    letterSpacing: "0.09em"
  figure:
    fontFamily: "'Overpass Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: "32px"
    fontWeight: 600
    lineHeight: "40px"
    letterSpacing: "-0.025em"
    fontFeature: "tnum"
  data:
    fontFamily: "'Overpass Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: "20px"
    fontFeature: "tnum"
  axis:
    fontFamily: "'Overpass Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: "11px"
    fontWeight: 400
    fontFeature: "tnum"
rounded:
  chip: "3px"
  sm: "4px"
  md: "6px"
  full: "9999px"
spacing:
  row: "8px"
  panel-head: "12px"
  inset: "20px"
  card: "24px"
  gutter: "24px"
  page: "32px"
components:
  signage-band:
    backgroundColor: "{colors.signage-navy}"
    textColor: "{colors.band-ink}"
    height: "56px"
  nav-tab:
    textColor: "{colors.band-muted}"
    typography: "{typography.body}"
    padding: "0 16px"
  nav-tab-active:
    textColor: "{colors.band-ink}"
  button-primary:
    backgroundColor: "{colors.signage-navy}"
    textColor: "{colors.band-ink}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
  button-primary-hover:
    backgroundColor: "{colors.signage-navy-deep}"
  button-header-cta:
    backgroundColor: "{colors.signage-navy}"
    textColor: "{colors.band-ink}"
    rounded: "{rounded.md}"
    padding: "0 16px"
    height: "40px"
  button-secondary:
    backgroundColor: "{colors.panel-white}"
    textColor: "{colors.ink-navy}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
  button-secondary-hover:
    backgroundColor: "{colors.sunken-grey}"
  button-danger:
    backgroundColor: "{colors.failure-red}"
    textColor: "{colors.band-ink}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
  button-danger-hover:
    backgroundColor: "{colors.danger-red-deep}"
  input-text:
    backgroundColor: "{colors.panel-white}"
    textColor: "{colors.ink-navy}"
    rounded: "{rounded.md}"
    padding: "8px 12px"
  panel:
    backgroundColor: "{colors.panel-white}"
    rounded: "{rounded.md}"
  status-chip-normal:
    textColor: "{colors.normal-green}"
    typography: "{typography.caption}"
    rounded: "{rounded.chip}"
    padding: "2px 8px"
  status-chip-disturbed:
    backgroundColor: "{colors.disturbed-amber}"
    textColor: "{colors.disturbed-amber-ink}"
    rounded: "{rounded.chip}"
    padding: "2px 8px"
  status-chip-quiet:
    textColor: "{colors.muted-slate}"
    rounded: "{rounded.chip}"
    padding: "2px 8px"
  status-chip-closed:
    backgroundColor: "{colors.closed-grey}"
    textColor: "{colors.closed-grey-ink}"
    rounded: "{rounded.chip}"
    padding: "2px 8px"
  line-bullet-sm:
    rounded: "{rounded.full}"
    size: "18px"
  line-bullet-md:
    rounded: "{rounded.full}"
    size: "26px"
  line-bullet-lg:
    rounded: "{rounded.full}"
    size: "36px"
  board-row-disturbed:
    backgroundColor: "{colors.disturbed-amber-soft}"
    padding: "8px 12px"
  code-block:
    backgroundColor: "{colors.code-night}"
    textColor: "{colors.code-night-ink}"
    rounded: "{rounded.md}"
    padding: "16px"
---

# Design System: GateCHA

## Overview

**Creative North Star: "Info trafic"**

The admin is a transit network's status board. Every API key is a numbered line of the network, drawn as a round metro bullet in one of eight line colours, and the dashboard opens on the board that says, per line, Normal service, No traffic today, Disturbed or Closed. A navy signage band carries the navigation in both themes; below it the ground is porcelain by day and night-navy by night, following the system theme. Light and dark are both designed: the dark theme redefines the same semantic tokens, so no component branches on the theme.

Density is Operate density: tables on strict columns, every count in Overpass Mono tabular figures, structure drawn by 1px hairlines rather than drop-shadowed cards. Colour is spent on two things only: the identity of a line, and state. Amber means disturbed, red means failures, green means normal service, grey means closed, and every one of those colours is backed by words. The type is Overpass, a face derived from highway signage, self-hosted with its mono sibling; nothing loads from a font CDN.

The previous system (slate and brand blue, Inter, large rounded cards) is retired and is not a reference.

**Key Characteristics:**
- One number per key everywhere: the line number is the key's database id, shown in the bullet on the board, the list, the detail page and the chart legend.
- Eight line colours cycle; the number, never the colour, is the identity.
- Navy signage band on every screen, including the signed-out login (mark only).
- Hairline grid: panels are bordered, KPI strips are divided by 1px gaps over the rule colour.
- Tabular mono figures for every count, formatted en-US; every date is UTC.
- Focus dims the rest: hovering a board row or pinning a legend entry steps every other line back.

## Colors

A cool navy-and-porcelain signage palette, with eight metro inks for line identity and four state colours held strictly for state.

### Primary
- **Signage Navy** (`signage-navy`): the band behind the navigation and the primary action fill in the light theme. In dark the band stays this navy while the action lifts to a brighter blue (#3366E0) so it holds contrast on the night panel.
- **Signage Navy Deep** (`signage-navy-deep`): primary action hover.
- **Band Ink / Band Muted / Band Rule** (`band-ink`, `band-muted`, `band-rule`): the active tab and mark, inactive tabs and the version, and the band's bottom edge and button outline.

### Secondary
- **Route Blue** (`route-blue`): links, focus rings, the default single-key chart stroke. Hover deepens to `route-blue-deep`. In dark, links and focus become #8AB0FF.

### Tertiary: line colours
- **Line 1 to 8** (`line-1` ... `line-8`): yellow, teal, violet, orange, sky, green, magenta, grey. Bullet fills. The number on the bullet is navy, except lines 3 and 7, which carry white. Each line also has a `-stroke` value tuned to hold 3:1 on the theme's panel for chart lines and sparks; dark redefines the strokes brighter (see sidecar). A 1px inset ring (`rgb(20 33 61 / 0.18)`, dark `rgb(231 235 243 / 0.28)`) keeps pale fills legible on white.

### State
- **Normal Green** (`normal-green`, soft `normal-green-soft`, edge `normal-green-edge`): normal service, verified counts, success rates.
- **Disturbed Amber** (`disturbed-amber`, ink `disturbed-amber-ink`, text `disturbed-amber-text`, soft `disturbed-amber-soft`, edge `disturbed-amber-edge`): the disturbed chip fill, disturbed board rows, and the text selection highlight.
- **Failure Red** (`failure-red`, soft, edge): failed counts, failure bars, error alerts. Destructive buttons use the same red with `danger-red-deep` on hover; dark sets danger to #C92A3E and failure text to #FF6B7A.
- **Closed Grey** (`closed-grey`, ink `closed-grey-ink`): the closed chip, for disabled keys.

### Neutral
- **Porcelain Ground** (`porcelain-ground`): the page. Night ground in dark: #0B1221.
- **Panel White** (`panel-white`): panels, inputs, the footer. Dark: #111A2E.
- **Sunken Grey** (`sunken-grey`): row hover, bar tracks, the cancel button, quiet tags. Dark: #0E1627.
- **Ink Navy / Ink Secondary / Muted Slate** (`ink-navy`, `ink-secondary`, `muted-slate`): primary text, labels and secondary text, hints and zero counts. Muted is also the placeholder colour.
- **Hairline / Soft / Strong** (`hairline`, `hairline-soft`, `hairline-strong`): panel borders and KPI dividers, row separators and chart gridlines, input borders and the off switch track.
- **Code Night** (`code-night`, ink `code-night-ink`): code snippets stay dark in both themes.

### Named Rules
**The State-Only Rule.** Green, amber, red and grey mean normal, disturbed, failed and closed, and nothing else. They never decorate.

**The Number Is the Line Rule.** Colours cycle every eight lines (`((id - 1) % 8) + 1`). Never let a colour alone identify a key; the bullet always carries its number.

**The Same Names Rule.** Dark mode redefines the same custom properties under `prefers-color-scheme: dark`. Components use semantic tokens only and never write a theme branch.

## Typography

**Display Font:** Overpass (with ui-sans-serif, system-ui fallback), self-hosted variable, Latin subset
**Body Font:** Overpass
**Label/Mono Font:** Overpass Mono (with ui-monospace fallback), self-hosted variable

**Character:** Signage sans for words, a monospaced sibling for every number, so figures line up on strict columns and read like a departures board.

### Hierarchy
- **Display** (800, 32px, 36px): the Dashboard page title only.
- **Headline** (800, 24px, 32px, tight): titles of every other page, including the key's name on its detail page.
- **Title** (800, 17px): panel and dialog headings. A 13px medium muted aside may follow on the same line ("Today (UTC)", "Line 5 selected").
- **Body** (400, 14px): copy, table cells, form labels (500, ink-secondary).
- **Caption** (400, 12px): domains under a key name, hints under KPI figures, panel subtitles.
- **Label** (700, 11px, 0.09em, uppercase): column headers of every table (board, keys list, MCP tokens) and labels above KPI and calibration figures. Scope tags beside a panel title use the same style at 0.08em inside a 3px hairline box.
- **Figure** (Overpass Mono 600, 32px, 40px, tabular): KPI numbers.
- **Data** (Overpass Mono, 14px, tabular): counts in tables, key ids, the version; 11px for chart axes and country codes.

### Named Rules
**The Tabular Figures Rule.** Every count is set in Overpass Mono with tabular figures and formatted with `toLocaleString('en-US')`. No bare `toLocaleString()`, no proportional digits in a column.

**The UTC Rule.** "Today" is the current UTC day and every chart date is formatted in UTC (`en-US`, short month and day). Say so in the UI where it matters ("Today (UTC)").

## Layout

Content sits in a centred 1280px column with 16px side padding, 24px from 640px and 32px from 1024px, and 32px of vertical page padding. Panels stack with 24px gaps. The dashboard places a page header (title and description left, "Create key" right), then a KPI strip of four figures (two columns below 1024px), then a three-column grid: the network board and the 30-day chart take two thirds, traffic by country and the HIS summary take the right third. Below 1024px everything stacks.

Panel headers take 12px vertical and 20px horizontal padding and are separated from the body by a hairline; detail and settings cards use a 24px inset. Board rows have 8px vertical padding, list rows 12px. On phones the status column folds under the key name, low-priority columns drop away (spark below 768px, counts below 640px), and the band's tabs move to a second row with 44px touch targets.

## Elevation & Depth

The system is flat. Depth comes from tone (ground, panel, sunken) and 1px hairlines, never from drop shadows on cards or panels. The only structural non-flat devices are the line bullet's 1px inset ring and the native top-layer backdrop of the confirm dialog (`--color-backdrop`, with a light blur). The dialog itself is drawn with a `hairline-strong` border, not a shadow; tooltips are flat too.

### Named Rules
**The Hairline Rule.** Structure is drawn with 1px rules: panel borders in `hairline`, row separators in `hairline-soft`, KPI strips as a grid with a 1px gap over the `hairline` fill. No shadow cards.

## Shapes

Small, firm corners. Panels, inputs, every button and code blocks use 6px; band controls and small tags use 4px; status chips and country code tags use 3px, like enamel signs. The only round forms are line bullets, the mark, the status dot (7px), the switch track and bar tracks. Closed lines are drawn as dashed strokes (3-3 on sparks, 4-4 on the chart) in `hairline-strong`.

## Components

### Buttons
- **Shape:** 6px corners and 14px semibold text on every button.
- **Primary:** signage navy fill, white text, 8px 16px. The page-header "Create key" is at least 40px tall, with a 14px plus glyph drawn as an SVG.
- **Hover / Focus:** hover deepens the fill; focus-visible shows a 2px route-blue ring (offset 2px on the header action). Disabled drops to 50% opacity.
- **Secondary:** panel fill, `hairline-strong` border, ink text, hover to sunken.
- **Danger:** failure red fill, white text, deeper red on hover. Destructive actions always pass through the confirm dialog.

### Chips
- **Status chip:** 12px bold, 3px corners, 2px 8px. Normal service is green text with a 7px dot and no fill; Disturbed is a solid amber fill with dark ink, a dot, and the failure count in a white mono inset ("Disturbed — 31 failed today"); No traffic today is muted text; Closed is a grey fill. The words are always present.
- **Compact:** in the keys table the disturbed chip keeps the count but drops "failed today". Disturbed rows sit on `warn-soft` in the keys table as on the board.

### Cards / Containers
- **Corner Style:** 6px.
- **Background:** panel on ground.
- **Shadow Strategy:** none (see Elevation & Depth).
- **Border:** 1px `hairline`; header separated by a hairline.
- **Internal Padding:** 20px on dashboard panels, 24px on detail and settings cards.

### Inputs / Fields
- **Style:** panel fill, 1px `hairline-strong` border, 6px corners, 8px 12px, muted placeholder.
- **Focus:** 2px route-blue ring, border shifting to the focus colour. Checkboxes take the action colour as accent.
- **Toggle switch:** a single 44x24 SVG, action fill when on, `hairline-strong` when off, white 16px knob.

### Navigation
- **Signage band:** sticky, 56px, navy in both themes. The mark is a white disc with a navy "G" beside the extrabold word mark. Tabs are 14px semibold `band-muted`; the active tab turns white over a 3px white bottom bar. The version sits in mono on the right beside an outlined Logout button. Signed out, the same band carries only the mark.
- **Footer:** panel fill with a top hairline, muted 14px links that turn route blue on hover.

### Line Bullet
A round metro bullet carrying the key's id in extrabold tabular figures: 18px in legends and inline references, 26px in tables, 36px beside the detail title. Ids of three digits or more shrink the type, never the circle.

### Network Status Board
The dashboard's signature table: Line, Key (name and first domain, "+n more"), 7-day spark (88x20, line stroke), today's challenges and failures, status. Disturbed lines sort first, closed last; within a status, by name. Disturbed rows sit on `disturbed-amber-soft` with `disturbed-amber-edge` borders above and below. Failures today above zero turn bold red. Hovering or focusing a row selects its line; only while a line is selected, every other row dims to 45% opacity.

### Lines Chart
Thirty days of challenges, one polyline per key in its line stroke, zero-filled so quiet days are real zeros. The legend is a row of small bullets that pin a line (`aria-pressed`). The active line, hovered on the board or pinned, draws on top at 2.75px; others drop to 22% stroke opacity, and inactive legend entries to 70%. On load the first disturbed line is pinned until the admin picks or clears one. Disabled keys draw dashed in `hairline-strong`.

### Key Stats Chart
The detail page's Chart.js chart reads its colours from the CSS tokens at render time (the key's line stroke for challenges, `normal-green` for verified, `failure-red` for failed, muted ticks, soft-hairline grid) and redraws when the system theme changes. Lines are straight (no tension), pointless until hover; ticks are Overpass Mono 11px, the legend Overpass 600 with circle markers.

### Traffic by Country
Rows with a 3px-cornered ISO code tag (mono, no flag emoji), the country name, the total in mono and a failure share that turns bold red at 50% or more, over a 6px bar track split green and red.

## Do's and Don'ts

### Do:
- **Do** set every count in Overpass Mono tabular figures and format it with `toLocaleString('en-US')`.
- **Do** compute "today" and format chart dates in UTC, and label it ("Today (UTC)").
- **Do** show a key as its line bullet with its number wherever the key appears.
- **Do** back every status colour with its words: Normal service, No traffic today, Disturbed, Closed.
- **Do** put disturbed board rows on `disturbed-amber-soft` with `disturbed-amber-edge` borders and sort them first.
- **Do** dim non-selected lines only while a line is selected (hover or pin); at rest every row is full strength.
- **Do** pin the first disturbed line on the dashboard chart by default.
- **Do** read Chart.js colours from the CSS tokens at runtime and redraw on a theme change.
- **Do** keep the signage band navy in both themes and redefine only semantic tokens for dark.

### Don't:
- **Don't** use green, amber, red or grey for decoration; they mean state.
- **Don't** let a line colour stand alone as an identifier; colours repeat every eight lines.
- **Don't** add drop shadows to panels or cards; draw structure with 1px hairlines.
- **Don't** load fonts or assets from a CDN; Overpass and Overpass Mono ship with the app.
- **Don't** use flag emoji for countries; use the ISO code tag.
- **Don't** bring back the retired slate/brand-blue, Inter, large-rounded-card look.
