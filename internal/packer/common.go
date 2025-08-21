package packer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/idelchi/godyl/pkg/logger"
	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"

	"github.com/idelchi/aggr/internal/config"
	"github.com/idelchi/aggr/internal/patterns"
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

// ActiveAggrignore searches for and returns an active .aggrignore file.
// It checks the current directory and user config directory for ignore files.
// Returns the found file and true if an ignore file exists, otherwise false.
func ActiveAggrignore() (file.File, bool) {
	files := files.New(".", config.DefaultIgnoreFile)

	configDir, err := os.UserConfigDir()
	if err == nil {
		files = append(files, file.New(configDir, config.DefaultIgnoreFile))
	}

	return files.Exists()
}

// ExtensionsToPatterns converts a list of file extensions to ignore patterns.
// Each extension is converted to a negated pattern (e.g., "go" becomes "!*.go").
func ExtensionsToPatterns(extensions []string) patterns.Patterns {
	patterns := patterns.Patterns{}

	for _, ext := range extensions {
		patterns = append(patterns, "!*."+ext, "!**/*."+ext)
	}

	return patterns
}

// ExtensionsToPaths converts a list of file extensions to paths.
// Each extension is converted to a path pattern (e.g., "go" becomes "**/*.go").
// A path 'internal' with an extension 'go' becomes "internal/**/*.go".
func ExtensionsToPaths(extensions []string, paths patterns.Patterns) patterns.Patterns {
	patterns := patterns.Patterns{}

	var pattern string

	for _, path := range paths {
		path = strings.TrimRight(path, "/")

		switch {
		case strings.HasSuffix(path, "**"):
			pattern = "%s/*.%s"
		}

		for _, ext := range extensions {
			patterns = append(patterns, fmt.Sprintf(pattern, path, ext))
		}

	}

	return patterns
}

// ExtensionsToPaths converts normalized search patterns to per-extension globs.
// Assumes search has been normalized: "." -> "**", "dir" -> "dir/**", etc.
func ExtensionsToPaths2(extensions []string, search patterns.Patterns, root string) patterns.Patterns {
	out := patterns.Patterns{}

	for _, pat := range search {
		p := file.New(pat).String()

		switch {
		// "**" or "dir/**" → base + "**/*.ext"
		case strings.HasSuffix(p, "**"):
			base := strings.TrimSuffix(p, "**")
			base = strings.TrimRight(base, "/")
			for _, ext := range extensions {
				out = append(out, fmt.Sprintf("%s/**/**/*.%s", base, ext)) // see note below
			}

		// no meta: could be file or dir
		default:
			// dir?
			if file.New(root, p).IsDir() || strings.HasSuffix(p, "/") {
				base := strings.TrimRight(p, "/")
				for _, ext := range extensions {
					out = append(out, fmt.Sprintf("%s/**/*.%s", base, ext))
				}
				break
			}
			// file?
			ext := strings.TrimPrefix(filepath.Ext(p), ".")
			for _, e := range extensions {
				if strings.EqualFold(ext, e) {
					out = append(out, p)
					break
				}
			}
		}
	}

	// Fallback: if nothing produced, search whole tree for the exts
	if len(out) == 0 {
		for _, ext := range extensions {
			out = append(out, fmt.Sprintf("**/*.%s", ext))
		}
	}

	return out
}
