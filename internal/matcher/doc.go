// Package matcher provides file matching and collection functionality.
//
// This package implements the Globber type which coordinates file discovery
// and filtering using various checker implementations. It uses the walker
// package to traverse file systems and applies a series of checks to
// determine which files should be included in the final collection.
//
// The matcher supports concurrent file processing and maintains a collection
// of matched files that can be processed further by other components.
package matcher