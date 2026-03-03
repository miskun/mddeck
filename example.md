---
title: "Welcome to mddeck"
theme: "default"
wrap: true
tabSize: 2
---

# Welcome to mddeck

Terminal-native presentations, powered by Markdown.

???
This is the opening slide. Welcome everyone!

---

## What is mddeck?

- **Terminal-native** slide rendering
- Plain Markdown source files
- ANSI color support
- Speaker notes & presenter mode
- Multiple layout modes

???
Emphasize that the source file remains valid Markdown.

---

## Code Blocks

Here's some Go code:

```go
func main() {
    fmt.Println("Hello, mddeck!")
}
```

And inline code like `fmt.Println()` works too.

---

## ANSI Art Support

```ansi
\033[32m  _____ _____ _____ _____ \033[0m
\033[32m |     |   __|   __|   __|\033[0m
\033[33m |   --|   __|   __|__   |\033[0m
\033[31m |__|__|_____|_____|_____|\033[0m
```

---

## Lists & Quotes

1. First ordered item
2. Second ordered item
3. Third ordered item

> "The best way to predict the future is to invent it."
> — Alan Kay

???
Classic quote. Good transition point.

---

---
layout: two-col
ratio: "50/50"
---

## Left Column

This content goes on the left side of the slide.

- Point A
- Point B
- Point C

## Right Column

This content goes on the right side.

- Detail 1
- Detail 2
- Detail 3

---

---
layout: title
---

# Thank You!

Questions?

???
Open the floor for Q&A.
