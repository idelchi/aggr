package config

import "github.com/idelchi/aggr/internal/patterns"

// Application constants.
const (
	// Name is the name of the application.
	Name = "aggr"

	// DefaultIgnoreFile is the default name for the ignore file.
	DefaultIgnoreFile = ".aggrignore"

	// DefaultPattern matches current directory and all it's subdirectories.
	DefaultPattern = "."

	// DefaultMaxSize is the default maximum size of files to include in aggregation.
	DefaultMaxSize = "1 mb"

	// DefaultMaxFiles is the default maximum number of files to include in aggregation.
	DefaultMaxFiles = 1000
)

// DefaultExcludes lists exclude patterns that are always applied.
//
//nolint:gochecknoglobals 	// Fair use of global variables.
var DefaultExcludes = patterns.Patterns{
	".git/",
}

// DefaultHidden is the default patterns for hidden files and directories.
//
//nolint:gochecknoglobals 	// Fair use of global variables.
var DefaultHidden = patterns.Patterns{
	// Exclude hidden files and directories by default
	".*",
}
