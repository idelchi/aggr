package checkers

import (
	"errors"
	"slices"
)

type Checker interface {
	Check(path string) error
}

// Contains returns true if the given file is already present in the list of matched files, false otherwise.
func Contains(file string, files []string) bool {
	return slices.Contains(files, file)
}

type Checkers []Checker

func (c Checkers) Check(path string) error {
	for _, checker := range c {
		if err := checker.Check(path); err != nil {
			return err
		}
	}

	return nil
}

var (
	ErrSkip  = errors.New("skipping")
	ErrAbort = errors.New("aborting")
)
