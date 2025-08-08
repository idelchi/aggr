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

	"gitlab.garfield-labs.com/apps/aggr/internal/checkers"
	"gitlab.garfield-labs.com/apps/aggr/internal/config"
	"gitlab.garfield-labs.com/apps/aggr/internal/matcher"
	"gitlab.garfield-labs.com/apps/aggr/internal/patterns"
)

// Packer orchestrates the file packing process using various internal packages.
type Packer struct {
	Options config.Options
	files   files.Files
}

// Pack executes the pack command.
func (p Packer) Pack(paths []string) error {
	log, err := Logger(p.Options.DryRun)
	if err != nil {
		return err
	}

	bytes, err := humanize.ParseBytes(p.Options.Rules.Size)
	if err != nil {
		return fmt.Errorf("parsing size value %q: %w", p.Options.Rules.Size, err)
	}

	ignorePatterns := patterns.Patterns{}

	if !p.Options.Rules.Hidden {
		log.Debug("- Adding default ignore patterns for hidden files and directories")
		// Exclude hidden folders & files if hidden is false
		ignorePatterns = append(ignorePatterns, config.DefaultHidden...)
	}

	if exe, err := os.Executable(); err == nil {
		path := file.New("", exe).Path()
		log.Debugf("- Excluding executable itself: %q", path)

		// Exclude the executable itself
		ignorePatterns = append(ignorePatterns, path)
	}

	log.Debug("- Adding default ignore patterns for executables and known directories")

	ignorePatterns = append(ignorePatterns, config.DefaultExcludes...)

	// Add output file to excludes if specified
	if !p.Options.IsStdout() {
		log.Debugf("- Excluding output file from aggregation: %q", p.Options.Output)
		ignorePatterns = append(ignorePatterns, p.Options.Output)
	}

	ignorePatterns = append(ignorePatterns, p.Options.Rules.Patterns...)

	if len(p.Options.Extensions) > 0 {
		extras := patterns.Patterns{"*"}
		log.Debugf("- Adding file extension patterns: %v", extras)
		extras = append(extras, ExtensionsToPatterns(p.Options.Extensions)...)
		ignorePatterns = append(ignorePatterns, extras...)
	}

	ignorer := ignorePatterns.AsGitIgnore()

	aggrignore, ok := ActiveAggrignore()
	if ok {
		log.Debugf("- Active .aggignore file: %q", aggrignore)
		ignorer, err = patterns.LoadIgnoreFile(aggrignore, ignorePatterns)
		if err != nil {
			return fmt.Errorf("failed to load ignore patterns: %w", err)
		}
	} else {
		log.Debug("- No active .aggignore file found")
	}

	if len(ignorePatterns) > 0 {
		log.Debug("- Using ignore patterns:")
		for _, pattern := range ignorePatterns {
			log.Debugf("  - %s", pattern)
		}
	}

	checkers := []checkers.Checker{
		checkers.NewIgnore(ignorer),
		checkers.NewSize(int(bytes)),
		checkers.NewBinary(),
	}

	m := matcher.New(checkers, p.Options.Rules.Max, log)

	for _, path := range paths {
		log.Debugf("\n- Processing pattern: %v", path)
		if err := m.Match(patterns.Normalize(path)); err != nil {
			return fmt.Errorf("matching pattern %q: %w", path, err)
		}
	}

	slices.SortFunc(m.Files, func(a, b file.File) int {
		return strings.Compare(strings.ToLower(a.Path()), strings.ToLower(b.Path()))
	})

	if p.Options.DryRun {
		p.Options.Output = "" // In dry run mode, we don't write anything
	}

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

	aggregator := NewAggregator(log, p.Options.DryRun, p.Options.Parallel)

	if err := aggregator.Pack(m.Files, writer); err != nil {
		return fmt.Errorf("failed to aggregate files: %w", err)
	}

	// Show completion message
	if !p.Options.IsStdout() {
		log.Infof("Successfully packed %d files into %s", len(m.Files), p.Options.Output)
	}

	return nil
}

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

func (p Packer) Unpack(path string) error {
	log, err := Logger(p.Options.DryRun)
	if err != nil {
		return err
	}

	// Read the packed file
	archive := file.New("", path)

	// Create unpacker instance
	unpacker := NewAggregator(log, p.Options.DryRun, p.Options.Parallel)

	ignorePatterns := patterns.Patterns(p.Options.Rules.Patterns)

	if len(ignorePatterns) > 0 {
		log.Debug("- Using ignore patterns:")
		for _, pattern := range ignorePatterns {
			log.Debugf("  - %s", pattern)
		}
	}

	checkers := []checkers.Checker{
		checkers.NewIgnore(ignorePatterns.AsGitIgnore()),
	}

	output := folder.New("", p.Options.Output)

	if p.Options.Output == "" {
		hash, err := archive.Hash()
		if err != nil {
			return fmt.Errorf("calculating archive hash: %w", err)
		}

		output = folder.New("", fmt.Sprintf("aggr-%s", hash))
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
