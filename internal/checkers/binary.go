package checkers

import (
	"fmt"

	"github.com/gabriel-vasile/mimetype"

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

	if !file.New(path).IsFile() {
		return nil // Directories are not considered
	}

	m, err := mimetype.DetectFile(path) // reads only header
	if err != nil {
		return nil // if we can't detect the MIME type, we assume it's not binary
	}

	for p := m; p != nil; p = p.Parent() { // walk MIME hierarchy
		if p.Is("text/plain") { // text/plain sits under every textual type
			return nil
		}
	}

	return fmt.Errorf("%w: detected as binary", ErrSkip)
}
