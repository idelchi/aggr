package checkers

import (
	"fmt"

	"github.com/idelchi/godyl/pkg/path/file"
)

type Binary struct{}

func NewBinary() *Binary {
	return &Binary{}
}

// Check returns an error if the given file is binary, nil otherwise.
func (b *Binary) Check(path string) error {
	binary := file.New(path).IsBinaryLike()
	if binary {
		return fmt.Errorf("%w: detected as binary", ErrSkip)
	}

	return nil
}
