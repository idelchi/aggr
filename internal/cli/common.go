package cli

import (
	"gitlab.garfield-labs.com/apps/aggr/internal/config"
	"gitlab.garfield-labs.com/apps/aggr/internal/packer"
)

// Packer creates a new packer instance with the provided arguments and configuration.
func Packer(args []string, configuration config.Options) packer.Packer {
	// Default to current directory if no args provided
	configuration.Search = []string{"."}

	if len(args) > 0 {
		configuration.Search = args
	}

	return packer.Packer{
		Options: configuration,
	}
}
