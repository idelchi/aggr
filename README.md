# aggr

<h1 align="center">aggr</h1>

<p align="center">
<img alt="aggr logo" src="assets/images/agg.png" height="150" />
<p align="center">A tool to aggregate and unpack files from directories</p>
</p>

---

[![GitHub release](https://img.shields.io/github/v/release/idelchi/agg)](https://github.com/idelchi/agg/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/idelchi/agg.svg)](https://pkg.go.dev/github.com/idelchi/agg)
[![Go Report Card](https://goreportcard.com/badge/github.com/idelchi/agg)](https://goreportcard.com/report/github.com/idelchi/agg)
[![Build Status](https://github.com/idelchi/agg/actions/workflows/github-actions.yml/badge.svg)](https://github.com/idelchi/agg/actions/workflows/github-actions.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`aggr` is a command-line utility that recursively aggregates files from specified paths into a single file and unpacks them back to their original directory structure.

## Features

* Pack multiple files and directories into a single, human-readable archive.
* Unpack archives, restoring the original directory structure.
* Filter files by size, name, and path using glob and `.gitignore`-style patterns.
* Support for `.aggrignore` files for project-specific ignore rules.
* Automatically excludes binary files, hidden files, and common VCS/build directories by default.

## Installation

```sh
curl -sSL https://raw.githubusercontent.com/idelchi/agg/refs/heads/main/install.sh | sh -s -- -d ~/.local/bin
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
# Pack to a file
aggr pack -o archive.agg
```

```sh
# Unpack an archive
aggr unpack archive.agg
```

```sh
# Unpack to a specific directory
aggr unpack -o extracted/ archive.agg
```

## Format

Archives are plain text files with simple markers to delimit file content. This makes them easy to read and manipulate with standard text-based tools.

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

## Filtering

To exclude files, create an `.aggrignore` file in your project directory with `.gitignore`-style patterns:

```
# Ignore log files and build artifacts
*.log
build/
dist/
.git/
```

You can also provide ignore patterns directly on the command line:

```sh
aggr pack -i "*.tmp" -i "node_modules/"
```

By default, `aggr` excludes binary files, hidden folders and files (like `.env`, `.vscode`), and common directories (`.git/`, `vendor/`, etc.). Use the `--hidden` and `--no-defaults` flags to override this behavior.

## Commands and Flags

<details>
<summary><strong>pack</strong> — Aggregate files into an archive</summary>

- **Usage:**
  - `aggr pack [patterns|paths...]`

- **Aliases:**
  - `p`

- **Flags:**
  - `--output`, `-o` – Specify output file (default: stdout).
  - `--ignore`, `-i` – Add an ignore pattern (can be used multiple times).
  - `--size` – Maximum size of a file to include (e.g., "500kb", "1mb"). Default: "1mb".
  - `--max` – Maximum number of files to include. Default: 1000.
  - `--hidden`, `-a` – Include hidden files and directories (those starting with '.').
  - `--no-defaults` – Disable the default ignore patterns.
  - `--dry` – Show which files would be processed without creating the archive.

</details>

<details>
<summary><strong>unpack</strong> — Extract files from an archive</summary>

- **Usage:**
  - `aggr unpack <file>`

- **Aliases:**
  - `u`, `x`

- **Flags:**
  - `--output`, `-o` – Specify the output directory (default: current directory).
  - `--ignore`, `-i` – Add a pattern to ignore files during extraction (can be used multiple times).
  - `--dry` – Show which files would be unpacked without writing them to disk.

</details>

<details>
<summary><strong>version</strong> — Display version information</summary>

- **Usage:**
  - `aggr version`

- **Description:**
  - Prints the application version, commit hash, and build date.

</details>

## Demo

![Demo](assets/gifs/agg.gif)