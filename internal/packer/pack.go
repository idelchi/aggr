package packer

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/idelchi/aggr/internal/checkers"
	"github.com/idelchi/aggr/internal/config"
	"github.com/idelchi/aggr/internal/patterns"
	"github.com/idelchi/aggr/internal/walker"
	gitignore "github.com/idelchi/go-gitignore"
	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/pretty"
)

// Pack aggregates files matching the given search patterns into a single output.
// It processes the patterns, applies filtering rules, and writes the aggregated
// result to the configured output destination.
//
//nolint:gocognit,funlen	// TODO(Idelchi): Refactor this function to reduce complexity.
func (p Packer) Pack(searchPatterns []string) error {
	log, err := Logger(p.Options.DryRun)
	if err != nil {
		return err
	}

	log.Debug("Packing files with options:")

	if p.Options.DryRun {
		pretty.PrintYAML(p.Options)

		//nolint:forbidigo	// Function prints out to the console.
		fmt.Printf("args: %v\n", searchPatterns)
	}

	search := patterns.Patterns(searchPatterns)

	if err := search.Validate(); err != nil {
		return fmt.Errorf(
			"validating search patterns: %w:\nuse --root/-C <path> to specify a different root directory",
			err,
		)
	}

	search = search.Normalized(p.Options.Rules.Root)

	log.Debugf("- Normalized search patterns: %v", search)

	bytes, err := humanize.ParseBytes(p.Options.Rules.Size)
	if err != nil {
		return fmt.Errorf("parsing size value %q: %w", p.Options.Rules.Size, err)
	}

	ignorePatterns := patterns.Patterns{}

	log.Debug("- Adding ignore patterns:")

	var aggrignore file.File

	if !p.Options.Rules.IgnoreFile.Set || p.Options.Rules.IgnoreFile.Path != "" {
		if !p.Options.Rules.IgnoreFile.Set {
			aggrignore = DefaultAggrignores()
		} else {
			aggrignore = file.New(p.Options.Rules.IgnoreFile.Path)

			if !aggrignore.Exists() {
				return fmt.Errorf("ignore file %q does not exist", aggrignore)
			}
		}
	}

	if aggrignore.Set() {
		lines, err := aggrignore.Lines()
		if err != nil {
			return fmt.Errorf("reading %q: %w", aggrignore, err)
		}

		ignorePatterns = append(ignorePatterns, patterns.Patterns(lines).TrimEmpty()...)

		log.Debugf("  - .aggrignore (from %q): %v", aggrignore, gitignore.New(ignorePatterns).Patterns())
	} else {
		log.Debug("  - .aggrignore: [none loaded]")
	}

	if len(p.Options.Rules.Extensions) > 0 {
		extras := patterns.Patterns{"*", "!*/"}

		extras = append(extras, ExtensionsToPatterns(p.Options.Rules.Extensions)...)
		log.Debugf("  - file extension patterns passed on commandline: %v", extras)

		ignorePatterns = append(ignorePatterns, extras...)
	}

	log.Debugf("  - default: %v", config.DefaultExcludes)

	ignorePatterns = append(ignorePatterns, config.DefaultExcludes...)

	// Exclude the executable itself
	if exe, err := os.Executable(); err == nil {
		path := file.New(exe).Path()
		log.Debugf("  - the executable: %q", path)

		ignorePatterns = append(ignorePatterns, path)
	}

	// Add output file to excludes if specified
	if !p.Options.IsStdout() {
		log.Debugf("  - the output file: %q", p.Options.Output)

		ignorePatterns = append(ignorePatterns, p.Options.Output)
	}

	log.Debugf("  - patterns passed on commandline: %v", p.Options.Rules.Patterns)

	ignorePatterns = append(ignorePatterns, p.Options.Rules.Patterns...)

	// Exclude hidden folders & files if hidden is false
	if !p.Options.Rules.Hidden {
		log.Debugf("  - hidden files and folders: %v", config.DefaultHidden)

		ignorePatterns = append(ignorePatterns, config.DefaultHidden...)
	}

	ignorer := ignorePatterns.AsGitIgnore()

	if len(ignorePatterns) > 0 {
		log.Debug("- The following patterns will be applied:")

		for _, pattern := range ignorePatterns {
			log.Debugf("  - %s", pattern)
		}
	}

	checks := []checkers.Checker{
		checkers.NewIgnore(ignorer),
		//nolint:gosec		// Cannot be negative.
		checkers.NewSize(int(bytes)),
	}

	if !p.Options.Rules.Binary {
		checks = append(checks, checkers.NewBinary())
	}

	walker := walker.New(checks, p.Options.Rules.Max, log)

	for _, path := range search {
		log.Debugf("\n- Processing pattern: %v", path)

		if err := walker.Walk(os.DirFS(p.Options.Rules.Root), path); err != nil {
			return fmt.Errorf("matching pattern %q: %w", path, err)
		}
	}

	files := walker.Files

	slices.SortFunc(files, func(a, b file.File) int {
		return strings.Compare(strings.ToLower(a.Path()), strings.ToLower(b.Path()))
	})

	if p.Options.DryRun {
		p.Options.Output = "" // In dry run mode, we don't write anything
	}

	aggregator := NewAggregator(
		log,
		p.Options.DryRun,
		p.Options.Parallel,
		p.Options.Rules.Root,
	)

	// Get output writer
	writer, err := GetOutputWriter(p.Options)
	if err != nil {
		return err
	}

	defer func() {
		if writer != os.Stdout {
			_ = writer.Close()
		}
	}()

	if err := aggregator.Pack(files, writer); err != nil {
		return fmt.Errorf("failed to aggregate files: %w", err)
	}

	// Show completion message
	if !p.Options.IsStdout() {
		log.Infof("Successfully packed %d files into %s", len(files), p.Options.Output)
	}

	return nil
}
