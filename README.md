# aggr

A tool to aggregate and unpack files from directories</p>

---

[![GitHub release](https://img.shields.io/github/v/release/idelchi/aggr)](https://github.com/idelchi/aggr/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/idelchi/aggr.svg)](https://pkg.go.dev/github.com/idelchi/aggr)
[![Go Report Card](https://goreportcard.com/badge/github.com/idelchi/aggr)](https://goreportcard.com/report/github.com/idelchi/aggr)
[![Build Status](https://github.com/idelchi/aggr/actions/workflows/github-actions.yml/badge.svg)](https://github.com/idelchi/aggr/actions/workflows/github-actions.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`aggr` is a command-line utility that recursively aggregates files from specified paths into a single file and
unpacks them back to their original directory structure.

## Features

- Pack multiple files and directories into a single text archive.
- Unpack archives, restoring the original directory structure.
- Filter files by size, extension, and path using glob + `.gitignore`-style patterns.
- Supports a project `.aggrignore` (and a user-level one in your OS config dir).
- Skips binary files automatically and, by default, hidden files and common VCS/build dirs.

## Installation

```sh
curl -sSL https://raw.githubusercontent.com/idelchi/aggr/refs/heads/main/install.sh | sh -s -- -d ~/.local/bin
```

## Usage

```sh
# Pack the current directory (excluding ignored files) into `pack.aggr`
aggr
```

```sh
# Pack specific paths
aggr src/ docs/
```

```sh
# Pack using glob patterns
aggr '**/folder/**/*.go'
```

```sh
# Unpack an archive to a specific directory
aggr --unpack -o __extracted__ pack.aggr
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

### Marker escaping

If a _line_ in your file content starts with the marker prefix (after optional spaces/tabs),
it gets escaped on pack and unescaped on unpack. Lines that contain the marker elsewhere are left alone.
This keeps the archive parseable without mutating normal content.

## Path semantics (important)

- **Root directory**: By default the root is the current working directory. Use `--root DIR` or `-C DIR` to change it.
- **Patterns**: Input patterns must be **relative** to the root and **cannot** contain absolute paths or any `..` segment.
  If you want to work outside CWD, use `-C`.
- **Normalization**:
  - `.` becomes `**` (current dir + all subdirs).
  - A directory path with no glob meta (e.g. `foo/`) is treated as recursive: `foo/**`.
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

**Defaults** (always applied unless you override with your own patterns):

- Common VCS/build dirs: `.git/`, `vendor/`, `node_modules/`, etc.
- Hidden files/dirs (those starting with `.`) are excluded **by default**. Use `-a/--hidden` to include them.
- The output file itself (when `-o` is used).
- The `aggr` executable.
- **Binary files** are skipped automatically.

### Extensions include list

You can use `-x/--extensions` to "invert" selection by extension, e.g.:

```sh
# Only include .go and .md (plus anything force-included by your own patterns)
aggr -x go -x md
```

This is implemented as an "allow-list" layer using ignore patterns under the hood.

### Flags

- `--unpack`, `-u` – Unpack from a packed file
- `--output`, `-o` – Specify output file/folder. For --unpack, defaults to '$(pwd)/aggr-[hash of <file>]'
- `--root`, `-C` – Root directory to use
- `--extensions`, `-x` – File extensions to include (repeatable)
- `--ignore`, `-i` – Additional .aggignore patterns (repeatable)
- `--hidden`, `-a` – Include hidden files and directories
- `--binary`, `-b` – Include binary files
- `--size`, `-s` – Maximum size of file to include
- `--max`, `-m` – Maximum number of files to include
- `--dry-run`, `-d` – Show which files would be processed without reading contents
- `--parallel`, `-j` – Number of parallel workers to use

**Note:** If the output directory already exists, you'll be prompted to confirm before potentially overwriting files.

## Peculiarities & gotchas

- **No absolute paths, no `..`:** Any absolute path or pattern containing a `..` segment is rejected.
  Use `-C` if you need to work elsewhere.
- **Pattern normalization is opinionated:** `.` becomes `**`, and a plain directory becomes recursive.
  If you want exact matching behavior, use explicit globs.
- **Binary detection is conservative:** Files that look binary are skipped.
  If you need to force-include something unusual, use `--binary/-b` to disable the check.
