package tree

import (
	"strings"

	"github.com/xlab/treeprint"

	"github.com/idelchi/godyl/pkg/path/files"
)

// Generate creates a visual tree structure from a list of file paths.
// It returns a treeprint.Tree that can be rendered as ASCII art showing
// the hierarchical organization of the provided files.
func Generate(fileList files.Files) treeprint.Tree {
	paths := fileList.AsSlice()
	root := treeprint.New()
	branches := map[string]treeprint.Tree{
		"": root, // key = joined path parts without leading slash
	}

	// Build the tree structure
	for _, p := range paths {
		addPathToTree(p, root, branches)
	}

	return root
}

// addPathToTree adds a single path to the tree structure.
func addPathToTree(path string, root treeprint.Tree, branches map[string]treeprint.Tree) {
	parts := strings.Split(path, "/")
	var keyBuilder []string
	currentTree := root

	for i, part := range parts {
		keyBuilder = append(keyBuilder, part)
		key := strings.Join(keyBuilder, "/")
		isLastPart := i == len(parts)-1

		// Check if this branch already exists
		if existingBranch, exists := branches[key]; exists {
			currentTree = existingBranch
			continue
		}

		// Create new branch or leaf node
		if isLastPart {
			// This is a file (leaf node)
			currentTree.AddNode(part)
		} else {
			// This is a directory (branch)
			newBranch := currentTree.AddBranch(part)
			branches[key] = newBranch
			currentTree = newBranch
		}
	}
}
