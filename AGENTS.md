# mddeck

Terminal-native Markdown slide deck presenter, written in Go. Renders `.mddeck` (or `.md`) files directly in the terminal using ANSI styling.

## Project Overview

- **Language:** Go
- **Module:** `github.com/miskun/mddeck`
- **Entry point:** `cmd/mddeck/main.go`
- **Source layout:**
  - `cmd/mddeck/` — CLI binary
  - `internal/model/` — Core data types (Deck, Slide, Block)
  - `internal/parser/` — `.mddeck` file parser (frontmatter, slides, notes, blocks)
  - `internal/ansi/` — ANSI escape sequence handling and safety filtering
  - `internal/theme/` — Color theme definitions (default, dark, light)
  - `internal/layout/` — Layout engine (auto, title, center, cols-2, rows-2, cols-3, grid-4, sidebar, terminal)
  - `internal/render/` — Markdown-to-ANSI renderer, presenter view, help overlay
  - `internal/runtime/` — Terminal raw mode, keyboard event loop, navigation

## Build & Test

```bash
go build -o mddeck ./cmd/mddeck/  # build binary (always use -o to produce fresh binary)
go test ./...                      # run all tests
./mddeck example.md               # run with sample deck
./mddeck --present example.md     # presenter mode
./mddeck --dump example.md        # dump all slides as text to stdout
./mddeck --dump --format json example.md  # dump as structured JSON
```

**Important:** `go build ./...` only checks compilation — it does NOT write an output binary. Always use `go build -o mddeck ./cmd/mddeck/` to produce a runnable binary. Do this after every code change so `./mddeck` is always up to date.

## Slide Authoring: Multi-Region Layouts

Multi-region layouts (`cols-2`, `rows-2`, `cols-3`, `grid-4`, `sidebar`, `title-cols-2`, etc.) require multiple content regions. The parser's `mergeRegionChunks` pass automatically absorbs subsequent `---`-separated chunks to fill the required regions. Each absorbed boundary becomes a `BlockRegionBreak` that the layout engine uses to split content into columns/rows.

This means a `---` between column blocks in a multi-region layout is **not** a slide separator — it is an intentional region boundary that gets merged into the same slide:

```markdown
---
layout: title-cols-2
---

## Slide Title

Left column content

---

Right column content
```

The `---` between the two content blocks looks like a slide break but is absorbed as the column boundary. The slide above produces one slide with a title and two columns.

The number of regions consumed depends on the layout:
- `cols-2`, `rows-2`, `sidebar` → 2 regions
- `cols-3` → 3 regions
- `grid-4` → 4 regions
- `title-cols-2` → 3 regions (1 title + 2 columns)
- `title-cols-3` → 4 regions (1 title + 3 columns)

## Debugging & Troubleshooting

Use `--dump` mode to inspect parsed slide data without launching the TUI. This is the fastest way to verify what the parser produced, check block types, confirm reveal steps, or debug layout issues.

```bash
# Text dump of a single slide (1-based index)
./mddeck --dump --slide 3 example.md

# JSON dump — pipe to jq for targeted queries
./mddeck --dump --format json example.md | jq '.slides[0].blocks'

# Check block types across the whole deck
./mddeck --dump --format json example.md | jq '[.slides[].blocks[].type] | unique'

# Verify reveal steps on a specific slide
./mddeck --dump --format json --slide 5 example.md | jq '.slides[0].steps, [.slides[0].blocks[] | {type, step}]'
```

Virtual viewport flags (`--width`, `--height`) force specific terminal dimensions in both dump and TUI modes. Use them for deterministic rendering in tests or to simulate specific terminal sizes:

```bash
./mddeck --width 120 --height 40 example.md          # TUI at fixed size
```

When `--dump` is set, `--present`, `--watch`, and `--start` are ignored — it outputs to stdout and exits immediately.

## Code Conventions

- Standard Go project layout (`cmd/`, `internal/`)
- No external dependencies beyond `gopkg.in/yaml.v3` and `golang.org/x/term`
- Tests live alongside source files (`*_test.go`)
- ANSI escape codes are centralized in `internal/ansi/`
- All user-facing text rendering goes through `internal/render/`

---

# Release Process

## Prerequisites

- Go installed (`brew install go`)
- GitHub CLI authenticated (`gh auth status`)
- Working directory is clean (`git status`)

## Steps

### 1. Update version constant

Edit `cmd/mddeck/main.go` and update the `version` variable:

```go
var version = "1.x.x"
```

### 2. Commit the version bump

```bash
git add -A
git commit -m "release: v1.x.x"
git push
```

### 3. Tag the release

The tag **must** follow Go module versioning (`vMAJOR.MINOR.PATCH`). Without a tag, `go install ...@latest` will not resolve.

```bash
git tag v1.x.x
git push origin v1.x.x
```

### 4. Create a GitHub release (optional but recommended)

```bash
gh release create v1.x.x --title "v1.x.x" --notes "Release notes here"
```

### 5. Verify the install works

Wait 1–2 minutes for the Go module proxy to index the new tag, then:

```bash
go install github.com/miskun/mddeck/cmd/mddeck@latest
mddeck --version
```

If the proxy hasn't indexed yet, you can force it:

```bash
GOPROXY=https://proxy.golang.org go install github.com/miskun/mddeck/cmd/mddeck@v1.x.x
```

## Checklist

- [ ] Version string updated in `cmd/mddeck/main.go`
- [ ] All tests pass (`go test ./...`)
- [ ] Binary builds cleanly (`go build ./cmd/mddeck/`)
- [ ] Changes committed and pushed
- [ ] Git tag created and pushed (format: `v1.x.x`)
- [ ] `go install ...@latest` resolves the new version
- [ ] GitHub release created (optional)
