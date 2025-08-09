package cli

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

// Execute runs the root command for the aggr CLI application.
func Execute(version string) error {
	root := &cobra.Command{
		Use:   "aggr",
		Short: "Aggregate and unpack files",
		Long: heredoc.Doc(`
			aggr is a command-line utility that recursively aggregates files
			from specified paths into a single file and unpacks them back to their
			original directory structure.
		`),
		Example: heredoc.Doc(`
			# Pack all files in the current directory and all subdirectories
			$ aggr pack -o output.aggr

			# Unpack the contents of the archive
			$ aggr unpack output.aggr -o ./extracted
		`),
		Version:       version,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			// Do not print usage after basic validation has been done.
			cmd.SilenceUsage = true
		},
	}

	root.SetVersionTemplate("{{ .Version }}\n")
	root.SetHelpCommand(&cobra.Command{Hidden: true})

	root.Flags().SortFlags = false
	root.CompletionOptions.DisableDefaultCmd = true
	cobra.EnableCommandSorting = false

	root.AddCommand(
		Pack(),
		Unpack(),
	)

	if err := root.Execute(); err != nil {
		return err //nolint:wrapcheck	// Error does not need additional wrapping.
	}

	return nil
}
