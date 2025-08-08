package cli

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

var (
	// Version information injected at build time
	appVersion  string
	buildCommit string
	buildDate   string
)

// Execute runs the root command with version information.
func Execute(version string) error {
	return newRootCmd(version).Execute()
}

// newRootCmd creates the root command.
func newRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "aggr",
		Version: version,
		Short:   "A tool to aggregate and unpack files from directories",
		Long: heredoc.Doc(`
			aggr is a command-line utility that recursively aggregates files
			from specified paths into a single file and unpacks them back to their
			original directory structure.
		`),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true

	// Add subcommands
	rootCmd.AddCommand(Pack())
	rootCmd.AddCommand(Unpack())

	return rootCmd
}
