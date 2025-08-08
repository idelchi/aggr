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

// Walker traverses file systems and applies filtering rules to discovered files.
// It uses doublestar for glob pattern matching and applies checker rules to each file.
type Walker struct {
	// Checkers contains the validation rules applied to each discovered file.
	Checkers checkers.Checkers
	// Logger receives debug messages during the walking process.
	Logger Logger
	// MaxWalk sets the maximum number of files to process before stopping.
	MaxWalk int
}

// Logger is an interface for logging formatted debug messages.
type Logger interface {
	// Debugf formats and logs a debug message.
	Debugf(format string, v ...any)
}

// Walk traverses the file system using the given pattern and returns all regular files
// that pass the configured checkers. It respects the maximum file limit and current count.
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
