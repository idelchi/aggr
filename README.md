# aggr

A tool to aggregate and unpack files from directories</p>

---

[![GitHub release](https://img.shields.io/github/v/release/idelchi/aggr)](https://github.com/idelchi/aggr/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/idelchi/aggr.svg)](https://pkg.go.dev/github.com/idelchi/aggr)
[![Go Report Card](https://goreportcard.com/badge/github.com/idelchi/aggr)](https://goreportcard.com/report/github.com/idelchi/aggr)
[![Build Status](https://github.com/idelchi/aggr/actions/workflows/github-actions.yml/badge.svg)](https://github.com/idelchi/aggr/actions/workflows/github-actions.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`aggr` is a command-line utility that recursively aggregates files from specified paths into a single file and unpacks them back to their original directory structure.

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
# Pack the current directory (excluding ignored files) into stdout
aggr pack
```

```sh
# Pack specific paths
aggr pack src/ docs/
```

```sh
# Pack using glob patterns
aggr pack '**/folder/**/*.go'
```

```sh
# Unpack an archive to a specific directory
aggr unpack -o extracted/ pack.aggr
```

## Format

Archives are plain text files with simple markers to delimit file content.

```
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

If a _line_ in your file content starts with the marker prefix (after optional spaces/tabs), it gets escaped on pack and unescaped on unpack. Lines that contain the marker elsewhere are left alone. This keeps the archive parseable without mutating normal content.

## Path semantics (important)

- **Root directory**: By default the root is the current working directory. Use `--root DIR` or `-C DIR` to change it.
- **Patterns**: Input patterns must be **relative** to the root and **cannot** contain absolute paths or any `..` segment. If you want to work outside CWD, use `-C`.
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

```
# Ignore logs and build artifacts
*.log
build/
dist/
```

You can also provide additional ignore patterns via CLI:

```sh
aggr pack -i "*.tmp" -i "node_modules/"
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
aggr pack -x go -x md
```

This is implemented as an "allow-list" layer using ignore patterns under the hood.

## Commands and Flags

<details>
<summary><strong>pack</strong> — Aggregate files into an archive</summary>

- **Usage:**

  - `aggr pack [patterns|paths...]`

- **Aliases:**

  - `p`

- **Flags:**
  - `--output`, `-o` – Output file (default: stdout).
  - `--ignore`, `-i` – Additional ignore pattern (repeatable).
  - `--size`, `-s` – Max file size to include (e.g., `500kb`, `1mb`). Default: `1mb`.
  - `--max`, `-m` – Max number of files to include. Default: `1000`.
  - `--hidden`, `-a` – Include hidden files and directories.
  - `--extensions`, `-x` – Only include listed file extensions (repeatable).
  - `--root`, `-C` – Set the root directory for matching and reading files.
  - `--binary`, `-b` – Include binary files.
  - `--dry-run`, `-d` – Show what would be packed without writing output.

</details>

<details>
<summary><strong>unpack</strong> — Extract files from an archive</summary>

- **Usage:**

  - `aggr unpack <file>`

- **Aliases:**

  - `u`, `x`

- **Flags:**

  - `--output`, `-o` – Output directory. Default: `aggr-<hash-of-archive>` in the current directory.
  - `--ignore`, `-i` – Ignore patterns applied _during extraction_.
  - `--ext`, `-x` – Only extract files with these extensions (repeatable).
  - `--dry` – Show what would be unpacked without writing files.

- **Note:** If the output directory already exists, you'll be prompted to confirm before potentially overwriting files.

</details>

## Peculiarities & gotchas

- **No absolute paths, no `..`:** For safety, any absolute path or pattern containing a `..` segment is rejected. Use `-C` if you need to work elsewhere.
- **Pattern normalization is opinionated:** `.` becomes `**`, and a plain directory becomes recursive. If you want exact matching behavior, use explicit globs.
- **Markers only escape at line start:** Only lines that _start_ with the marker (after spaces/tabs) get escaped. Text in the middle of a line is left as-is.
- **Binary detection is conservative:** Files that look binary are skipped. If you need to force-include something unusual, use `--binary/-b` to disable the check.
