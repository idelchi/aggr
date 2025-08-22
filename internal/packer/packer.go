package packer

import (
	"fmt"
	"strings"

	"github.com/idelchi/aggr/internal/config"
	"github.com/idelchi/godyl/pkg/path/folder"
)

// Packer orchestrates the file packing and unpacking processes.
type Packer struct {
	// Options contains the configuration settings for the packer.
	Options config.Options
}

// PromptForFolderExists prompts the user for confirmation if the target folder already exists.
// It returns true if the user confirms to proceed, false otherwise.
//
//nolint:forbidigo	// Function prints out to the console.
func PromptForFolderExists(folder folder.Folder) bool {
	if !folder.Exists() {
		return true
	}

	fmt.Printf("The folder %q already exists.\n", folder)
	fmt.Println("This may overwrite existing files. Proceed with caution.")
	fmt.Print("Continue? (y/N): ")

	var response string

	_, _ = fmt.Scanln(&response)

	return strings.ToLower(response) == "y"
}
