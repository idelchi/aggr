// Package checkers provides file filtering and validation functionality.
//
// This package defines a Checker interface and various implementations
// for filtering files during the aggregation process. Checkers can validate
// files based on different criteria such as size limits, ignore patterns,
// binary file detection, and duplicate detection.
//
// The package includes the following checker types:
//   - Binary: Filters out binary files
//   - Ignore: Applies gitignore-style patterns
//   - Seen: Prevents duplicate file inclusion
//   - Size: Enforces file size limits
package checkers
