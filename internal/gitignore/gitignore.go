package gitignore

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type pattern struct {
	original string
	pattern  string
	negated  bool
	dirOnly  bool
	rooted   bool
}

type GitIgnore struct {
	patterns []pattern
}

// New creates a GitIgnore from lines of a .gitignore file
func New(lines []string) *GitIgnore {
	g := &GitIgnore{
		patterns: make([]pattern, 0),
	}

	for _, line := range lines {
		p := parsePattern(line)
		if p != nil {
			g.patterns = append(g.patterns, *p)
		}
	}

	return g
}

func parsePattern(line string) *pattern {
	// Blank lines are ignored
	if len(line) == 0 {
		return nil
	}

	// Comments start with # (unless escaped)
	if strings.HasPrefix(line, "#") {
		return nil
	}

	p := &pattern{
		original: line,
	}

	// Handle escaped characters at the beginning
	if strings.HasPrefix(line, "\\!") {
		line = line[1:] // Remove the backslash, keep the !
	} else if strings.HasPrefix(line, "\\#") {
		line = line[1:] // Remove the backslash, keep the #
	} else if strings.HasPrefix(line, "!") {
		// Negation pattern
		p.negated = true
		line = line[1:]
	}

	// Trim trailing spaces unless escaped
	line = trimTrailingSpaces(line)

	// Empty pattern after trimming
	if len(line) == 0 {
		return nil
	}

	// Check if pattern matches directories only (trailing /)
	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	// Check if pattern is rooted (starts with /)
	if strings.HasPrefix(line, "/") {
		p.rooted = true
		line = strings.TrimPrefix(line, "/")
	}

	p.pattern = line
	return p
}

func trimTrailingSpaces(s string) string {
	// Count trailing spaces that are not escaped
	end := len(s)
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != ' ' {
			break
		}
		// Check if this space is escaped
		if i > 0 && s[i-1] == '\\' {
			// This space is escaped, stop here
			break
		}
		end = i
	}

	result := s[:end]

	// Handle escaped spaces - remove the backslash
	result = strings.ReplaceAll(result, `\ `, " ")

	return result
}

// IsIgnored checks if a path should be ignored
func (g *GitIgnore) IsIgnored(path string, isDir bool) bool {
	// Special cases
	if path == "" || path == "." {
		return false
	}

	// Clean the path
	path = filepath.Clean(path)
	path = filepath.ToSlash(path) // Ensure forward slashes

	ignored := false
	
	// Track which directories are permanently excluded
	// Once a directory is excluded, its contents can NEVER be re-included
	excludedDirs := make(map[string]bool)
	
	// Build list of all parent paths to check
	var pathsToCheck []string
	parts := strings.Split(path, "/")
	for i := 1; i <= len(parts); i++ {
		checkPath := strings.Join(parts[:i], "/")
		pathsToCheck = append(pathsToCheck, checkPath)
	}
	
	// First pass: determine which directories are excluded
	// Check ALL patterns (not just dirOnly) as they can exclude directories
	for _, p := range g.patterns {
		for _, checkPath := range pathsToCheck {
			// Only check parent directories, not the target path itself
			if checkPath == path {
				continue
			}
			
			// Check if this pattern EXPLICITLY excludes the directory itself
			// The parent exclusion rule only applies when a directory is explicitly excluded,
			// not when wildcard patterns match files inside it
			dirMatches := false
			if p.dirOnly {
				// Directory-only patterns (ending with /) explicitly exclude directories
				dirMatches = matchesDirectoryPath(p, checkPath)
			} else {
				// Regular patterns can exclude directories if they match the directory name
				// Special handling for wildcard patterns
				if p.pattern == "*" {
					// "*" pattern can exclude directories by matching their basename
					basename := filepath.Base(checkPath)
					dirMatches = matchGlob(p, basename)
				} else if strings.Contains(p.pattern, "**") {
					// "**" patterns can match directories at any level
					dirMatches = matchGlob(p, checkPath)
				} else if strings.Contains(p.pattern, "/") {
					// Pattern with slash - only matches if it specifically matches directory path
					dirMatches = matchGlob(p, checkPath)
				} else {
					// Pattern without slash - check if it matches directory basename
					// e.g., "build" pattern excludes directory named "build"
					basename := filepath.Base(checkPath)
					dirMatches = matchGlob(p, basename)
				}
			}
			
			if dirMatches {
				if p.negated {
					// Negated patterns CAN un-exclude directories for parent exclusion in specific cases:
					// 1. Wildcard directory patterns like !*/, !**/
					// 2. Explicit directory patterns like !/foo
					if (p.dirOnly && strings.Contains(p.pattern, "*")) || 
					   (!p.dirOnly && !strings.Contains(p.pattern, "*")) {
						// Allow un-exclusion for wildcard dir patterns and explicit non-wildcard patterns
						delete(excludedDirs, checkPath)
					}
					// But NOT for patterns like !build/ (explicit directory patterns)
				} else {
					// Directory is excluded - mark it permanently
					excludedDirs[checkPath] = true
				}
			}
		}
	}
	
	// Check if any parent directory is excluded (implements parent exclusion rule)
	// This applies to both files AND directories
	parentExcluded := false
	if len(parts) > 1 { // Has parent directories
		for i := 1; i < len(parts); i++ {
			parentPath := strings.Join(parts[:i], "/")
			if excludedDirs[parentPath] {
				parentExcluded = true
				break
			}
		}
	}
	
	// Second pass: apply patterns to the target path
	for _, p := range g.patterns {
		if matches(p, path, isDir) {
			if p.negated {
				// Git rule: cannot re-include file if parent directory is excluded
				if !parentExcluded {
					ignored = false
				}
				// Note: even directories cannot be re-included if parent is excluded
			} else {
				ignored = true
			}
		}
	}

	return ignored
}

