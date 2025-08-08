package config

import "gitlab.garfield-labs.com/apps/aggr/internal/patterns"

// Application constants.
const (
	// DefaultIgnoreFile is the default name for the ignore file.
	DefaultIgnoreFile = ".aggrignore"

	// DefaultPattern matches current directory and all it's subdirectories.
	DefaultPattern = "."

	// DefaultMaxSize is the default maximum size of files to include in aggregation.
	DefaultMaxSize = "1 mb"

	// DefaultMaxFiles is the default maximum number of files to include in aggregation.
	DefaultMaxFiles = 1000
)

// Default exclude patterns that are always applied.
var DefaultExcludes = patterns.Patterns{
	// Exclude all kinds of executables
	"*.exe",

	// Exclude some commonly ignore files
	".aggrignore",
	"go.mod",
	"go.sum",

	// Exclude some known directories
	".git/",
	".vscode-server/",
	"node_modules/",
	"vendor/",
	".task/",
	".cache/",
}

// DefaultHidden is the default patterns for hidden files and directories.
var DefaultHidden = patterns.Patterns{
	// Exclude hidden files and directories by default
	".*",
}
