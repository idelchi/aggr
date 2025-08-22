package checkers

import (
	"fmt"
	"strings"

	"github.com/idelchi/godyl/pkg/path/file"
)

// Ignore is a checker that filters files based on gitignore-style patterns.
type Ignore struct {
	ignore Ignorer
}

type Ignorer interface {
	IsIgnored(path string, isDir bool) bool
}

// NewIgnore creates a new Ignore checker with the provided GitIgnore matcher.
func NewIgnore(ignore Ignorer) *Ignore {
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

	if ok := i.ignore.IsIgnored(matchPath, isDir); ok {
		if isDir {
			return fmt.Errorf("%w: dir in ignore patterns", ErrPrune)
		}

		return fmt.Errorf("%w: file in ignore patterns", ErrSkip)
	}

	return nil
}
