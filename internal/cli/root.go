package cli

import (
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"github.com/idelchi/aggr/internal/config"
	"github.com/idelchi/aggr/internal/packer"
)

// Execute runs the root command for the aggr CLI application.
//
//nolint:funlen // This function is long due to command and flag definitions.
func Execute(version string) error {
	var configuration config.Options

	root := &cobra.Command{
		Use:   config.Name,
		Short: "Aggregate and unpack files",
		Long: heredoc.Doc(`
			aggr is a command-line utility that recursively aggregates files
			from specified paths into a single file and unpacks them back to their
			original directory structure.

			Walks through the provided paths/patterns (or the current directory if none are given)
			and concatenates all found files into a single output.

			In '--unpack' mode, reads an aggregated file and recreates the original files and directories.
			The command extracts all files from the archive and restores them to their
			original relative paths within the specified output directory.
		`),
		Example: heredoc.Doc(`
			# Pack all files in the current directory and all subdirectories
			aggr -o pack.aggr

			# Unpack the contents of the archive
			aggr -u -o __extracted__ pack.aggr
		`),
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args: func(cmd *cobra.Command, args []string) error {
			if configuration.Unpack {
				if err := cobra.ExactArgs(1)(cmd, args); err != nil {
					return fmt.Errorf(
						"when unpacking, exactly one file argument is required, received %d arguments: %v",
						len(args),
						args,
					)
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration.Rules.IgnoreFile.Set = cmd.Flags().Lookup("ignore-file").Changed

			packer := packer.Packer{
				Options: configuration,
			}

			if configuration.Unpack {
				if !cmd.Flags().Lookup("output").Changed {
					packer.Options.Output = ""
				}

				return packer.Unpack(args)
			}

			// Default to current directory if no args provided
			if len(args) == 0 {
				args = []string{config.DefaultPattern}
			}

			return packer.Pack(args)
		},
	}

	root.SetVersionTemplate("{{ .Version }}\n")
	root.SetHelpCommand(&cobra.Command{Hidden: true})

	root.Flags().SortFlags = false
	root.CompletionOptions.DisableDefaultCmd = true
	cobra.EnableCommandSorting = false

	// Core operation
	root.Flags().BoolVarP(&configuration.Unpack, "unpack", "u", false, "Unpack from a packed file")
	root.Flags().StringVarP(&configuration.Output, "output", "o", "pack.aggr",
		"Specify output file/folder. For --unpack, defaults to './aggr-[hash of <file>]")

	// What to include/exclude
	root.Flags().StringVarP(&configuration.Rules.Root, "root", "C", ".", "Root directory to use")
	root.Flags().StringVarP(&configuration.Rules.IgnoreFile.Path, "ignore-file", "f", "",
		"Path to the .aggrignore file. Set to an empty string to completely ignore. When not passed, uses defaults")
	root.Flags().
		StringSliceVarP(&configuration.Rules.Extensions, "extensions", "x", []string{}, "File extensions to include")
	root.Flags().
		StringSliceVarP(&configuration.Rules.Patterns, "ignore", "i", []string{}, "Additional .aggrignore patterns")
	root.Flags().BoolVarP(&configuration.Rules.Hidden, "hidden", "a", false, "Include hidden files and directories")
	root.Flags().BoolVarP(&configuration.Rules.Binary, "binary", "b", false, "Include binary files")

	// Limits
	root.Flags().StringVarP(&configuration.Rules.Size, "size", "s", config.DefaultMaxSize,
		"Max file size to include (e.g., `500kb`, `1mb`)")
	root.Flags().
		IntVarP(&configuration.Rules.Max, "max", "m", config.DefaultMaxFiles, "Maximum number of files to include")

	// Behavior
	root.Flags().
		BoolVarP(&configuration.Dry, "dry", "d", false, "Show which files would be processed without reading contents")

	defaultWorkers := 4 * runtime.NumCPU() //nolint:mnd	// 4xCPUs
	root.Flags().
		IntVarP(&configuration.Parallel, "parallel", "j", defaultWorkers, "Number of parallel workers to use")

	options := []fang.Option{
		fang.WithVersion(version),
		fang.WithoutManpage(),
		fang.WithoutCompletions(),
		fang.WithErrorHandler(func(_ io.Writer, _ fang.Styles, _ error) {}),
	}

	//nolint:wrapcheck	// Error does not need additional wrapping.
	return fang.Execute(context.Background(), root, options...)
}
