// Package cli provides command-line interface functionality for the aggr tool.
//
// This package implements the Cobra-based CLI commands for the aggr file aggregation tool.
// It provides two main commands:
//   - pack: Aggregates files from specified directories into a single output file
//   - unpack: Extracts files from an aggregated file back to their original structure
//
// The CLI supports various options including file filtering by size, patterns, and
// extensions, as well as dry-run modes for testing operations.
package cli