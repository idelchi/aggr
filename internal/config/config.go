package config

// Options holds the available configurations for the aggregation tool.
type Options struct {
	// Output file for the aggregated data.
	Output string
	// DryRun indicates whether to perform a dry run without writing output.
	DryRun bool
	// Parallel defines the number of parallel workers to use.
	Parallel int
	// Rules defines the rules for file aggregation.
	Rules Rules
}

type Rules struct {
	// Root defines the root directory for the aggregation.
	Root string
	// Patterns to consider when collecting files.
	Patterns []string
	// Extensions defines the file extensions to include in the aggregation.
	Extensions []string
	// Hidden indicates whether to consider hidden files.
	Hidden bool
	// StripPrefix defines whether to strip the prefix from file paths.
	StripPrefix bool
	// Max defines the maximum number of files to collect.
	Max int
	// Size defines the maximum size for a file to be included in the aggregation.
	Size string
	// Binary indicates whether to include binary files.
	Binary bool
}

func (o Options) IsStdout() bool {
	return o.Output == "" || o.Output == "-"
}
