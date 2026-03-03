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
  - `internal/layout/` — Layout engine (auto, title, center, two-col, split, terminal)
  - `internal/render/` — Markdown-to-ANSI renderer, presenter view, help overlay
  - `internal/runtime/` — Terminal raw mode, keyboard event loop, navigation

## Build & Test

```bash
go build ./cmd/mddeck/        # build binary
go test ./...                  # run all tests
./mddeck example.mddeck       # run with sample deck
./mddeck --present example.mddeck  # presenter mode
```

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
