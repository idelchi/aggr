// Package packer provides functionality for packing files into aggregated streams and unpacking them.
//
// This package implements the core file aggregation and extraction functionality.
// It can pack multiple files into a single text-based stream with special markers,
// and unpack such streams back into their original file structure. The packing
// format uses begin/end markers to delimit individual files within the stream.
//
// Key features:
//   - Concurrent processing using worker pools
//   - Support for dry-run operations
//   - File filtering through checker interfaces
//   - Tree visualization of packed files
//   - Path prefix stripping for cleaner archives
package packer
