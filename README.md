# mddeck

Terminal-native Markdown slide decks, written in Go.

mddeck renders presentations directly in your terminal using character cells and ANSI styling. The source file is plain Markdown with minimal slide semantics — it remains readable as a normal document.

## Install

```bash
go install github.com/miskun/mddeck/cmd/mddeck@latest
```

Or build from source:

```bash
git clone https://github.com/miskun/mddeck.git
cd mddeck
go build -o mddeck ./cmd/mddeck/
```

## Quick Start

```bash
# Audience mode (default)
mddeck slides.mddeck

# Presenter mode with timer, notes, and next-slide preview
mddeck --present slides.mddeck

# Watch for file changes and reload
mddeck --watch slides.mddeck
```

## CLI

```
mddeck [flags] <file>
```

Both `.mddeck` and `.md` files are accepted — the parser operates on content, not file extension.

### Flags

| Flag | Description |
|------|-------------|
| `--present`, `-p` | Start in presenter mode |
| `--theme <name>` | Override theme (`default`, `dark`, `light`) |
| `--safe-ansi` | Force safe ANSI mode (strip non-SGR sequences) |
| `--unsafe-ansi` | Disable safe ANSI mode |
| `--start <n>` | Start at slide number (1-based) |
| `--watch` | Reload on file change |
| `--version` | Show version |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime error |

---

## File Format Specification

### Encoding & Line Endings

- Encoding: **UTF-8**
- Line endings: **LF** preferred, CRLF accepted and normalized automatically

### Slide Model

A deck is a sequence of slides. Each slide contains:

- **Slide body** — a subset of Markdown
- **Optional frontmatter** — YAML metadata
- **Optional speaker notes** — hidden text for the presenter

---

## Slide Boundaries

A slide break is a line containing exactly `---` with **at least one blank line before** and **at least one blank line after**.

This preserves compatibility with standard Markdown horizontal rules.

**Valid slide break:**

```markdown
Some text.

---

# Next slide
```

**Not a slide break** (no blank lines):

```markdown
Some text.
---
# Same slide (this --- renders as a horizontal rule)
```

### Header-Based Splitting

