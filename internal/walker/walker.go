// Package walker provides a file system walker that filters results based on checkers.
package walker

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/idelchi/aggr/internal/checkers"
	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"
)

// Logger is an interface for logging formatted debug messages.
type Logger interface {
	// Debugf formats and logs a debug message.
	Debugf(format string, v ...any)
}

// New creates a new Walker with the specified checkers, file limit, and logger.
// It automatically adds a Seen checker to prevent duplicate file inclusion.
func New(checks checkers.Checkers, maxFiles int, logger Logger) *Walker {
	walker := Walker{
		Logger: logger,
		Max:    maxFiles,
	}

	checkers := append(checkers.Checkers{
		checkers.NewSeen(&walker.Files),
	}, checks...,
	)

	walker.Checkers = checkers

	return &walker
}

// Walker traverses file systems and applies filtering rules to discovered files.
// It uses doublestar for glob pattern matching and applies checker rules to each file.
type Walker struct {
	// Checkers contains the validation rules applied to each discovered file.
	Checkers checkers.Checkers
	// Logger receives debug messages during the walking process.
	Logger Logger
	// Max sets the maximum number of files to process before stopping.
	Max int
	// Files holds the collection of files that passed all checks.
	Files files.Files
}

// Walk traverses the file system using the given pattern and returns all regular files
// that pass the configured checkers. It stops and returns an error if the maximum file limit is reached.
//
// TODO(Idelchi): Write tests where fsys is mocked by fstest.MapFS{}.
func (w *Walker) Walk(fsys fs.FS, pattern string, opts ...doublestar.GlobOption) error {
	base := fmt.Sprintf("%s", fsys)

	err := doublestar.GlobWalk(
		fsys, pattern,
		func(p string, dir fs.DirEntry) error {
			if p == "." {
				return nil
			}

			fullPath := file.New(p)

			if err := w.Checkers.Check(base, fullPath.Path()); err != nil {
				w.Logger.Debugf("  - %q: %v", fullPath, err)

				switch {
				case errors.Is(err, checkers.ErrAbort):
					return fs.SkipAll
				case errors.Is(err, checkers.ErrPrune):
					return fs.SkipDir
				default:
					return nil // skip this file but keep walking siblings
				}
			}

			if !dir.IsDir() {
				w.Files.AddFile(fullPath)

				w.Logger.Debugf("  - %q: included", fullPath)

				if len(w.Files) > w.Max {
					w.Logger.Debugf("%v: max files reached: %d", checkers.ErrAbort, w.Max)

					return fmt.Errorf("%w: max files reached: %d: %w", checkers.ErrAbort, w.Max, fs.SkipAll)
				}
			}

			return nil
		},

		opts...,
	)

	return err
}
