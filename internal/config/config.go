package config

// Options holds the configuration settings for the aggregation tool.
type Options struct {
	// Output specifies the output file path for aggregated data.
	Output string
	// Dry indicates whether to perform a dry run without writing output.
	Dry bool
	// Parallel defines the number of parallel workers to use during processing.
	Parallel int
	// Rules contains the file filtering and processing rules.
	Rules Rules
	// Unpack specifies whether to unpack.
	Unpack bool
}

// Rules defines the filtering and processing rules for file aggregation.
type Rules struct {
	// Root defines the root directory for the aggregation operation.
	Root string
	// IgnoreFile specifies the path to the .aggrignore file.
	IgnoreFile IgnoreFile
	// Patterns contains ignore patterns to apply when collecting files.
	Patterns []string
	// Extensions defines the file extensions to include in aggregation.
	Extensions []string
	// Hidden indicates whether to include hidden files and directories.
	Hidden bool
	// Max defines the maximum number of files to collect.
	Max int
	// Size defines the maximum file size to include in aggregation.
	Size string
	// Binary indicates whether to include binary files in the aggregation.
	Binary bool
}

// IgnoreFile represents a .aggrignore file.
type IgnoreFile struct {
	// Path is the file path to the .aggrignore file.
	Path string
	// Set indicates whether the ignore file is set.
	Set bool
}

// IsStdout returns true if the output is set to stdout.
func (o Options) IsStdout() bool {
	return o.Output == "" || o.Output == "-"
}
