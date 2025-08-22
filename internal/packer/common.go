package packer

import (
	"fmt"
	"os"

	"github.com/idelchi/aggr/internal/config"
	"github.com/idelchi/aggr/internal/patterns"
	"github.com/idelchi/godyl/pkg/logger"
	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"
)

// Logger creates and returns a logger with the appropriate level based on dry run mode.
// In dry run mode, it sets the level to DEBUG for verbose output.
func Logger(dry bool) (*logger.Logger, error) {
	level := logger.INFO

	if dry {
		level = logger.DEBUG
	}

	log, err := logger.New(level)
	if err != nil {
		return nil, fmt.Errorf("creating logger with level %d: %w", level, err)
	}

	return log, nil
}

// GetOutputWriter returns an output writer based on the provided options.
// If output is set to stdout, it returns os.Stdout, otherwise it creates a new file.
func GetOutputWriter(options config.Options) (*os.File, error) {
	if options.IsStdout() {
		return os.Stdout, nil
	}

	file := file.New(options.Output)

	if err := file.Create(); err != nil {
		return nil, fmt.Errorf("creating output file %s: %w", options.Output, err)
	}

	return file.OpenForWriting()
}

// DefaultAggrignores searches for and returns the default .aggrignore file.
// It checks the current directory and ~/.config/aggr for ignore files, finally
// falling back to .gitignore in the current directory.
// Returns the found file and true if an ignore file exists, otherwise false.
func DefaultAggrignores() file.File {
	files := files.New(config.DefaultIgnoreFile)

	home, err := os.UserHomeDir()
	if err == nil {
		files.AddFile(file.New(home, ".config", config.Name, config.DefaultIgnoreFile))
	}

	files.AddFile(file.New(".gitignore"))

	file, _ := files.Exists()

	return file
}

// ExtensionsToPatterns converts a list of file extensions to ignore patterns.
// Each extension is converted to a negated pattern (e.g., "go" becomes "!*.go").
func ExtensionsToPatterns(extensions []string) patterns.Patterns {
	patterns := patterns.Patterns{}

	for _, ext := range extensions {
		patterns = append(patterns, "!*."+ext)
	}

	return patterns
}
