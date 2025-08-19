package checkers

import (
	"fmt"

	"github.com/dustin/go-humanize"

	"github.com/idelchi/godyl/pkg/path/file"
)

// Size is a checker that filters files based on their size.
type Size struct {
	// Size is the maximum allowed file size in bytes.
	Size int
}

// NewSize creates a new Size checker with the specified size limit in bytes.
func NewSize(size int) *Size {
	return &Size{Size: size}
}

// Check returns an error if the file is larger than the configured size limit.
func (s *Size) Check(path string) error {
	if !file.New(path).IsFile() {
		return nil // Directories are not considered
	}

	file := file.New(path)
	if !file.IsFile() {
		return nil
	}

	if ok, err := file.LargerThan(int64(s.Size)); err == nil && ok {
		//nolint:gosec 	// File size from os.FileInfo cannot be negative.
		return fmt.Errorf("%w: larger than requested max size %s", ErrSkip, humanize.Bytes(uint64(s.Size)))
	}

	return nil
}
