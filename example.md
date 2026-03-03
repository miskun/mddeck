---
title: "Markdown in mddeck"
footer:
  left: "mddeck"
  center: "Feature Tour"
---

# Markdown in mddeck

A tour of supported Markdown features, one slide at a time.

???
This is so cool!

---
autosplit: false
---

## Headings

# Heading Level 1

## Heading Level 2

### Heading Level 3

Each heading level gets distinct styling — bold weight plus accent color.

---
autosplit: true
---

## Paragraphs & Wrapping

This is a plain paragraph. When `wrap: true` is set in frontmatter (the default), long lines are automatically wrapped to fit the terminal width. This makes it easy to write flowing prose without worrying about line length.

Short paragraphs work too.

Multiple paragraphs are separated by blank lines, just like standard Markdown.

## Bold & Italic

Use **bold** for strong emphasis and *italic* for lighter emphasis.

You can combine them: ***bold and italic*** together.

They work inside sentences — the **quick** brown *fox* jumps over the **lazy** dog.

## Inline Code

Use backticks for inline code: `fmt.Println("hello")`.

Variable names like `ctx`, function calls like `handleRequest()`, and types like `io.Reader` are all rendered in a distinct color.

## Unordered Lists

- First item
- Second item with **bold** text
- Third item with `inline code`
- Nested concepts are expressed as separate items
- Lists use accent-colored bullet characters

## Ordered Lists

1. Parse the `.mddeck` file
2. Build the slide deck model
3. Compute layout for each slide
4. Render blocks to ANSI output
5. Enter the terminal event loop

Ordered lists get accent-colored numbers.

## Blockquotes

> "Simplicity is the ultimate sophistication."
> — Leonardo da Vinci

Blockquotes are rendered with a muted vertical bar indicator and dimmed text.

> You can have multiple blockquotes on the same slide.

## Fenced Code Blocks

```go
package main

import "fmt"

func main() {
    deck := LoadDeck("slides.md")
    for _, slide := range deck.Slides {
        fmt.Println(slide.Title)
    }
}
```

Fenced blocks preserve whitespace and never wrap.

## Multiple Languages

```python
def fibonacci(n):
    a, b = 0, 1
    for _ in range(n):
        yield a
        a, b = b, a + b
```

```bash
# Build and run
go build -o mddeck ./cmd/mddeck/
./mddeck slides.md
```

## Horizontal Rules

Content above the rule.
***
Content between two rules.
___
Content below. Use `***` or `___` for in-slide rules.

## Links

