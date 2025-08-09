package checkers

import (
	"fmt"

	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"
)

// Seen is a checker that prevents duplicate file inclusion by tracking processed files.
type Seen struct {
	// Files is a reference to the collection of files that have been processed.
	Files *files.Files
}

// NewSeen creates a new Seen checker with a reference to the files collection.
func NewSeen(files *files.Files) *Seen {
	return &Seen{Files: files}
}

// Check returns an error if the file has already been included in the collection.
func (s *Seen) Check(path string) error {
	if s.Files.Contains(file.New(path)) {
		return fmt.Errorf("%w: already included", ErrSkip)
	}
	return nil
}
