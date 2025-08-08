package checkers

import (
	"fmt"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"

	"github.com/idelchi/godyl/pkg/path/file"
)

type Ignore struct {
	ignore *ignore.GitIgnore
}

func NewIgnore(ignore *ignore.GitIgnore) *Ignore {
	return &Ignore{ignore: ignore}
}

func (i *Ignore) check(path string) error {
	matchPath := path
	for strings.HasPrefix(matchPath, "../") {
		matchPath = strings.TrimPrefix(matchPath, "../")
	}

	if ok, pattern := i.ignore.MatchesPathHow(matchPath); ok {
		return fmt.Errorf("%w: in ignore patterns %q", ErrSkip, pattern.Pattern)
	}

	return nil // not a file to ignore, continue processing
}

// Check returns true if the given file is detected as a binary file, false otherwise.
func (i *Ignore) Check(path string) error {
	if err := i.check(path); err != nil {
		return err
	}

	// // TODO(Idelchi): Do we really need this?
	if file.New(path).IsDir() {
		path = strings.TrimRight(path, "/") + "/"
	}

	return i.check(path)
}
