package packer

import (
	"errors"
	"fmt"

	"github.com/idelchi/aggr/internal/checkers"
	"github.com/idelchi/aggr/internal/patterns"
	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/folder"
)

// Unpack extracts files from an aggregated file and recreates the original directory structure.
// It reads the packed file from the given path and writes the extracted files to the
// configured output directory.
func (p Packer) Unpack(packs []string) error {
	path := packs[0] // Expecting a single file path for unpacking

	log, err := Logger(p.Options.Dry)
	if err != nil {
		return err
	}

	// Read the packed file
	archive := file.New(path)

	// Create unpacker instance
	unpacker := NewAggregator(log, p.Options.Dry, p.Options.Parallel, p.Options.Rules.Root)

	ignorePatterns := patterns.Patterns(p.Options.Rules.Patterns)

	if len(ignorePatterns) > 0 {
		log.Debug("- Using ignore patterns:")

		for _, pattern := range ignorePatterns {
			log.Debugf("  - %s", pattern)
		}
	}

	if len(p.Options.Rules.Extensions) > 0 {
		extras := patterns.Patterns{"*", "!*/"}
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

		output = folder.New(fmt.Sprintf("%s-%s", archive.Base(), hash))
	}

	// if output exists as a directory, prompt the user
	if !PromptForFolderExists(output) {
		return errors.New("aborted unpacking")
	}

	// Unpack the files
	files, err := unpacker.Unpack(archive, output.Path(), checkers)
	if err != nil {
		return fmt.Errorf("unpacking files: %w", err)
	}

	if len(files) == 0 {
		log.Warn("No files found matching the specified patterns and rules")

		return nil
	}

	if p.Options.Dry {
		log.Info("Unpacking files:")

		for _, f := range files {
			log.Debugf("- %q", f)
		}

		return nil
	}

	log.Infof("Successfully unpacked %d files from %q to %q", len(files), archive, output)

	return nil
}
