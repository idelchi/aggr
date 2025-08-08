package patterns

import (
	"strings"

	ignore "github.com/sabhiram/go-gitignore"

	"github.com/idelchi/godyl/pkg/path/file"
)

// Lines retrieves the file contents as a slice of strings.
func AsLines(data []byte) []string {
	// Count lines for pre-allocation
	lineCount := strings.Count(string(data), "\n") + 1

	returns := make([]string, 0, lineCount)

	for line := range strings.SplitSeq(string(data), "\n") {
		returns = append(returns, strings.TrimRight(line, "\r"))
	}

	return returns
}

// LoadIgnoreFile reads ignore patterns from a file (like .aggignore).
func LoadIgnoreFile(file file.File, patterns Patterns) (ignorer *ignore.GitIgnore, err error) {
	return ignore.CompileIgnoreFileAndLines(file.Path(), patterns...)
}
