package cli

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"gitlab.garfield-labs.com/apps/aggr/internal/config"
	"gitlab.garfield-labs.com/apps/aggr/internal/packer"
)

// Pack creates the pack command.
func Pack() *cobra.Command {
	var configuration config.Options

	cmd := &cobra.Command{
		Use:   "pack [patterns|paths...]",
		Short: "Aggregate files into a single archive",
		Long: heredoc.Doc(`
			Walks through the provided paths/patterns (or the current directory if none are given)
			and concatenates all found files into a single output.
		`),
		Aliases: []string{"p"},
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = []string{"."} // Default to current directory if no args provided
			}
			packer := packer.Packer{
				Options: configuration,
			}

			return packer.Pack(args)
		},
	}

	// Setup flags for the pack command
	cmd.Flags().StringVarP(&configuration.Output, "output", "o", "pack.agg", "Specify output file (default: stdout)")
	cmd.Flags().
		StringSliceVarP(&configuration.Rules.Patterns, "ignore", "i", []string{}, "Additional .aggignore patterns.")
	cmd.Flags().
		StringVar(&configuration.Rules.Size, "size", config.DefaultMaxSize, "Maximum size of file to include")
	cmd.Flags().
		IntVar(&configuration.Rules.Max, "max", config.DefaultMaxFiles, "Maximum number of files to include")
	cmd.Flags().
		BoolVar(&configuration.DryRun, "dry", false, "Show which files would be processed without reading contents")
	cmd.Flags().
		BoolVarP(&configuration.Rules.Hidden, "hidden", "a", false, "Include hidden files and directories")
	cmd.Flags().
		StringSliceVarP(&configuration.Extensions, "ext", "e", []string{}, "File extensions to include")

	return cmd
}