When a file contains **no `---` slide breaks**, mddeck automatically splits on headers instead (similar to [patat](https://github.com/jaspervdj/patat)):

- The deepest heading level used becomes the **split level**
- Each heading at or above that level starts a new slide
- Headings below the split level stay within the current slide

For example, if the deepest heading in the file is `##`, then every `#` and `##` starts a new slide, while `###` and below remain part of the current slide.

### Frontmatter as Slide Boundary

A per-slide YAML frontmatter block (`---` / YAML / `---`) also starts a new slide, even in header-split mode. This lets you mix header-based splitting with layout-specific slides:

```markdown
## Regular Slide

Content split by headers.

---
layout: title-cols-2
ratio: "50/50"
---

## Left Column

Left content.

## Right Column

Right content.

## Back to Normal

This slide splits on the header again.
```

The parser automatically absorbs the correct number of headers based on region count. For example, `title-body` absorbs 2, `title-cols-2` and `title-rows-2` absorb 3, `title-grid-4` absorbs 5. Single-region layouts (`title`, `section`, `blank`) absorb 1 header. Custom layouts compute regions as cols × rows.

### Disabling Auto-Split (`autosplit: false`)

To group multiple headings onto a single slide, use `autosplit: false` in per-slide frontmatter. All content is absorbed until the next frontmatter block:

```markdown
---
autosplit: false
---

## All Heading Levels

# Heading 1
## Heading 2
### Heading 3

All on one slide.

---
autosplit: true
---

## Next Slide

Normal splitting resumes here.
```

The `autosplit: true` block acts as a resume marker — it produces no visible slide itself.

---

## Frontmatter

### Deck Frontmatter

If the file begins with `---`, the first YAML block is interpreted as **deck-level metadata**.

```yaml
---
title: "My Talk"
theme: "default"
wrap: true
tabSize: 2
slideWidth: 80
---
```

#### Deck Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `title` | string | `""` | Deck title |
| `theme` | string | `"default"` | Color theme |
| `wrap` | bool | `true` | Enable paragraph wrapping |
| `tabSize` | int | `2` | Tab expansion width |
| `slideWidth` | int | `80` | Slide stage width in characters (see Slide Dimensions) |
| `slideHeight` | int | `-1` (auto) | Slide stage height in characters (see Slide Dimensions) |
| `safeAnsi` | bool | `true` | Strip non-SGR ANSI sequences |
| `incrementalLists` | bool | `true` | Reveal list items one at a time |
| `aspect` | string | `"16:9"` | Target aspect ratio (e.g. `"16:9"`, `"4:3"`) |
| `padding` | object | all `1` | Global padding for all layouts (see Padding) |
| `footer` | object | `{}` | Footer bar configuration (see below) |
| `layouts` | map | `{}` | Custom layout definitions |

Unknown fields are silently ignored.

#### Slide Dimensions

Two parameters — `slideWidth` and `slideHeight` — control the content stage size. The stage is always centered in the terminal window. The footer sits outside the stage at full terminal width.

Each accepts three kinds of values:

| Value | Meaning |
|-------|---------|
| `> 0` | Explicit size in characters |
| `0` | Fill the terminal (no padding on that axis) |
| `-1` | Auto-calculate from the other dimension + `aspect` ratio |

Default: `slideWidth: 80`, `slideHeight: -1` (auto). This creates an 80-character-wide stage whose height is derived from the 16:9 aspect ratio.

Common configurations:

```yaml
slideWidth: 80                  # 80 chars wide, height from aspect (default)
slideWidth: 0                    # fill terminal width, height from aspect
slideWidth: 100
slideHeight: 30                  # explicit 100×30, aspect ignored
slideWidth: -1
slideHeight: 25                  # height 25, width from aspect
slideWidth: 0
slideHeight: 0                   # fill entire terminal, no padding
```

When both are `-1` (auto), the slide fills the terminal constrained by the aspect ratio.

#### Footer

The footer bar spans the bottom row of each slide. Configure it with up to three sections:

```yaml
---
footer:
  left: "Company Name"
  center: "Confidential"
  right: "Custom text"
---
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `left` | string | `""` | Left-aligned text |
| `center` | string | `""` | Centered text |
| `right` | string | slide counter | Right-aligned text (defaults to `N / M`) |

### Slide Frontmatter

Individual slides may begin with YAML frontmatter (after a slide break).

```yaml
---
layout: title-cols-2
ratio: "60/40"
align: top
---
```

#### Slide Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `layout` | enum | `"auto"` | Layout mode |
| `ratio` | string | `""` | Column ratio for `title-cols-2` (e.g. `"60/40"`) |
| `align` | enum | `"top"` | Vertical alignment |
| `title` | string | `""` | Slide title |
| `class` | string | `""` | Style class |
| `autosplit` | bool | `true` | Enable header-based splitting within this slide |
| `incrementalLists` | bool | inherited | Override deck-level `incrementalLists` for this slide |

**Layout values:** `auto`, `title`, `section`, `title-body`, `title-cols-2`, `title-rows-2`, `title-grid-4`, `blank`

**Align values:** `top`, `middle`, `bottom`

Unknown fields are silently ignored.

---

## Speaker Notes

Speaker notes begin after a line containing exactly `???`.

```markdown
# Architecture

- Access boundary
- Authorization-aware AI

???
Mention Datadog comparison here.
```

Rules:

- `???` must be on its own line
- Only the first `???` per slide is recognized
- Everything after `???` belongs to notes
- Notes are hidden in audience mode, shown in presenter mode

---

## Progressive Reveal

Slide content can be revealed incrementally with each keypress.

### Pause Markers

Insert a `. . .` line (three dots separated by spaces) to split a slide into reveal steps:

```markdown
# Architecture

First point is visible immediately.

. . .

Second point appears on the next keypress.

. . .

Third point appears after another keypress.
```

Blocks before the first `. . .` are visible immediately. Each `. . .` advances the step counter so subsequent blocks require additional keypresses to appear. The slide counter shows `[step/total]` during stepped slides.

### Incremental Lists

By default, list items are revealed one at a time — each item requires a keypress. This applies to unordered, ordered, and task lists.

```markdown
- First item visible immediately
- Second item on next keypress
- Third item on another keypress
```

Disable this at the deck level:

```yaml
---
incrementalLists: false
---
```

Or override per slide:

```yaml
---
incrementalLists: false
---
```

When combined with `. . .` markers, incremental list items count as additional steps after the pause marker.

---

## Markdown Support

### Block Elements

| Element | Syntax |
|---------|--------|
| Headings | `#`, `##`, `###` |
| Paragraphs | Plain text separated by blank lines |
| Unordered lists | `- item` or `* item` (supports nesting via indentation) |
| Ordered lists | `1. item`, `2. item` (supports nesting via indentation) |
| Task lists | `- [ ] unchecked`, `- [x] checked` |
| Blockquotes | `> text` |
| Alerts/Callouts | `> [!NOTE]`, `> [!TIP]`, `> [!WARNING]`, `> [!IMPORTANT]`, `> [!CAUTION]` |
| Tables | Pipe-delimited `\| col \| col \|` with header separator |
| Fenced code blocks | ` ``` ` with optional language tag |
| Horizontal rules | `---`, `***`, `___` (when not a slide break) |

### Inline Elements

| Element | Syntax | Rendering |
|---------|--------|-----------|
| Bold | `**text**` | Bold weight |
| Italic | `*text*` | Italic style |
| Strikethrough | `~~text~~` | Struck-through text |
| Inline code | `` `code` `` | Colored text |
| Links | `[text](url)` | Underlined accent text |
| Hard line break | Trailing `\` or two spaces | Forces a new line |

### Nested Lists

Indent list items by 2 spaces per level for nesting:

```markdown
- Top level
  - Second level
    - Third level
```

Unordered lists use distinct bullets per depth (•, ◦, ▪). Ordered lists maintain per-depth numbering.

Trailing `\` works as a hard line break within list items — the continuation text appears on a new line, indented under the bullet:

```markdown
* **Bold heading** \
  Continuation text on the next line.
```

### Tables

Pipe-delimited tables render with Unicode box-drawing characters:

```markdown
| Feature | Status |
|---------|--------|
| Tables  | Done   |
```

The separator row (`|---|---|`) is required between the header and body rows. Column widths auto-size to content and shrink proportionally when the terminal is narrow. Cells that overflow are truncated with an ellipsis (`…`).

#### Mid-Table Separators

Additional separator rows within the table body render as horizontal dividers:

```markdown
| Name  | Team    |
|-------|---------|
| Alice | Alpha   |
|-------|---------|
| Bob   | Beta    |
```

#### Headerless Tables

To render a table without a bold header row, start with a separator line:

```markdown
|-------|---------|
| Alice | Alpha   |
| Bob   | Beta    |
```

All rows render as plain text with no header emphasis.

### Alerts / Callouts

GitHub-flavored alert syntax inside blockquotes:

```markdown
> [!NOTE]
> Additional context the reader should know.

> [!WARNING]
> Something that could cause problems.
```

Supported types: `NOTE`, `TIP`, `IMPORTANT`, `WARNING`, `CAUTION`. Each type renders with a distinct icon and color.

---

## Art Blocks

mddeck treats certain fenced code blocks as first-class art:

| Language tag | Type | Description |
|-------------|------|-------------|
| `ansi` | ANSI art | Colored text with escape sequences |
| `ascii` | ASCII art | Plain monospace art |
| `braille` | Braille art | Unicode braille patterns |
| Any other | Code block | Standard syntax-highlighted code |

### Example

````markdown
```ansi
\033[32mPASS\033[0m  All tests passed
\033[31mFAIL\033[0m  Connection timeout
```
````

### Escape Parsing

In `ansi` blocks, these literal sequences are converted to actual escapes:

- `\033` → ESC
- `\e` → ESC
- `\x1b` / `\x1B` → ESC

### Wrapping

Art blocks (`ansi`, `ascii`, `braille`) are **nowrap** by default. Overflow is cropped.

---

## ANSI Safety Model

**Default: `safeAnsi = true`**

When safe mode is enabled, only **ANSI SGR sequences** (colors, bold, underline, reset) are allowed. All other sequences are stripped:

- Cursor movement
- Screen clearing
- OSC sequences (clipboard, hyperlinks)
- Alternate screen switching

When `safeAnsi = false`, additional sequences pass through, but the runtime never executes commands, accesses the filesystem, or performs network operations.

---

## Layout System

All layouts — built-in and custom — use the same grid engine. Built-in layouts are simply pre-defined grid configurations that save you from writing them out each time.

### Built-in Layouts

All built-in layouts use the same defaults: padding of 1 on all sides, gutterX of 2 (horizontal gap between columns), and gutterY of 1 (vertical gap between rows). No special-case overrides — they behave identically to custom layouts.

| Layout | Structure | Description |
|--------|-----------|-------------|
| `auto` | (detected) | Automatically detected from content |
| `title` | 1 region, centered | Centered title slide (h1 + optional subtitle) |
| `section` | 1 region, centered | Section header (centered, like title) |
| `title-body` | title row + 1 body | Title row + single content area |
| `title-cols-2` | title row + 2 cols | Title row + two columns |
| `title-rows-2` | title row + 2 rows | Title row + two stacked rows |
| `title-grid-4` | title row + 2×2 | Title row + four quadrants |
| `blank` | 1 region, top-aligned | No title, single region, full space |

### Auto-Detection Heuristics

When `layout: auto` (the default), the layout is chosen as follows:

| Condition | Layout |
|-----------|--------|
| No blocks | `blank` |
| H1 present + minimal text (≤3 blocks, headings + ≤1 paragraph) | `title` |
| No H1 + minimal text (≤3 blocks, headings + ≤1 paragraph) | `section` |
| Single code/art block (≤2 blocks) | `blank` |
| Two major blocks (top-level headings) | `title-cols-2` |
| Otherwise | `title-body` |

### Column Ratio

The `title-cols-2` layout defaults to a 50/50 column split. Override per slide:

```yaml
---
layout: title-cols-2
ratio: "60/40"
---
```

### Aspect Ratio

The default aspect ratio is `16:9`. Override it in deck frontmatter:

```yaml
---
aspect: "4:3"
---
```

The aspect ratio controls centering of the content stage within the terminal. When the terminal is wider than the target ratio, horizontal centering padding (pillarbox) is added. When taller, vertical centering padding (letterbox) is added. Terminal character cells are approximately 1:2 (width:height), so the computation accounts for this cell ratio.

Layout padding (configurable via the `padding` field or per-layout `padX`/`padY`/`padTop`/`padBottom`/`padLeft`/`padRight`) is applied inside the content stage, independent of aspect ratio centering.

### Stage Background (`PadBg`)

By default the area outside the slide content (centering margins from aspect ratio) uses the terminal's own background — the slide blends seamlessly into the terminal.

To visually distinguish the slide canvas from the surrounding terminal, set the `PadBg` field in a theme. This paints the area **outside** the content stage with a background color, making the slide boundary visible.

Built-in themes leave `PadBg` empty (transparent). To enable it, override the theme in code or define a custom theme with `PadBg` set to an ANSI background escape (e.g. `ansi.BgRGB(28, 28, 34)` for a subtle dark border).

### Custom Layouts

Define custom grid layouts in deck frontmatter under `layouts`. Custom layouts use the exact same parameters as built-in layouts — columns, rows, gutterX, gutterY, padding.

```yaml
---
layouts:
  hero:
    columns: [40, 60]
    gutterX: 4
  dashboard:
    columns: [33, 34, 33]
    rows: [60, 40]
---
```

Reference a custom layout per slide:

```yaml
---
layout: hero
---
```

The parser automatically absorbs the correct number of headings per slide based on region count (cols × rows).

#### Layout Fields

These fields apply to both custom layouts and built-in overrides:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `columns` | `[]int` | `[100]` | Column widths as percentages |
| `rows` | `[]int` | `[100]` | Row heights as percentages |
| `gutterX` | int | `2` | Horizontal gap between columns in characters |
| `gutterY` | int | `1` | Vertical gap between rows in lines |
| `padX` | int | `2` | Horizontal padding (sets both left and right) |
| `padY` | int | `1` | Vertical padding (sets both top and bottom) |
| `padTop` | int | `1` | Top padding |
| `padBottom` | int | `1` | Bottom padding |
| `padLeft` | int | `2` | Left padding |
| `padRight` | int | `2` | Right padding |
| `align` | enum | `"top"` | Content alignment within cells |

#### Padding

Padding controls the inset between the content stage and the layout regions. It is resolved in priority order (lowest to highest):

1. **Hard-coded default**: top=1, bottom=1, left=2, right=2
2. **Deck-level `padding:`** in deck frontmatter (global override for all layouts)
3. **Layout-level `padX`/`padY`** convenience fields (set both sides at once)
4. **Layout-level `padTop`/`padBottom`/`padLeft`/`padRight`** (most specific)

Deck-level global padding:
```yaml
---
padding:
  top: 2
  bottom: 2
  left: 3
  right: 3
---
```

Per-layout override (in deck frontmatter `layouts:` section):
```yaml
layouts:
  blank:
    padTop: 0
    padBottom: 0
  title:
    padX: 2
    padTop: 0
```

#### Grid Region Order

Regions are created in **row-major order** (left-to-right, top-to-bottom). Content blocks are distributed across regions by splitting at H1/H2 headings (major blocks) and assigning them round-robin.

Example: a `columns: [50, 50]` + `rows: [50, 50]` layout creates 4 regions. With 4 major blocks (## A, ## B, ## C, ## D), they map to:

| Region | Position | Content |
|--------|----------|---------|
| 0 | top-left | ## A |
| 1 | top-right | ## B |
| 2 | bottom-left | ## C |
| 3 | bottom-right | ## D |

#### Overriding Built-in Layouts

Override any built-in layout's parameters by defining it in your deck frontmatter. Unspecified fields keep their defaults:

```yaml
---
layouts:
  title-body:
    padX: 10
    padY: 3
  title-cols-2:
    gutterX: 4
---
```

---

## Rendering Rules

### Styling

All styling uses ANSI SGR sequences:

- **Title headings** (on `title`/`section` slides) — bold + title color (TitleStyle)
- **Slide title headings** (title row on grid layouts) — bold + slide title color (SlideTitleStyle)
- **Content headings** (H1-H3 in body regions) — bold + accent color
- **Content headings** (H4-H6) — bold + body foreground
- **Inline code** — colored text
- **Blockquotes** — muted color + `│` indicator
- **Lists** — accent-colored bullets (`•`) and numbers

### Wrapping

- Paragraphs wrap when `wrap = true` (default)
- Lists use hanging indentation
- Art/code blocks never wrap

### Whitespace

- Tabs expanded using `tabSize` (default: 2)
- Fenced blocks preserve whitespace exactly

### Cropping

When content exceeds its region:

- Horizontal overflow → truncated with `…`
- Vertical overflow → truncated with `↓` indicator

No scaling in v1.

---

## Keyboard Controls

### Navigation

| Key(s) | Action |
|--------|--------|
| Space, Enter, →, PgDn, `n` | Next slide |
| Backspace, ←, PgUp, `p` | Previous slide |
| Home | First slide |
| End | Last slide |

### Modes

| Key | Action |
|-----|--------|
| `t` | Toggle presenter mode |
| `?` | Toggle help overlay |
| `q`, Ctrl+C | Quit |

### Presenter Mode

Displays:

- Current slide
- Next slide preview
- Speaker notes
- Elapsed timer
- Slide index (e.g. `3 / 18`)

---

## Themes

Three built-in themes:

| Theme | Description |
|-------|-------------|
| `default` | Cyan accent on default background |
| `dark` | Steel blue accent for dark terminals |
| `light` | Blue accent for light terminals |

Themes define: base foreground/background, accent color, muted color, heading styles, title styles, and heading margins.

Override via CLI (`--theme dark`) or deck frontmatter (`theme: "dark"`).

### Title Style Tokens

Slide titles use dedicated style tokens that are visually distinct from content headings:

| Token | Purpose | Fallback |
|-------|---------|----------|
| `TitleStyle` | Main heading on `title` and `section` centered slides | `H1Style` |
| `SlideTitleStyle` | Heading in the title row of grid layouts (`title-body`, `title-cols-2`, etc.) | `H2Style` |

Content headings (H1-H3) use a uniform accent color (`H1Style`/`H2Style`/`H3Style`), while H4-H6 use bold body foreground.

### Heading Margin Tokens

Each heading level has a configurable margin-bottom (number of blank lines after the heading). All default to 1.

| Token | Purpose |
|-------|---------|
| `TitleMargin` | After main title on centered slides |
| `SlideTitleMargin` | After title-row heading on grid layouts |
| `H1Margin` – `H6Margin` | After content headings at each level |

---

## Error Handling

| Condition | Behavior |
|-----------|----------|
| YAML parse error | Exit non-zero with message |
| Unknown YAML keys | Silently ignored |
| Invalid `ratio` | Fallback to default (50/50) |
| Invalid UTF-8 | Replaced with `�` |

---

## Example Deck

```markdown
---
title: "My Talk"
footer:
  left: "Acme Corp"
  center: "Internal"
---

# Hello, World!

A terminal-native presentation.

???
Opening remarks go here.

## Key Features

- **Fast** — renders instantly
- **Portable** — runs in any terminal
- **Simple** — just Markdown

---
layout: title-cols-2
ratio: "50/50"
---

## Title

### Left Side

Content for the left column.

### Right Side

Content for the right column.

# Thank You!
```

---

## Not Supported (v1)

- HTML rendering
- Inline images (Kitty, Sixel, etc.)
- Embedded scripting
- Component DSLs
- Region markers (e.g. `:::left`)
- Fenced block options (e.g. `{fit=width}`)
- Scaling logic
- Animation

The `.mddeck` file must remain readable and meaningful as a normal Markdown document.

## License

MIT
