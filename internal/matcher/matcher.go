package matcher

import (
	"os"

	"github.com/idelchi/godyl/pkg/path/files"

	"gitlab.garfield-labs.com/apps/aggr/internal/checkers"
	"gitlab.garfield-labs.com/apps/aggr/internal/walker"
)

// Logger is an interface for logging formatted messages.
type Logger interface {
	Debugf(format string, v ...any)
}

// Globber is a file matcher that compiles a list of files matching a given pattern, while
// excluding files based on provided exclude patterns and options.
type Globber struct {
	// Checkers are additional checks that can be applied to files.
	Checkers checkers.Checkers
	// Logger is a logger for debug messages (mainly).
	Logger Logger
	// Files is the list of files that are added to the matcher, after matching and applying the options.
	Files files.Files
	// Max is the maximum number of files to collect.
	Max int
}

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
