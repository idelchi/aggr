// Package patterns provides utilities for working with file path patterns.
// It includes functions for normalizing, validating, and converting patterns.
package patterns

import (
	"fmt"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"

	"github.com/idelchi/godyl/pkg/path/file"
)

// Patterns contains path and gitignore patterns.
type Patterns []string

// AsGitIgnore compiles the patterns into a gitignore matcher.
func (p Patterns) AsGitIgnore() *ignore.GitIgnore {
	return ignore.CompileIgnoreLines(p...)
}

// TrimEmpty removes empty patterns from the list.
func (p Patterns) TrimEmpty() Patterns {
	var out Patterns
	for _, pat := range p {
		if strings.TrimSpace(pat) != "" {
			out = append(out, pat)
		}
	}
	return out
}

// Normalized returns a new Patterns instance with all patterns normalized.
func (p Patterns) Normalized(root string) Patterns {
	var out Patterns
	for _, pat := range p {
		out = append(out, Normalize(pat, root))
	}

	return out
}

// Validate checks all patterns for validity.
// Invalid patterns include:
// - Relative path traversals (e.g. "../foo")
// - Absolute paths (e.g. "/foo")
func (p Patterns) Validate() error {
	for _, pat := range p {
		if err := Validate(pat); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks a single pattern for validity.
// Invalid patterns include:
// - Relative path traversals (e.g. "../foo")
// - Absolute paths (e.g. "/foo")
func Validate(pat string) error {
	// Normalize separators so SplitSeq works consistently
	pat = filepath.ToSlash(pat)

	// Reject any segment that is exactly ".."
	for seg := range strings.SplitSeq(pat, "/") {
		if seg == ".." {
			return fmt.Errorf("relative path traversal (%q) is not allowed", pat)
		}
	}

	if filepath.IsAbs(pat) {
		return fmt.Errorf("absolute paths (%q) are not allowed", pat)
	}

	return nil
}

func ContainsMeta(pat string) bool {
	// Check for any glob meta characters
	return strings.ContainsAny(pat, "*?[{")
}

// Normalize turns "." → "**" and any dir path → "dir/**".
// It never touches patterns that already contain meta (*?[{) or /../.
func Normalize(pat, root string) string {
	pat = filepath.ToSlash(pat)

	// 1. Meta already present? leave unchanged
	if ContainsMeta(pat) {
		return pat
	}

	pat = filepath.ToSlash(filepath.Clean(pat))

	// 2. "." means "everything here"
	if pat == "." || pat == "./" {
		return "**"
	}

	// 3. If it ends with "/" assume dir
	if strings.HasSuffix(pat, "/") {
		return pat + "**"
	}

	// 4. Stat the path – is it an existing dir?
	if file.New(root, pat).IsDir() {
		return pat + "/**"
	}

	// 5. It’s either a file or a non‑existing path: leave it alone
	return pat
}
