# mddeck

Terminal-native Markdown slide decks, written in Go.

mddeck renders presentations directly in your terminal using character cells and ANSI styling. The source file is plain Markdown with minimal slide semantics — it remains readable as a normal document.

## Install

```bash
go install github.com/miska/mddeck/cmd/mddeck@latest
```

Or build from source:

```bash
git clone https://github.com/miska/mddeck.git
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

**Layout values:** `auto`, `title`, `center`, `two-col`, `split`, `terminal`

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
| Unordered lists | `- item` or `* item` |
| Ordered lists | `1. item`, `2. item` |
| Blockquotes | `> text` |
| Fenced code blocks | ` ``` ` with optional language tag |
| Horizontal rules | `---`, `***`, `___` (when not a slide break) |

### Inline Elements

| Element | Syntax | Rendering |
|---------|--------|-----------|
| Bold | `**text**` | Bold weight |
| Italic | `*text*` | Italic style |
| Inline code | `` `code` `` | Colored text |
| Links | `[text](url)` | Underlined accent text |

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

Rendering is cell-based, filling the terminal viewport.

### Layout Modes

| Layout | Description |
|--------|-------------|
| `auto` | Automatically detected from content |
| `title` | Centered title slide |
| `center` | Content with horizontal padding |
| `two-col` | Two columns side by side |
| `split` | Top/bottom split (60/40) |
| `terminal` | Full-width for code/art |

### Auto-Detection Heuristics

When `layout: auto` (the default), the layout is chosen as follows:

| Condition | Layout |
|-----------|--------|
| Single H1 + minimal text (≤3 blocks) | `title` |
| Single code/art block (≤2 blocks) | `terminal` |
| Two major blocks (top-level headings) | `two-col` |
| Otherwise | `center` |

### Two-Column Layout

- Default ratio: **62/38**
- Custom ratio via `ratio: "A/B"` (e.g. `"50/50"`, `"70/30"`)
- Gutter: 2 spaces
- Major blocks distributed alternately left/right

### Split Layout

- Top region: 60% height
- Bottom region: 40% height
- First major block → top, remaining → bottom

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
