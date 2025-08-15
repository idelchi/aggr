package cli

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"gitlab.garfield-labs.com/apps/aggr/internal/config"
	"gitlab.garfield-labs.com/apps/aggr/internal/packer"
)

// Pack creates and returns the pack command for aggregating files.
// The pack command collects files from specified patterns or paths and
// aggregates them into a single output file.
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
	cmd.Flags().StringVarP(&configuration.Output, "output", "o", "pack.aggr", "Specify output file (default: stdout)")
	cmd.Flags().
		StringSliceVarP(&configuration.Rules.Patterns, "ignore", "i", []string{}, "Additional .aggignore patterns.")
	cmd.Flags().
		StringVarP(&configuration.Rules.Size, "size", "s", config.DefaultMaxSize, "Maximum size of file to include")
	cmd.Flags().
		IntVarP(&configuration.Rules.Max, "max", "m", config.DefaultMaxFiles, "Maximum number of files to include")
	cmd.Flags().
		BoolVarP(&configuration.DryRun, "dry-run", "d", false, "Show which files would be processed without reading contents")
	cmd.Flags().
		BoolVarP(&configuration.Rules.Hidden, "hidden", "a", false, "Include hidden files and directories")
	cmd.Flags().
		StringSliceVarP(&configuration.Rules.Extensions, "extensions", "x", []string{}, "File extensions to include")
	cmd.Flags().
		StringVarP(&configuration.Rules.Root, "root", "C", ".", "Root directory to use")
	cmd.Flags().
		BoolVarP(&configuration.Rules.Binary, "binary", "b", false, "Include binary files")
	return cmd
}
