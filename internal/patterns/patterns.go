package patterns

import (
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

// Normalize turns "." → "**" and any dir path → "dir/**".
// It never touches patterns that already contain meta (*?[{) or /../.
func Normalize(pat string) string {
	pat = filepath.ToSlash(filepath.Clean(pat))

	// 1. Meta already present? leave unchanged
	if strings.ContainsAny(pat, "*?[{") {
		return pat
	}

	// 2. "." means "everything here"
	if pat == "." || pat == "./" {
		return "**"
	}

	// 3. If it ends with "/" assume dir
	if strings.HasSuffix(pat, "/") {
		return pat + "**"
	}

	// 4. Stat the path – is it an existing dir?
	if file.New("", pat).IsDir() {
		return pat + "/**"
	}

	// 5. It’s either a file or a non‑existing path: leave it alone
	return pat
}
