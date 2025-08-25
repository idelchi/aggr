package checkers

import (
	"fmt"

	"github.com/idelchi/godyl/pkg/path/file"
)

// Ignorer is an interface for checking if a file or directory is ignored.
type Ignorer interface {
	Ignored(path string, isDir bool) bool
}

// Ignore is a checker that filters files based on gitignore-style patterns.
type Ignore struct {
	ignore Ignorer
}

// NewIgnore creates a new Ignore checker with the provided GitIgnore matcher.
func NewIgnore(ignore Ignorer) *Ignore {
	return &Ignore{ignore: ignore}
}

// Check returns an error if the file matches any of the configured ignore patterns.
func (i *Ignore) Check(path string) error {
	isDir := file.New(path).IsDir()

	if ok := i.ignore.Ignored(path, isDir); ok {
		if isDir {
			return fmt.Errorf("%w: dir in ignore patterns", ErrPrune)
		}

		return fmt.Errorf("%w: file in ignore patterns", ErrSkip)
	}

	return nil
}
