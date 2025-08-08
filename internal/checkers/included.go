package checkers

import (
	"fmt"

	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"
)

type Included struct {
	Files files.Files
}

func NewIncluded(files files.Files) *Included {
	return &Included{Files: files}
}

// Check returns true if the given file is not contained in the list of matched files.
func (c *Included) Check(path string) error {
	if !c.Files.Contains(file.New("", path)) {
		return fmt.Errorf("%w: not contained", ErrSkip)
	}

	return nil
}
