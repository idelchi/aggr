package checkers

import (
	"fmt"

	"github.com/dustin/go-humanize"

	"github.com/idelchi/godyl/pkg/path/file"
)

type Size struct {
	// Size is the size in bytes to check against.
	Size int
}

func NewSize(size int) *Size {
	return &Size{Size: size}
}

// Check returns true if the given file is larger than the specified size, false otherwise.
func (s *Size) Check(path string) error {
	if !file.New("", path).IsFile() {
		return nil // Directories are not considered
	}

	file := file.New("", path)
	if !file.IsFile() {
		return nil
	}

	if ok, err := file.LargerThan(int64(s.Size)); err != nil {
		return nil
	} else if ok {
		return fmt.Errorf("%w: larger than requested max size %s", ErrSkip, humanize.Bytes(uint64(s.Size)))
	}

	return nil
}
