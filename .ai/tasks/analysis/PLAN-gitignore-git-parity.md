# Git-Parity Action Plan for Gitignore Implementation

## Context
We're implementing a **1-1 gitignore** system where only ONE gitignore file/patterns are considered (no recursive .gitignore files). The goal is to achieve Git-accurate semantics within this constraint.

## Critical Issues to Fix

### 1. The `*` Special Case Bug (HIGH PRIORITY)
**Problem**: Current code has `if p.pattern == "*" { return true }` which breaks `/*` patterns.
- `/*` should only match top-level entries, not everything
- Current behavior makes `/*` ignore ALL files at any depth

**Fix**: Modify the special case to respect `p.rooted` flag and path depth:
```go
// Only apply special case for non-rooted * patterns
if p.pattern == "*" && !p.rooted {
    return !strings.Contains(path, "/")
}
```

### 2. Slash-Pattern Anchoring (HIGH PRIORITY)
**Problem**: Patterns with `/` match at any depth instead of being anchored to root.
- `doc/frotz` currently matches `a/doc/frotz` (wrong)
- Should only match at root unless using `**/doc/frotz`

**Fix**: Remove subpath iteration for slash-containing patterns in:
- `matchesDirectoryPath()`
- `matchesFilePattern()`

## Tests to Change

### Tests That Must Be Updated (Currently Encode Wrong Behavior)

1. **Middle-slash group tests** - These expect floating behavior:
   - `"pattern with middle slash matches in subdirectory"` - WRONG
   - `"complex middle slash pattern"` - WRONG
   - Change expectation: `doc/frotz` should NOT match `a/doc/frotz`

2. **foo-star-special-case tests** - These expect `foo/*` to match anywhere:
   - `"foo/* matches foo/bar at any depth"` - WRONG
   - `"foo/* in deep nested structure"` - WRONG
   - Change expectation: `foo/*` should only match at root level

3. **Wildcard tests affected by * special case**:
   - Any test expecting `*` to match paths with slashes
   - Update to expect `*` only matches files/dirs without slashes

## New Tests to Add

### A. Anchoring Tests (Verify Correct Git Behavior)
```go
// Slash patterns must be anchored to root
{"anchoring", "doc/frotz at root", []string{"doc/frotz"}, "doc/frotz", false, true}
{"anchoring", "doc/frotz NOT in subdir", []string{"doc/frotz"}, "a/doc/frotz", false, false}
{"anchoring", "foo/bar at root", []string{"foo/bar"}, "foo/bar", false, true}
{"anchoring", "foo/bar NOT deeper", []string{"foo/bar"}, "x/y/foo/bar", false, false}
{"anchoring", "**/ for any depth", []string{"**/doc/frotz"}, "a/doc/frotz", false, true}
```

### B. Fixed /* Pattern Tests
```go
// /* should only match top-level
{"anchored-star", "/* matches top dir", []string{"/*"}, "folder", true, true}
{"anchored-star", "/* NOT nested files", []string{"/*"}, "folder/nested.txt", false, false}
{"anchored-star", "/* matches top file", []string{"/*"}, "top.txt", false, true}
```

### C. Edge Cases from Critique
```go
// Parent exclusion boundaries
{"parent-edges", "!dir/ doesn't re-include contents", []string{"build/", "!build/"}, "build/file.txt", false, true}
{"parent-edges", "non-slash blocks deep re-inclusion", []string{"tmp*", "!tmpcache/keep.txt"}, "tmpcache/keep.txt", false, true}

// Dir-only vs non-dirs
{"dironly-symlink", "dir/ doesn't match non-dir", []string{"symlinked-dir/"}, "symlinked-dir", false, false}

// Leading space comments
{"space-comment", "spaces before # = literal", []string{"  #notacomment"}, "#notacomment", false, true}
```

### D. Escape Sequence Tests (From Previous Session)
```go
// Already added but ensure they're included
{"escape-sequences", "literal backslash", []string{"file\\\\"}, "file\\", false, true}
{"escape-sequences", "escaped wildcard literal", []string{"\\*.txt"}, "*.txt", false, true}
{"escape-sequences", "escaped bracket literal", []string{"\\[test\\]"}, "[test]", false, true}
```

## Implementation Order

1. **Fix the * special case** - Quick win, clearly broken
2. **Add new test cases** - Establish correct expectations
3. **Fix anchoring behavior** - Remove subpath iterations
4. **Update existing tests** - Align with Git semantics
5. **Handle escape sequences** - Ensure proper processing before glob matching
6. **Run full test suite** - Iterate until all pass

## Key Principle
With single-root gitignore (no recursion), patterns should behave as if in Git's root `.gitignore`:
- Patterns with `/` are anchored to that root
- Patterns without `/` match at any depth
- `**/` explicitly matches at any depth
- `/*` only matches immediate children of root