package packer

import (
	"fmt"
	"os"

	"github.com/idelchi/godyl/pkg/logger"
	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"

	"gitlab.garfield-labs.com/apps/aggr/internal/config"
	"gitlab.garfield-labs.com/apps/aggr/internal/patterns"
)

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

func ActiveAggrignore() (file.File, bool) {
	files := files.New(".", config.DefaultIgnoreFile)
	configDir, err := os.UserConfigDir()
	if err == nil {
		files = append(files, file.New(configDir, config.DefaultIgnoreFile))
	}

	return files.Exists()
}

func ExtensionsToPatterns(extensions []string) patterns.Patterns {
	patterns := patterns.Patterns{}

	for _, ext := range extensions {
		patterns = append(patterns, fmt.Sprintf("!*.%s", ext))
	}

	return patterns
}
