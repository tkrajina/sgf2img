# AGENTS.md - Guidelines for AI Coding Agents

This document provides guidelines for AI agents working on the sgf2img codebase.
sgf2img is a Go toolset for working with SGF (Smart Game Format) files used to record Go game records.

## Project Structure

```
sgf2img/
├── cmd/                    # CLI applications (12 tools)
│   ├── sgf2img/            # Convert SGF to images (PNG/SVG)
│   ├── sgf2ankicsv/        # Convert SGF to Anki CSV flashcards
│   ├── sgf2kifu/           # Generate kifu (game record) HTML
│   ├── sgfaddcolortoplay/  # Add player color info to SGF
│   ├── sgfcleancomments/   # Clean/modify SGF comments
│   ├── sgfcleankatrain/    # Remove KaTrain analysis data
│   ├── sgffindpos/         # Find board positions in SGF files
│   ├── sgfinfo/            # Display SGF game information
│   ├── sgflongestmainline/ # Extract longest game line
│   ├── sgfrename/          # Rename SGF files based on metadata
│   ├── sgfs2md/            # Generate markdown/HTML from SGFs
│   └── sgfstrip/           # Strip comments/branches from SGF
├── sgfutils/               # Core library package
│   └── sgf2img/            # Image generation subpackage
├── utils/                  # Generic utility functions
└── sgf/                    # Example SGF files
```

## Build, Test, and Lint Commands

```bash
# Build all packages
go build ./...

# Build a specific CLI tool
go build ./cmd/sgf2img/

# Install all tools to a directory
make install DIR=/path/to/bin

# Run all tests
go test ./...

# Run tests in a specific package
go test ./sgfutils/sgf2img/

# Run a single test by name
go test -run TestExpandCoordinates ./sgfutils/sgf2img/

# Run tests with verbose output
go test -v ./...

# Format all Go files
gofmt -w .

# Run go vet
go vet ./...

# Generate example images (runs clean first)
make run
```

## Code Style Guidelines

### Imports

Group imports in this order, separated by blank lines:
1. Standard library
2. External dependencies
3. Internal packages

```go
import (
    "fmt"
    "strings"

    "github.com/rooklift/sgf"

    "github.com/tkrajina/sgf2img/sgfutils"
)
```

### Naming Conventions

- **Exported identifiers**: PascalCase (`ProcessSGFFile`, `GameInfo`, `Options`)
- **Unexported identifiers**: camelCase (`panicIfErr`, `boardToImage`, `walkNodes`)
- **Constants**: PascalCase with category prefix (`SGFTagComment`, `SGFTagBlackMove`)
- **Struct fields**: PascalCase for exported, camelCase for internal

### Constants

Define related constants in `const` blocks with a common prefix:

```go
const (
    SGFTagComment   = "C"
    SGFTagBlackMove = "B"
    SGFTagWhiteMove = "W"
)
```

### Error Handling

**In CLI tools** - Use a `panicIfErr` helper for fatal errors:

```go
func panicIfErr(err error) {
    if err != nil {
        panic(err)
    }
}
```

**In library code** - Return errors as the last return value:

```go
func ProcessSGFFile(sgfFn string, opts *Options) (*sgf.Node, []GobanImageFile, error) {
    node, err := sgf.Load(sgfFn)
    if err != nil {
        return nil, nil, err
    }
    // ...
}
```

Use `fmt.Errorf` for error wrapping with context:

```go
return nil, fmt.Errorf("can't find node with img name '%s': %w", name, err)
```

### Testing Patterns

Use table-driven tests with `testify/assert`:

```go
func TestExpandCoordinates(t *testing.T) {
    t.Parallel()

    for _, data := range []struct {
        str      string
        expected []string
    }{
        {"dd:de", []string{"dd", "de"}},
        {"dd:df", []string{"dd", "de", "df"}},
    } {
        coords, err := expandCoordinatesRange(data.str, 19)
        assert.Nil(t, err)
        assert.Equal(t, data.expected, coords)
    }
}
```

### Types and Generics

Use Go 1.18+ generics for utility functions where appropriate:

```go
func MinMax[T int | int32 | int64 | float32 | float64](t1, t2 T) (T, T) {
    if t1 < t2 {
        return t1, t2
    }
    return t2, t1
}
```

### CLI Tool Pattern

Each CLI tool follows this structure:
- Single file in `cmd/<toolname>/<toolname>.go`
- Package `main`
- Uses `flag` package for argument parsing
- Defines `panicIfErr` helper locally
- Main logic often in a separate `doStuff()` function that returns error

```go
func main() {
    flag.BoolVar(&args.verbose, "v", false, "Verbose output")
    flag.Parse()
    panicIfErr(doStuff())
}
```

## Key Dependencies

- `github.com/rooklift/sgf` - SGF file parsing and manipulation
- `github.com/llgcode/draw2d` - 2D graphics rendering (PNG/SVG)
- `github.com/kettek/apng` - Animated PNG support
- `github.com/stretchr/testify` - Test assertions
- `github.com/golang/freetype` - TrueType font handling

## Go Version

This project requires Go 1.18 or later (uses generics).
