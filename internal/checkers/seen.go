package checkers

import (
	"fmt"

	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"
)

type Seen struct {
	Files *files.Files
}

func NewSeen(files *files.Files) *Seen {
	return &Seen{Files: files}
}

// Check returns true if the given file has already been seen, false otherwise.
func (s *Seen) Check(path string) error {
	if s.Files.Contains(file.New(path)) {
		return fmt.Errorf("%w: already included", ErrSkip)
	}
	return nil
}
