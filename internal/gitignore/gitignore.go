package gitignore

import (
	pathpkg "path"
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

	// Handle edge case: if pattern becomes empty after trimming "/" (i.e., the original was just "/")
	// This should be treated as a no-op pattern
	if line == "" {
		return nil
	}

	p.pattern = line
	return p
}

// trimTrailingSpaces removes all unescaped trailing spaces from a pattern.
// Escaped trailing spaces (preceded by an odd number of backslashes) are kept,
// with the escape backslash removed.
//
// Examples:
//   "file   " -> "file"           (unescaped trailing spaces removed)
//   "file\\ " -> "file "          (escaped trailing space kept, escape removed)
//   "file\\\\ " -> "file\\\\"     (unescaped space after escaped backslash)
//   "file\\\\\\ " -> "file\\\\ "  (escaped space after escaped backslash)
//
// This matches Git's exact behavior for gitignore pattern processing.
func trimTrailingSpaces(s string) string {
	// Git's behavior: trim trailing spaces unless they are escaped
	// An escaped space is a backslash followed by a space: "\ "
	// But we need to be careful with multiple backslashes
	
	// First, find where to trim
	end := len(s)
	i := len(s) - 1
	
	// Scan backwards through trailing spaces
	for i >= 0 && s[i] == ' ' {
		// Check if this space is escaped
		if i > 0 && s[i-1] == '\\' {
			// Count consecutive backslashes before this space
			backslashCount := 0
			for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
				backslashCount++
			}
			// If odd number of backslashes, the space is escaped
			if backslashCount%2 == 1 {
				// This space is escaped, include it and the escape backslash
				// The escape backslash will be removed later
				end = i + 1
				break
			}
		}
		i--
	}
	
	// If we found no trailing spaces, end will be len(s)
	// If we found unescaped trailing spaces, end will point to first trailing space
	// If we found escaped trailing space(s), end will point after the last escaped space
	if i < len(s) - 1 {
		// We have trailing spaces (escaped or not)
		end = i + 1
	}
	
	result := s[:end]
	
	// Now handle escaped trailing spaces by removing the escape backslash
	// Only remove the backslash right before a trailing space
	if len(result) > 0 && result[len(result)-1] == ' ' {
		// Check if the last space is escaped
		if len(result) > 1 && result[len(result)-2] == '\\' {
			// Remove the escape backslash before the trailing space
			result = result[:len(result)-2] + " "
		}
	}
	
	// Return the pattern - all other escape sequences are preserved for doublestar
	return result
}