Links like [mddeck on GitHub](https://github.com/miskun/mddeck) render as underlined accent-colored text.

You can have [multiple](https://example.com) links [in one](https://example.com) paragraph.

## ANSI Art

```ansi
\033[36m╔══════════════════════════════════╗\033[0m
\033[36m║\033[0m  \033[32m✓\033[0m Parse Markdown                \033[36m║\033[0m
\033[36m║\033[0m  \033[32m✓\033[0m Compute Layout                \033[36m║\033[0m
\033[36m║\033[0m  \033[32m✓\033[0m Render ANSI                   \033[36m║\033[0m
\033[36m║\033[0m  \033[33m⧗\033[0m Present to Audience           \033[36m║\033[0m
\033[36m╚══════════════════════════════════╝\033[0m
```

## ASCII Art

```ascii
    +-------------------+
    |   Terminal Input   |
    +--------+----------+
             |
    +--------v----------+
    |    Parse Blocks    |
    +--------+----------+
             |
    +--------v----------+
    |   Render Output    |
    +-------------------+
```

## Mixed Inline Styles

A paragraph with **bold text**, *italic text*, `code spans`, and [links](https://example.com) all mixed together.

- **Bold** list items with `code` and *emphasis*
- Regular items for contrast
- A [linked item](https://example.com) in a list

> A blockquote with **bold**, *italic*, and `code` inside.

---
layout: two-col
ratio: "50/50"
---

## Two-Column Layout

This content appears in the left column. Use slide frontmatter to set `layout: two-col`.

- Left point A
- Left point B
- Left point C

## Right Side

This content appears in the right column. Set custom ratios with `ratio: "50/50"`.

- Right point 1
- Right point 2
- Right point 3

---
layout: two-col
ratio: "70/30"
---

## Code Example

The `ratio` field controls column widths. This slide uses `70/30`.

```go
func render(slide Slide) string {
    buf := newScreenBuf(w, h)
    for _, block := range slide.Blocks {
        buf.Set(row, col, block)
    }
    return buf.String()
}
```

## Notes

- 70/30 ratio
- Code on the left
- Notes on the right

---
layout: terminal
---

## Terminal Layout

```
$ mddeck --present slides.md
╭─────────────────────────────────────────╮
│  Welcome to mddeck                      │
│                                          │
│  Terminal-native presentations,          │
│  powered by Markdown.                    │
│                                          │
│                               1 / 7      │
╰─────────────────────────────────────────╯
```

Full viewport width — ideal for code and art.

## Strikethrough

Use `~~text~~` to render ~~struck-through text~~.

Useful for showing ~~old~~ revised wording or ~~deprecated~~ updated features.

## Task Lists

Track progress with checkbox lists:

- [x] Implement parser
- [x] Build renderer
- [ ] Add syntax highlighting
- [ ] Write documentation
- [ ] Ship v2.0

## Nested Lists

Lists can be nested using indentation:

- Top-level item
  - Sub-item one
  - Sub-item two
    - Deep nested item
- Another top-level
  - With a child

Ordered lists nest too:

1. First step
   1. Sub-step A
   2. Sub-step B
2. Second step
   1. Another sub-step

## Tables

| Feature        | Status | Priority |
|----------------|--------|----------|
| Strikethrough  | Done   | High     |
| Nested Lists   | Done   | Medium   |
| Task Lists     | Done   | High     |
| Tables         | Done   | High     |
| Line Breaks    | Done   | Medium   |
| Callouts       | Done   | Low      |

Tables render with box-drawing characters for clean terminal display.

## Line Breaks

Markdown normally joins consecutive lines into one paragraph.

A trailing backslash forces a\
hard line break right here.

Two trailing spaces also work  
for hard line breaks.

Without a break marker,
these lines merge into one paragraph.

## Alerts & Callouts

> [!NOTE]
> This is a note callout — useful for additional context.

> [!TIP]
> Quick tips stand out with their own icon and color.

> [!WARNING]
> Warns about potential problems or pitfalls.

> [!IMPORTANT]
> Highlights critical information the reader must know.

> [!CAUTION]
> Signals danger — something that could cause data loss or harm.

## Speaker Notes

This slide has speaker notes attached. Press `t` to toggle presenter mode and see them.

In presenter mode you get:

- Current slide and next slide preview
- Speaker notes
- Elapsed timer

???
These are the speaker notes! Only visible in presenter mode.

Remind the audience about the `t` key.

## Keyboard Shortcuts

- **Space** / **Enter** / **→** — Next slide
- **Backspace** / **←** — Previous slide
- **Home** / **End** — First / last slide
- **t** — Toggle presenter mode
- **?** — Help overlay
- **q** — Quit

## Themes

Three built-in themes, set via `--theme` or frontmatter:

- `default` — Cyan accent on default background
- `dark` — Magenta accent for dark terminals
- `light` — Blue accent for light terminals

Try: `mddeck --theme dark example.md`

## How Slide Splitting Works

Most slides in this file split on `##` headers automatically — no `---` needed.

For special layouts, a frontmatter block starts a new slide:

```
---
layout: two-col
ratio: "50/50"
---
```

Both styles coexist in the same file.

---
layout: sidebar
---

## Sidebar Layout

This is a narrow sidebar.

## Main Content

The **sidebar** layout uses `columns: [30, 70]` to create a narrow left panel and a wide right panel. Content blocks are distributed round-robin across the regions.

---
layout: thirds
---

## Column A

First third of the content.

## Column B

Second third with *emphasis*.

## Column C

Third column to the right.

---
layout: quad
---

## Top Left

Quarter one.

## Top Right

Quarter two.

## Bottom Left

Quarter three.

## Bottom Right

Quarter four.

# Thank You!

Press `q` to quit.
