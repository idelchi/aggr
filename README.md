# aggr

A tool to filter and aggregate files into a single text file.

---

[![GitHub release](https://img.shields.io/github/v/release/idelchi/aggr)](https://github.com/idelchi/aggr/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/idelchi/aggr.svg)](https://pkg.go.dev/github.com/idelchi/aggr)
[![Go Report Card](https://goreportcard.com/badge/github.com/idelchi/aggr)](https://goreportcard.com/report/github.com/idelchi/aggr)
[![Build Status](https://github.com/idelchi/aggr/actions/workflows/github-actions.yml/badge.svg)](https://github.com/idelchi/aggr/actions/workflows/github-actions.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`aggr` is a command-line utility that recursively aggregates files from specified paths into a single file and
unpacks them back to their original directory structure.

- Filter files by size, extension, and path using glob + `.gitignore`-style patterns.
- Supports loading of a `.aggrignore` following `gitignore` conventions.
- Skips binary files automatically and, by default, hidden files and common VCS/build dirs.
- Unpack archives, restoring the original directory structure.

## Installation

```sh
curl -sSL https://raw.githubusercontent.com/idelchi/aggr/refs/heads/main/install.sh | sh -s -- -d ~/.local/bin
```

## Usage

```sh
# Pack the current directory (excluding ignored files) into `pack.aggr`
aggr -o pack.aggr
```

```sh
# Pack specific folders
aggr src docs
```

```sh
# Pack using glob patterns
aggr "**/folder/**/*.go"
```

## Format

Archives are plain text files with simple markers to delimit file content.

```text
// === AGGR: BEGIN: src/main.go ===
package main

import "fmt"

func main() {
    fmt.Println("Hello, world!")
}

// === AGGR: END: src/main.go ===

// === AGGR: BEGIN: README.md ===
# Project

Description here.
// === AGGR: END: README.md ===
```

## Path semantics

- **Root directory**: By default the root is the current working directory. Use `--root DIR` or `-C DIR` to change it.
- **Patterns**: Input patterns must be **relative** to the root and **cannot** contain absolute paths or any `..` segment.
  If you want to work outside CWD, use `-C`.
- **Normalization**:
  - `.` becomes `**` (current dir + all subdirs).
  - A directory path with no glob meta (e.g. `foo/`) is treated as recursive: `foo/**`.
    Same goes for a plain path (e.g. `foo`) that is detected as a directory.
  - If your pattern already has meta (`*?[{`), it's passed through.
- **Glob engine**: Uses [doublestar](https://github.com/bmatcuk/doublestar). Examples:
  - `**/*.go`
  - `pkg/**/testdata/**`
  - `**/Dockerfile`

## Filtering

Create an `.aggrignore` with `.gitignore`-style patterns:

```gitignore
# Ignore logs and build artifacts
*.log
build/
dist/
```

You can also provide additional ignore patterns via CLI:

```sh
aggr -i "*.tmp" -i "node_modules/"
```

**Default exclusions** (always applied unless you override with your own patterns):

- `.git/`
- Hidden files/dirs (those starting with `.`). Use `-a/--hidden` to include them
- The output file itself (when `-o` is used)
- The `aggr` executable
- **Binary files**

### Extensions include list

You can use `-x/--extensions` to "invert" selection by extension, e.g.:

```sh
# Only include .go and .md (plus anything force-included by your own patterns)
aggr -x go -x md
```

This is implemented as an "allow-list" layer using ignore patterns under the hood.

### Flags

- `--unpack`, `-u` – Unpack from a packed file
- `--output`, `-o` – Specify output file/folder.
  For packing, defaults to `<folder>.aggr`, for unpacking to `<file>-[hash of <file>]`
- `--root`, `-C` – Root directory to use
- `--file`, `-f` - Path to the `.aggrignore` file. Set to an empty string to completely ignore. When not passed, uses defaults
- `--extensions`, `-x` – File extensions to include (repeatable)
- `--ignore`, `-i` – Additional .aggrignore patterns (repeatable)
- `--hidden`, `-a` – Include hidden files and directories
- `--binary`, `-b` – Include binary files
- `--size`, `-s` – Maximum size of file to include
- `--max`, `-m` – Maximum number of files to include
- `--dry`, `-d` – Show which files would be processed without reading contents
- `--parallel`, `-j` – Number of parallel workers to use

When `--file` is not set, it defaults to the first found of `.aggrignore`, `~/.config/aggr/.aggrignore` and `.gitignore`.

**Note:** If the output directory already exists, you'll be prompted to confirm before potentially overwriting files.

## Peculiarities & gotchas

- **No absolute paths, no `..`:** Any absolute path or pattern containing a `..` segment is rejected.
  Use `-C` if you need to work elsewhere.
- **Pattern normalization is opinionated:** `.` becomes `**`, and a plain directory becomes recursive.
  If you want exact matching behaviour, use explicit globs.
- **Binary detection is conservative:** Files that look binary are skipped.
  If you need to force-include something unusual, use `--binary/-b` to disable the check.
- **Marker escaping:** If a line in your file content starts with the marker prefix (after optional spaces/tabs),
  it gets escaped on pack and unescaped on unpack. Lines that contain the marker elsewhere are left alone.
  This keeps the archive parseable without mutating normal content.

When multiple filtering rules are in play, patterns are applied in the following order
(later patterns can add to or negate earlier ones):

1. **Extension filters** - When using `-x/--extensions`, creates include patterns for those extensions
2. **Hidden files** - `.` prefixed files/folders are excluded (unless `-a/--hidden` is used)
3. **CLI ignore patterns** - Patterns specified with `-i/--ignore` flags
4. **Executable exclusion** - The `aggr` binary itself is automatically excluded
5. **Output file exclusion** - When using `-o`, the output file is excluded to prevent recursion
6. **`.aggrignore` file** - Patterns from the loaded ignore file
7. **Default excludes** - Built-in patterns for VCS directories (`.git/`, etc.)

This means that:

- `.aggrignore` and default excludes are applied last, allowing them to override earlier patterns
- CLI patterns (`-i`) are applied early, so `.aggrignore` patterns can override them
- Use negation patterns (e.g., `!.config/`) in `.aggrignore` to include specific hidden files/directories

You can see the order by passing `--dry`.
