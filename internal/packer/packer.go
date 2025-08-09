package packer

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"
	"github.com/idelchi/godyl/pkg/path/folder"
	"github.com/idelchi/godyl/pkg/pretty"

	"gitlab.garfield-labs.com/apps/aggr/internal/checkers"
	"gitlab.garfield-labs.com/apps/aggr/internal/config"
	"gitlab.garfield-labs.com/apps/aggr/internal/patterns"
	"gitlab.garfield-labs.com/apps/aggr/internal/walker"
)

// Packer orchestrates the file packing and unpacking processes.
type Packer struct {
	// Options contains the configuration settings for the packer.
	Options config.Options
	// files holds the collection of files being processed.
	files files.Files
}

// Pack aggregates files matching the given search patterns into a single output.
// It processes the patterns, applies filtering rules, and writes the aggregated
// result to the configured output destination.
func (p Packer) Pack(searchPatterns []string) error {
	log, err := Logger(p.Options.DryRun)
	if err != nil {
		return err
	}

	log.Debug("Packing files with options:")
	if p.Options.DryRun {
		pretty.PrintYAML(p.Options)
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

	log.Debug("- Checking for .aggignore file")

	aggrignore, ok := ActiveAggrignore()
	if ok {
		log.Debugf("  - Found .aggignore file: %q", aggrignore)
		lines, err := aggrignore.Lines()
		if err != nil {
			return fmt.Errorf("reading %q: %w", aggrignore, err)
		}

		ignorePatterns = append(ignorePatterns, patterns.Patterns(lines).TrimEmpty()...)
	} else {
		log.Debug("  - No active .aggignore file found")
	}

	log.Debug("- Adding ignore patterns:")
	log.Debugf("  - .aggignore: %v", ignorePatterns)
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

	if len(p.Options.Rules.Extensions) > 0 {
		extras := patterns.Patterns{"*"}
		extras = append(extras, ExtensionsToPatterns(p.Options.Rules.Extensions)...)
		log.Debugf("  - file extension patterns passed on commandline: %v", extras)
		ignorePatterns = append(ignorePatterns, extras...)
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
		checkers.NewSize(int(bytes)),
	}

	if !p.Options.Rules.Binary {
		checks = append(checks, checkers.NewBinary())
	}

	w := walker.New(checks, p.Options.Rules.Max, log)

	for _, path := range search {
		log.Debugf("\n- Processing pattern: %v", path)
		if err := w.Walk(os.DirFS(p.Options.Rules.Root), path); err != nil {
			return fmt.Errorf("matching pattern %q: %w", path, err)
		}
	}

	files := w.Files

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
			writer.Close()
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

// PromptForFolderExists prompts the user for confirmation if the target folder already exists.
// It returns true if the user confirms to proceed, false otherwise.
func PromptForFolderExists(folder folder.Folder) bool {
	if !folder.Exists() {
		return true
	}

	fmt.Printf("The folder %q already exists.\n", folder)
	fmt.Println("This may overwrite existing files. Proceed with caution.")
	fmt.Print("Continue? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) == "y" {
		return true
	}

	return false
}

// Unpack extracts files from an aggregated file and recreates the original directory structure.
// It reads the packed file from the given path and writes the extracted files to the
// configured output directory.
func (p Packer) Unpack(path string) error {
	log, err := Logger(p.Options.DryRun)
	if err != nil {
		return err
	}

	// Read the packed file
	archive := file.New(path)

	// Create unpacker instance
	unpacker := NewAggregator(log, p.Options.DryRun, p.Options.Parallel, p.Options.Rules.Root)

	ignorePatterns := patterns.Patterns(p.Options.Rules.Patterns)

	if len(ignorePatterns) > 0 {
		log.Debug("- Using ignore patterns:")
		for _, pattern := range ignorePatterns {
			log.Debugf("  - %s", pattern)
		}
	}

	if len(p.Options.Rules.Extensions) > 0 {
		extras := patterns.Patterns{"*"}
		log.Debugf("- Adding file extension patterns: %v", extras)
		extras = append(extras, ExtensionsToPatterns(p.Options.Rules.Extensions)...)
		ignorePatterns = append(ignorePatterns, extras...)
	}

	checkers := []checkers.Checker{
		checkers.NewIgnore(ignorePatterns.AsGitIgnore()),
	}

	output := folder.New(p.Options.Output)

	if p.Options.Output == "" {
		hash, err := archive.Hash()
		if err != nil {
			return fmt.Errorf("calculating archive hash: %w", err)
		}

		output = folder.New(fmt.Sprintf("aggr-%s", hash))
	}

	// if output exists as a directory, prompt the user
	if !PromptForFolderExists(output) {
		return fmt.Errorf("aborted unpacking")
	}

	// Unpack the files
	files, err := unpacker.Unpack(archive, output.Path(), checkers)
	if err != nil {
		return fmt.Errorf("unpacking files: %w", err)
	}

	if p.Options.DryRun {
		if len(files) == 0 {
			log.Warn("  - No files would be unpacked in dry run mode")

			return nil
		}

		log.Info("- Unpacking files:")

		for _, f := range files {
			log.Debugf("  - %q", f)
		}
		return nil
	}

	log.Infof("Successfully unpacked %d files from %q to %q", len(files), archive, output)
	return nil
}
