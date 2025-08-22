package checkers

import (
	"errors"
)

// Checker defines the interface for file validation and filtering.
type Checker interface {
	// Check validates a file path and returns an error if the file should be excluded.
	Check(path string) error
}

// Checkers is a collection of Checker instances that can be applied sequentially.
type Checkers []Checker

// Check applies all checkers in the collection to the given path.
// It returns the first error encountered, or nil if all checks pass.
func (c Checkers) Check(path string) error {
	for _, checker := range c {
		if err := checker.Check(path); err != nil {
			return err
		}
	}

	return nil
}

var (
	// ErrSkip indicates that a file should be skipped.
	ErrSkip = errors.New("skipping")
	// ErrPrune indicates that a directory should be pruned.
	ErrPrune = errors.New("pruning directory")
	// ErrAbort indicates that the checking process should be aborted.
	ErrAbort = errors.New("aborting")
)
