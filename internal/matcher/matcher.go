package matcher

import (
	"os"

	"github.com/idelchi/godyl/pkg/path/files"

	"gitlab.garfield-labs.com/apps/aggr/internal/checkers"
	"gitlab.garfield-labs.com/apps/aggr/internal/walker"
)

// Logger is an interface for logging formatted debug messages.
type Logger interface {
	// Debugf formats and logs a debug message.
	Debugf(format string, v ...any)
}

// Globber collects files matching specified patterns while applying various filtering rules.
// It maintains a collection of matched files and supports limiting the total number of files.
type Globber struct {
	// Checkers contains the validation rules applied to each file.
	Checkers checkers.Checkers
	// Logger receives debug messages during the matching process.
	Logger Logger
	// Files holds the collection of files that passed all checks.
	Files files.Files
	// Max specifies the maximum number of files to collect.
	Max int
}

// New creates a new Globber with the specified checkers, file limit, and logger.
// It automatically adds a Seen checker to prevent duplicate file inclusion.
func New(checks checkers.Checkers, max int, logger Logger) *Globber {
	matcher := Globber{
		Logger: logger,
		Max:    max,
	}

	checkers := append(checkers.Checkers{
		checkers.NewSeen(&matcher.Files),
	}, checks...,
	)

	matcher.Checkers = checkers

	return &matcher
}

// Match processes files matching the given path pattern within the specified root directory.
// It applies all configured checkers and adds matching files to the Files collection.
func (m *Globber) Match(root, path string) (err error) {
	walker := walker.Walker{
		Checkers: m.Checkers,
		Logger:   m.Logger,
		MaxWalk:  m.Max,
	}

	collected, err := walker.Walk(os.DirFS(root), path, len(m.Files))
	if err != nil {
		return err
	}

	m.Files = append(m.Files, collected...)

	return nil
}
