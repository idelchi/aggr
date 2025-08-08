package checkers

import (
	"fmt"

	"github.com/idelchi/godyl/pkg/path/file"
)

// Binary is a checker that filters out binary files.
type Binary struct{}

// NewBinary creates a new Binary checker.
func NewBinary() *Binary {
	return &Binary{}
}

// Check returns an error if the file is detected as binary content.
func (b *Binary) Check(path string) error {
	binary := file.New(path).IsBinaryLike()
	if binary {
		return fmt.Errorf("%w: detected as binary", ErrSkip)
	}

	return nil
}
