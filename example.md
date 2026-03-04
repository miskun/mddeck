---
title: "mddeck Feature Tour"
theme: "dark"
incrementalLists: false
footer:
  left: "mddeck"
  center: "Feature Tour"
---

# mddeck

*Terminal-native slide decks, powered by Markdown*

```ascii
    ┌─────────────────────────────────────────┐
    │                                         │
    │   Write in Markdown.                    │
    │   Present in the terminal.              │
    │   No browser. No GUI. Just text.        │
    │                                         │
    └─────────────────────────────────────────┘
```

???
Welcome to the mddeck feature tour! This deck showcases every major feature.
Press Space or → to advance. Press ? for help anytime.

## 01 / TEXT STYLING

### Inline Styles

Markdown formatting renders as ANSI terminal styles:

- **Bold text** for strong emphasis
- *Italic text* for lighter emphasis
- ***Bold and italic*** combined
- ~~Strikethrough~~ for deleted or deprecated content
- `inline code` for variables, functions, and commands
- [Links](https://github.com/miskun/mddeck) render as underlined accent text

### Hard Line Breaks

A trailing backslash forces a\
hard line break right here.

Two trailing spaces also work  
for hard line breaks.

Without a break marker,
these lines merge into one paragraph.

## 02 / HEADINGS

---
autosplit: false
layout: default
---

### Heading Levels

Three heading levels with distinct styling:

# Heading Level 1

## Heading Level 2

### Heading Level 3

Each level gets its own color and weight from the active theme.

---
autosplit: true
---

## 03 / LISTS

---
layout: cols-2
ratio: "55/45"
incrementalLists: false
---

### Unordered Lists

Nested with distinct bullets per depth:

- Top level — `•`
  - Second level — `◦`
    - Third level — `▪`
- Lists support **bold**, *italic*, `code`
- And [links](https://example.com) too

Hard line breaks work in lists:

* **Feature name** \
  Description on the next line, \
  indented under the bullet.

### Ordered & Task Lists

Ordered with per-depth numbering:

1. First step
   1. Sub-step A
   2. Sub-step B
2. Second step
3. Third step

Task lists with checkboxes:

- [x] Parser implementation
- [x] ANSI renderer
- [x] Layout engine
- [ ] Syntax highlighting
- [ ] Ship v2.0

## 04 / PROGRESSIVE REVEAL

### Click-to-Reveal

Use `. . .` pause markers to reveal content step by step.

. . .

**This appeared on the first click.**

. . .

**This appeared on the second click.**

. . .

The slide counter shows `[step/total]` during reveals. Incremental lists reveal items one at a time by default — disable per-slide with `incrementalLists: false`.

## 05 / CODE BLOCKS

### Syntax Examples

```go
package main

import "fmt"

func main() {
    deck := Parse("slides.md")
    for _, slide := range deck.Slides {
        fmt.Printf("Slide %d: %s\n", slide.Index+1, slide.Meta.Title)
    }
}
```

```python
def fibonacci(n: int) -> list[int]:
    a, b = 0, 1
    result = []
    for _ in range(n):
        result.append(a)
        a, b = b, a + b
    return result
```

```bash
# Build and run
go build -o mddeck ./cmd/mddeck/
./mddeck --theme dark slides.md
./mddeck --present --watch slides.md
```

## 06 / ANSI ART

```ansi
\033[38;5;201m ██\033[38;5;207m╗      \033[38;5;213m██████\033[38;5;219m╗ \033[38;5;225m███\033[38;5;189m╗  \033[38;5;153m██\033[38;5;147m╗\033[38;5;141m███████\033[38;5;135m╗\033[0m
\033[38;5;201m ██\033[38;5;207m║      \033[38;5;213m██\033[38;5;219m╔═══\033[38;5;225m╝ \033[38;5;189m████\033[38;5;153m╗ \033[38;5;147m██\033[38;5;141m║██\033[38;5;135m╔════╝\033[0m
\033[38;5;201m ██\033[38;5;207m║      \033[38;5;213m█████\033[38;5;219m╗  \033[38;5;225m██\033[38;5;189m╔██\033[38;5;153m╗██\033[38;5;147m║\033[38;5;141m███████\033[38;5;135m╗\033[0m
\033[38;5;201m ██\033[38;5;207m║      \033[38;5;213m██\033[38;5;219m╔══╝  \033[38;5;225m██\033[38;5;189m║╚\033[38;5;153m████\033[38;5;147m║╚═\033[38;5;141m═══██\033[38;5;135m║\033[0m
\033[38;5;201m ███████\033[38;5;207m╗\033[38;5;213m██████\033[38;5;219m╗  \033[38;5;225m██\033[38;5;189m║ ╚\033[38;5;153m███\033[38;5;147m║\033[38;5;141m███████\033[38;5;135m║\033[0m
\033[38;5;201m ╚══════\033[38;5;207m╝\033[38;5;213m╚═════\033[38;5;219m╝  \033[38;5;225m╚═\033[38;5;189m╝  \033[38;5;153m╚══\033[38;5;147m╝\033[38;5;141m╚══════\033[38;5;135m╝\033[0m
```

ANSI art blocks parse escape sequences (`\033[...m`, `\e[...m`, `\x1b[...m`) and render full 256-color terminal graphics.

## 07 / ASCII & BRAILLE ART

---
layout: cols-2
incrementalLists: false
---

### ASCII Art

```ascii
    ┌───────────────────┐
    │  Terminal Input   │
    └────────┬──────────┘
             │
    ┌────────▼──────────┐
    │   Parse Blocks    │
    └────────┬──────────┘
             │
    ┌────────▼──────────┐
    │  Compute Layout   │
    └────────┬──────────┘
             │
    ┌────────▼──────────┐
    │  Render to ANSI   │
    └───────────────────┘
```

### Braille Art

```braille
⠀⠀⠀⠀⠀⠀⣀⣤⣴⣶⣶⣶⣶⣤⣄⡀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠀⠀⠀⠀
⠀⠀⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠀⠀⠀
⠀⠀⣸⣿⣿⣿⡿⠿⠛⠛⠛⠛⠿⢿⣿⣿⣿⣿⣿⡆⠀⠀
⠀⠀⣿⣿⣿⠁⠀⠀⠀⠀⠀⠀⠀⠀⠈⣿⣿⣿⣿⡇⠀⠀
⠀⠀⣿⣿⣿⠀⠀⢸⣿⡆⠀⣿⣿⠀⠀⣿⣿⣿⣿⡇⠀⠀
⠀⠀⢹⣿⣿⡀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣿⣿⣿⣿⠃⠀⠀
⠀⠀⠀⢿⣿⣷⡀⠀⠙⠿⠿⠋⠀⣠⣿⣿⣿⣿⠏⠀⠀⠀
⠀⠀⠀⠀⠻⣿⣿⣦⣄⣀⣀⣤⣾⣿⣿⣿⠿⠃⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠉⠛⠿⣿⣿⡿⠿⠛⠉⠀⠀⠀⠀⠀⠀⠀
```

Three art block types. Use `ansi`, `ascii`, or `braille` as the fence language tag.

## 08 / TABLES

### Feature Matrix

| Feature          | Status | Notes                           |
|------------------|--------|---------------------------------|
| Bold & Italic    | ✅     | Standard Markdown syntax        |
| Strikethrough    | ✅     | `~~text~~`                      |
| Code Blocks      | ✅     | With language tags               |
| Tables           | ✅     | Auto-sizing, box-drawing        |
| Art Blocks       | ✅     | ANSI, ASCII, Braille            |
| Progressive Reveal | ✅   | `. . .` pauses + incremental lists |
| Nested Lists     | ✅     | 3 depth levels                  |
| Callouts         | ✅     | 5 types with icons              |

Tables auto-size columns, render with Unicode box-drawing, and truncate overflow cells with `…` ellipsis.

---
incrementalLists: false
---

### Headerless Tables & Separators

Start with a separator row for headerless tables:

|---------|---------|
| Alice   | Alpha   |
| Bob     | Beta    |

Mid-table separators create visual groups:

| Team    | Members |
|---------|---------|
| Core    | 4       |
|---------|---------|
| GTM     | 6       |
|---------|---------|
| Labs    | 3       |

## 09 / BLOCKQUOTES & ALERTS

### Blockquotes

> "Simplicity is the ultimate sophistication."
> — Leonardo da Vinci

### Alerts & Callouts

> [!NOTE]
> Additional context the reader should know.

> [!TIP]
> Helpful advice for getting the most out of a feature.

> [!WARNING]
> Something that could cause problems if not careful.

> [!IMPORTANT]
> Critical information the reader must not miss.

> [!CAUTION]
> Danger zone — could cause data loss or other serious issues.

## 10 / HORIZONTAL RULES

Content above.

***

A horizontal rule: `***`, `___`, or `---` (when not a slide break).

___

Content below.

---
layout: cols-2
---

## 11 / TWO-COLUMN LAYOUT

Use `layout: cols-2` in slide frontmatter. Content distributes across columns by major blocks (heading + content).

```yaml
---
layout: cols-2
ratio: "60/40"
---
```

- Default ratio: 50/50
- Custom via `ratio: "70/30"`
- Each column is a separate region

### Right Column

```ascii
  ┌──────────┬──────────┐
  │          │          │
  │  Left    │  Right   │
  │  Column  │  Column  │
  │          │          │
  └──────────┴──────────┘
```

---
layout: rows-2
---

## Top Region — `rows-2`

Horizontal split: top region (60%) and bottom region (40%). Great for showing a concept above and details below.

## Bottom Region

```ascii
  ┌──────────────────────┐
  │     Top Region       │
  ├──────────────────────┤
  │   Bottom Region      │
  └──────────────────────┘
```

---
layout: sidebar
---

### Sidebar

Quick links:

- Intro
- Styling
- Lists
- Code
- Art
- Tables
- Layouts

### Main Content Area

The **sidebar** layout uses a `30/70` column split. The narrow left panel works well for navigation, table of contents, or metadata.

Content in the wider right panel has room for longer text, code blocks, and diagrams.

```ascii
  ┌─────┬────────────────┐
  │     │                │
  │ Nav │    Content     │
  │     │                │
  └─────┴────────────────┘
```

---
layout: cols-3
---

### Column A

Three equal columns for comparison layouts.

Each gets ~33% width.

### Column B

All three columns get the same treatment and styling.

Use for feature comparisons or team structures.

### Column C

Headers distribute automatically across regions.

```
layout: cols-3
```

---
layout: grid-4
---

### Top Left

Four-quadrant grid.

### Top Right

50/50 columns × 50/50 rows.

### Bottom Left

Great for dashboards or comparison matrices.

### Bottom Right

```
layout: grid-4
```

---
layout: terminal
---

## Terminal Layout

```ascii
╔═════════════════════════════════════════════════════════════════════════╗
║                                                                         ║
║   $ mddeck --present --watch slides.md                                  ║
║                                                                         ║
║   ┌─────────────────────────────────────────────────────────────────┐   ║
║   │                                                                 │   ║
║   │   Your presentation renders here                                │   ║
║   │   in full terminal glory.                                       │   ║
║   │                                                                 │   ║
║   │                                                   1 / 42        │   ║
║   └─────────────────────────────────────────────────────────────────┘   ║
║                                                                         ║
╚═════════════════════════════════════════════════════════════════════════╝
```

Full viewport width — ideal for code demos and ASCII art.

---
layout: center
---

## Centered Layout

Content is centered both vertically and horizontally.

Perfect for section dividers, quotes, or dramatic reveals.

## 12 / SLIDE DIMENSIONS

### Controlling the Stage

```yaml
---
slideWidth: 80        # chars wide (default)
slideHeight: -1       # auto from aspect ratio
aspect: "16:9"        # target aspect ratio
---
```

| Value | Meaning                                      |
|-------|----------------------------------------------|
| `> 0` | Explicit size in characters                 |
| `0`   | Fill the terminal — no padding              |
| `-1`  | Auto-calculate from other dim + aspect ratio |

The slide stage is always centered in the terminal. The footer sits outside at full terminal width.

## 13 / SPEAKER NOTES & PRESENTER MODE

### Presenter View

Press `t` to toggle presenter mode:

- **Top 55%** — Current slide
- **Bottom left** — Next slide preview
- **Bottom right** — Speaker notes
- **Timer** — Elapsed time since start

Notes are written after `???` on any slide:

```markdown
## My Slide

Content here.

???
These notes are only visible in presenter mode.
Remind audience about the demo.
```

???
This is a speaker note! You're seeing it because you're in presenter mode.
Toggle with `t`. The timer in the corner tracks elapsed time.

## 14 / CONFIGURATION REFERENCE

### Slide Frontmatter

```yaml
---
layout: cols-2          # layout mode
ratio: "60/40"          # column ratio (cols-2 only)
align: middle           # vertical alignment
autosplit: false         # disable header splitting
incrementalLists: false  # disable per-slide
---
```

### Themes

Three built-in themes:

| Theme     | Accent     | Best for               |
|-----------|------------|------------------------|
| `default` | Cyan       | Default terminal bg    |
| `dark`    | Steel blue | Dark terminal bg       |
| `light`   | Blue       | Light terminal bg      |

Set via `--theme` flag or `theme:` in deck frontmatter.

---
layout: center
---

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Space` `Enter` `→` `n` | Next slide / step |
| `Backspace` `←` `p` | Previous slide / step |
| `Home` | First slide |
| `End` | Last slide |
| `t` | Toggle presenter mode |
| `?` | Help overlay |
| `q` | Quit |

---
layout: center
---

## Thank You!

```ascii
    ┌─────────────────────────────────────┐
    │                                     │
    │   github.com/miskun/mddeck          │
    │                                     │
    │   go install github.com/miskun/     │
    │     mddeck/cmd/mddeck@latest        │
    │                                     │
    └─────────────────────────────────────┘
```

Press `q` to quit.
