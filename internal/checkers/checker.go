package checkers

import (
	"errors"
)

type Checker interface {
	Check(path string) error
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
