package cli

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"gitlab.garfield-labs.com/apps/aggr/internal/config"
	"gitlab.garfield-labs.com/apps/aggr/internal/packer"
)

// Unpack creates and returns the unpack command for extracting aggregated files.
// The unpack command reads an aggregated file and recreates the original files
// and directories in their original structure.
func Unpack() *cobra.Command {
	var configuration config.Options

	cmd := &cobra.Command{
		Use:   "unpack <file>",
		Short: "Unpack an aggregated file into its original directory structure",
		Long: heredoc.Doc(`
			Reads an aggregated file and recreates the original files and directories.
			The command extracts all files from the archive and restores them to their
			original relative paths within the specified output directory.
		`),
		Example: heredoc.Doc(`
			# Unpack to current directory
			$ agg unpack archive.aggr

			# Unpack to extracted/ directory
			$ agg unpack -o extracted/ archive.aggr
		`),
		Aliases: []string{"u", "x"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = []string{"."} // Default to current directory if no args provided
			}
			packer := packer.Packer{
				Options: configuration,
			}

			return packer.Unpack(args[0])
		},
	}

	// Setup flags for the unpack command
	cmd.Flags().
		StringVarP(&configuration.Output, "output", "o", "", "Specify output directory. Defaults to '$(pwd)/aggr-<hash of <file>'")
	cmd.Flags().
		StringSliceVarP(&configuration.Rules.Patterns, "ignore", "i", []string{}, "Additional .aggignore patterns.")
	cmd.Flags().
		BoolVar(&configuration.DryRun, "dry-run", false, "Show which files would be processed without reading contents")
	cmd.Flags().
		StringSliceVarP(&configuration.Rules.Extensions, "extensions", "x", []string{}, "File extensions to include")

	return cmd
}