// IsIgnored checks if a path should be ignored
func (g *GitIgnore) IsIgnored(path string, isDir bool) bool {
	// Special cases
	if path == "" || path == "." {
		return false
	}

	// No path normalization - gitignore should work with paths exactly as provided
	// The caller is responsible for providing paths in the correct format
	// This preserves literal backslashes in filenames on all platforms

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
			// not when patterns like "foo/*" match content inside the directory
			dirMatches := false
			if p.dirOnly {
				// Directory-only patterns (ending with /) explicitly exclude directories
				dirMatches = matchesDirectoryPath(p, checkPath)
			} else {
				// For parent exclusion, only certain patterns actually exclude directories:
				// - Patterns that match the directory name directly (like "build")
				// - Rooted wildcard patterns like "/*" that match top-level directories  
				// - NOT patterns like "foo/*" which match content inside directories
				
				// Skip patterns that end with "/*" - these match content, not the directory
				if strings.HasSuffix(p.pattern, "/*") {
					dirMatches = false
				} else if p.pattern == "*" {
					if p.rooted {
						// Issue #3 fix: "/*" pattern SHOULD cause parent exclusion
						// for top-level directories it matches
						if !strings.Contains(checkPath, "/") {
							// This is a top-level directory that /* matches
							dirMatches = true
						} else {
							dirMatches = false
						}
					} else {
						// "*" pattern can exclude directories by matching their basename
						basename := pathpkg.Base(checkPath)
						dirMatches = matchGlob(p, basename)
					}
				} else if strings.Contains(p.pattern, "**") {
					// "**" patterns can match directories at any level
					dirMatches = matchGlob(p, checkPath)
				} else if strings.Contains(p.pattern, "/") {
					// Pattern with slash - only matches if it specifically matches directory path
					// But NOT if it's a content pattern like "foo/*"
					dirMatches = matchGlob(p, checkPath)
				} else {
					// Pattern without slash - check if it matches directory basename
					// e.g., "build" pattern excludes directory named "build"
					basename := pathpkg.Base(checkPath)
					dirMatches = matchGlob(p, basename)
				}
			}
			
			if dirMatches {
				if p.negated {
					// Issue #2 fix: Allow explicit directory negation to clear exclusion
					// Negated patterns that match directories should un-exclude them
					// This includes both:
					// - Directory-only patterns like !build/
					// - Regular patterns that match directories like !build
					delete(excludedDirs, checkPath)
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
	
	// Apply parent exclusion rule: if parent is excluded but no pattern matched this path,
	// and the path should be ignored due to parent exclusion
	if parentExcluded && !ignored {
		ignored = true
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
		basename := pathpkg.Base(path)
		if matchGlob(p, basename) {
			return true
		}
	} else {
		// For patterns with slash, they should be anchored to root (Git behavior)
		// Only match the full path since non-rooted slash patterns are treated as root-anchored
		return matchGlob(p, path)
	}

	return false
}

func matchesFilePattern(p pattern, path string, isDir bool) bool {
	// Special case: * pattern should only match files/dirs without slashes
	// BUT if it's rooted (/*), it should only match at root level
	if p.pattern == "*" {
		if p.rooted {
			// /* should only match top-level entries
			return !strings.Contains(path, "/")
		}
		// Unrooted * matches files/dirs without slashes at any depth
		basename := pathpkg.Base(path)
		return basename != "." && basename != "" // Don't match current dir or empty
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
		basename := pathpkg.Base(path)
		if matchGlob(p, basename) {
			return true
		}

		// Also check if any parent directory matches the pattern
		parts := strings.Split(path, "/")
		for i := 1; i < len(parts); i++ {
			parentPath := strings.Join(parts[:i], "/")
			parentBasename := pathpkg.Base(parentPath)
			if matchGlob(p, parentBasename) {
				return true
			}
		}
	} else {
		// For patterns with slash, they should be anchored to root (Git behavior)
		// Only match the full path since non-rooted slash patterns are treated as root-anchored
		return matchGlob(p, path)
	}

	return false
}

func matchGlob(p pattern, path string) bool {
	// The pattern has already been processed by trimTrailingSpaces,
	// which handles escape sequences for trailing spaces.
	pattern := p.pattern
	
	// Git does not support brace expansion, but doublestar does by default.
	// We need to escape unescaped braces to prevent expansion.
	pattern = escapeBraces(pattern)
	
	// Use doublestar for glob matching - it handles both escaped and unescaped chars
	matched, _ := doublestar.Match(pattern, path)
	return matched
}

// escapeBraces escapes unescaped { and } characters to prevent brace expansion
// Git treats braces as literal characters, not as expansion syntax
func escapeBraces(pattern string) string {
	var result []byte
	inCharClass := false
	
	for i := 0; i < len(pattern); i++ {
		char := pattern[i]
		
		// Track character class boundaries
		if char == '[' && (i == 0 || pattern[i-1] != '\\') {
			inCharClass = true
		} else if char == ']' && (i == 0 || pattern[i-1] != '\\') && inCharClass {
			inCharClass = false
		}
		
		// Only escape braces outside of character classes
		if !inCharClass && (char == '{' || char == '}') {
			// Check if this brace is already escaped
			// Count preceding backslashes
			backslashCount := 0
			for j := i - 1; j >= 0 && pattern[j] == '\\'; j-- {
				backslashCount++
			}
			// If even number of backslashes (including 0), the brace is not escaped
			if backslashCount%2 == 0 {
				// Add escape before the brace
				result = append(result, '\\')
			}
		}
		result = append(result, char)
	}
	return string(result)
}

// Oneliner to just test.
func Ignore(patterns []string, path string, isDir bool) bool {
	g := New(patterns)
	return g.IsIgnored(path, isDir)
}
