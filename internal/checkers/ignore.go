package checkers

import (
	"fmt"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"

	"github.com/idelchi/godyl/pkg/path/file"
)

// Ignore is a checker that filters files based on gitignore-style patterns.
type Ignore struct {
	ignore *ignore.GitIgnore
}

// NewIgnore creates a new Ignore checker with the provided GitIgnore matcher.
func NewIgnore(ignore *ignore.GitIgnore) *Ignore {
	return &Ignore{ignore: ignore}
}

// Check returns an error if the file matches any of the configured ignore patterns.
func (i *Ignore) Check(path string) error {
	if err := i.check(path); err != nil {
		return err
	}

	// // TODO(Idelchi): Do we really need this?
	if file.New(path).IsDir() {
		path = strings.TrimRight(path, "/") + "/"
	}

	return i.check(path)
}

// check checks if the given path matches any ignore patterns.
func (i *Ignore) check(path string) error {
	matchPath := path
	for strings.HasPrefix(matchPath, "../") {
		matchPath = strings.TrimPrefix(matchPath, "../")
	}

	isDir := file.New(path).IsDir()
	if isDir {
		matchPath = strings.TrimRight(matchPath, "/") + "/"
	}

	if ok, pattern := i.ignore.MatchesPathHow(matchPath); ok {
		if isDir {
			return fmt.Errorf("%w: dir in ignore patterns %q", ErrPrune, pattern.Pattern)
		}

		return fmt.Errorf("%w: file in ignore patterns %q", ErrSkip, pattern.Pattern)
	}

	fmt.Printf("  - %q: not in ignore patterns\n", path)

	return nil
}
