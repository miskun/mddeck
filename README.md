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
layout: two-col
ratio: "50/50"
---

## Left Column

Left content.

## Right Column

Right content.

## Back to Normal

This slide splits on the header again.
```

When a frontmatter block specifies `layout: two-col` or `layout: split`, two subsequent headers are absorbed into that slide (for the two regions). Other layouts absorb one header.

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
---
```

#### Deck Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `title` | string | `""` | Deck title |
| `theme` | string | `"default"` | Color theme |
| `wrap` | bool | `true` | Enable paragraph wrapping |
| `tabSize` | int | `2` | Tab expansion width |
| `maxWidth` | int | `0` (auto) | Maximum viewport width |
| `maxHeight` | int | `0` (auto) | Maximum viewport height |
| `safeAnsi` | bool | `true` | Strip non-SGR ANSI sequences |
| `aspect` | string | `""` | Target aspect ratio (e.g. `"16:9"`, `"4:3"`) |
| `layouts` | map | `{}` | Custom layout definitions |

Unknown fields are silently ignored.

### Slide Frontmatter

Individual slides may begin with YAML frontmatter (after a slide break).

```yaml
---
layout: two-col
ratio: "60/40"
align: top
---
```

#### Slide Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `layout` | enum | `"auto"` | Layout mode |
| `ratio` | string | `""` | Column ratio for `two-col` (e.g. `"60/40"`) |
| `align` | enum | `"top"` | Vertical alignment |
| `title` | string | `""` | Slide title |
| `class` | string | `""` | Style class |
| `autosplit` | bool | `true` | Enable header-based splitting within this slide |

**Layout values:** `auto`, `default`, `title`, `center`, `two-col`, `split`, `terminal`

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

### Tables

Pipe-delimited tables render with Unicode box-drawing characters:

```markdown
| Feature | Status |
|---------|--------|
| Tables  | Done   |
```

The separator row (`|---|---|`) is required between the header and body rows. Column widths auto-size to content and shrink proportionally when the terminal is narrow.

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

| Layout | Grid | Padding | Description |
|--------|------|---------|-------------|
| `auto` | (detected) | — | Automatically detected from content |
| `default` | 1×1 | padX: W/8, padY: 1 | Proportional horizontal padding, top-aligned |
| `title` | 1×1 | padX: 0, padY: 0 | Centered title slide (large heading) |
| `center` | 1×1 | padX: W/8, padY: 0 | Content centered vertically and horizontally |
| `two-col` | 2×1 (62/38) | padX: 0, padY: 1 | Two columns, gutter: 2 |
| `split` | 1×2 (60/40) | padX: 2, padY: 1 | Top/bottom split, gutter: 1 |
| `terminal` | 1×1 | padX: 2, padY: 1 | Minimal padding for code/art |

### Auto-Detection Heuristics

When `layout: auto` (the default), the layout is chosen as follows:

| Condition | Layout |
|-----------|--------|
| Single H1 + minimal text (≤3 blocks) | `title` |
| Single code/art block (≤2 blocks) | `terminal` |
| Two major blocks (top-level headings) | `two-col` |
| Otherwise | `default` |

### Two-Column Ratio

The `two-col` layout defaults to a 62/38 column split. Override per slide:

```yaml
---
layout: two-col
ratio: "50/50"
---
```

### Aspect Ratio

Set `aspect` in deck frontmatter to constrain the slide area to a target aspect ratio. Padding is added automatically to letterbox or pillarbox the content.

```yaml
---
aspect: "16:9"
---
```

Terminal character cells are approximately 1:2 (width:height), so a 16:9 aspect target computes the correct padding accounting for this cell ratio. Aspect padding acts as a minimum — it never shrinks a layout's own padding.

### Custom Layouts

Define custom grid layouts in deck frontmatter under `layouts`. Custom layouts use the exact same parameters as built-in layouts — columns, rows, gutter, padding.

```yaml
---
layouts:
  sidebar:
    columns: [30, 70]
    gutter: 3
  thirds:
    columns: [33, 34, 33]
  quad:
    columns: [50, 50]
    rows: [50, 50]
    gutter: 1
---
```

Reference a custom layout per slide:

```yaml
---
layout: sidebar
---
```

The parser automatically absorbs the correct number of headings per slide based on region count (cols × rows).

#### Layout Fields

These fields apply to both custom layouts and built-in overrides:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `columns` | `[]int` | `[100]` | Column widths as percentages |
| `rows` | `[]int` | `[100]` | Row heights as percentages |
| `gutter` | int | `2` | Gap between cells in characters |
| `padX` | int | proportional (W/8) | Horizontal padding |
| `padY` | int | `1` | Vertical padding |
| `align` | enum | `"top"` | Content alignment within cells |

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
  default:
    padX: 10
    padY: 3
  two-col:
    columns: [50, 50]
    gutter: 4
---
```

---

## Rendering Rules

### Styling

All styling uses ANSI SGR sequences:

- **Headings** — bold + accent color
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
| `dark` | Magenta accent for dark terminals |
| `light` | Blue accent for light terminals |

Themes define: base foreground/background, accent color, muted color, heading styles.

Override via CLI (`--theme dark`) or deck frontmatter (`theme: "dark"`).

---

## Error Handling

| Condition | Behavior |
|-----------|----------|
| YAML parse error | Exit non-zero with message |
| Unknown YAML keys | Silently ignored |
| Invalid `ratio` | Fallback to default (62/38) |
| Invalid UTF-8 | Replaced with `�` |

---

## Example Deck

```markdown
---
title: "My Talk"
theme: "default"
---

# Hello, World!

A terminal-native presentation.

???
Opening remarks go here.

---

## Key Features

- **Fast** — renders instantly
- **Portable** — runs in any terminal
- **Simple** — just Markdown

---

---
layout: two-col
ratio: "50/50"
---

## Left Side

Content for the left column.

## Right Side

Content for the right column.

---

---
layout: title
---

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
