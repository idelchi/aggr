// Package walker provides a file system walker that filters results based on checkers.
package walker

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"

	"gitlab.garfield-labs.com/apps/aggr/internal/checkers"
)

// Walker filters doublestar results through a list of checkers.
type Walker struct {
	// Checkers are the checkers that will be applied to each file.
	Checkers checkers.Checkers
	// Logger is the logger for debug messages.
	Logger Logger
	// MaxWalk is the maximum number of files to walk.
	MaxWalk int
}

// Logger is an interface for logging formatted messages.
type Logger interface {
	Debugf(format string, v ...any)
}

// Walk returns every regular file that satisfies the checkers, up to the maximum number of files.
func (w *Walker) Walk(fsys fs.FS, pattern string, current int, opts ...doublestar.GlobOption) (files.Files, error) {
	var keep files.Files
	err := doublestar.GlobWalk(
		fsys, pattern,
		func(p string, d fs.DirEntry) error {
			if p == "." {
				return nil
			}

			fullPath := file.New(p)

			if err := w.Checkers.Check(fullPath.Path()); err != nil {

				if !strings.Contains(fullPath.Path(), ".git") {
					w.Logger.Debugf("  - %q: %v", fullPath, err)
				}

				if errors.Is(err, checkers.ErrAbort) {
					return fs.SkipAll
				}

				return nil
			}

			if !d.IsDir() {
				keep.AddFile(fullPath)

				w.Logger.Debugf("  - %q: included", fullPath)

				if len(keep)+current > w.MaxWalk {
					w.Logger.Debugf("%v: max files reached: %d", checkers.ErrAbort, w.MaxWalk)

					return fmt.Errorf("%w: max files reached: %d: %w", checkers.ErrAbort, w.MaxWalk, fs.SkipAll)
				}
			}

			return nil
		},

		opts...,
	)

	return keep, err
}