func matches(p pattern, path string, isDir bool) bool {
	// Special handling for directory-only patterns
	if p.dirOnly {
		// Directory patterns can match:
		// 1. The directory itself
		// 2. Files inside the directory
		return matchesDirectoryPattern(p, path, isDir)
	}

	// Regular patterns
	return matchesFilePattern(p, path, isDir)
}

func matchesDirectoryPattern(p pattern, path string, isDir bool) bool {
	// Directory patterns work differently for positive and negative patterns:
	// - Positive patterns (build/) match the directory AND files inside it
	// - Negative patterns have special cases:
	//   - Wildcard patterns like !*/ or !**/ only match directories
	//   - Simple patterns like !build/ match directory and files for re-inclusion

	if isDir {
		// Check if this directory matches the pattern directly
		if matchesDirectoryPath(p, path) {
			return true
		}
		
		// For positive patterns, also check if directory is inside a matching directory
		if !p.negated {
			parts := strings.Split(path, "/")
			for i := 1; i < len(parts); i++ {
				parentPath := strings.Join(parts[:i], "/")
				if matchesDirectoryPath(p, parentPath) {
					return true
				}
			}
		}
		
		return false
	}

	// For files: 
	if p.negated {
		// CRITICAL: Negated directory patterns (!build/) should NEVER match files
		// They only match the directory itself for the purpose of un-ignoring the directory
		// but NOT for re-including files inside the directory
		return false
	}
	
	// For positive directory patterns, check if file is inside a matching directory
	parts := strings.Split(path, "/")
	for i := 1; i < len(parts); i++ {
		parentPath := strings.Join(parts[:i], "/")
		if matchesDirectoryPath(p, parentPath) {
			return true
		}
	}

	return false
}

func matchesDirectoryPath(p pattern, path string) bool {
	if p.rooted {
		// Rooted patterns match only from the repository root
		return matchGlob(p, path)
	}

	// Non-rooted patterns can match at any level
	// Try matching the full path
	if matchGlob(p, path) {
		return true
	}

	// For patterns without slash, also try matching just the basename
	if !strings.Contains(p.pattern, "/") {
		basename := filepath.Base(path)
		if matchGlob(p, basename) {
			return true
		}
	} else {
		// For patterns with slash, try matching at each directory level
		parts := strings.Split(path, "/")
		for i := 0; i < len(parts); i++ {
			subpath := strings.Join(parts[i:], "/")
			if matchGlob(p, subpath) {
				return true
			}
		}
	}

	return false
}

func matchesFilePattern(p pattern, path string, isDir bool) bool {
	// Special case: * pattern matches everything except paths with slashes
	if p.pattern == "*" {
		return true
	}

	if p.rooted {
		// Rooted patterns match only from the repository root
		return matchGlob(p, path)
	}

	// Non-rooted patterns can match at any level

	// Try matching the full path
	if matchGlob(p, path) {
		return true
	}

	// For patterns without slash, also try matching just the basename
	// AND check if file is inside a directory that matches the pattern
	if !strings.Contains(p.pattern, "/") {
		basename := filepath.Base(path)
		if matchGlob(p, basename) {
			return true
		}

		// Also check if any parent directory matches the pattern
		parts := strings.Split(path, "/")
		for i := 1; i < len(parts); i++ {
			parentPath := strings.Join(parts[:i], "/")
			parentBasename := filepath.Base(parentPath)
			if matchGlob(p, parentBasename) {
				return true
			}
		}
	} else {
		// For patterns with slash, try matching at each directory level
		parts := strings.Split(path, "/")
		for i := 0; i < len(parts); i++ {
			subpath := strings.Join(parts[i:], "/")
			if matchGlob(p, subpath) {
				return true
			}
		}
	}

	return false
}

func matchGlob(p pattern, path string) bool {
	// Use doublestar for glob matching
	matched, _ := doublestar.Match(p.pattern, path)
	return matched
}

// Oneliner to just test.
func Ignore(patterns []string, path string, isDir bool) bool {
	g := New(patterns)
	return g.IsIgnored(path, isDir)
}
