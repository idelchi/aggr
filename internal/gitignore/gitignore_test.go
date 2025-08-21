package gitignore_test

import (
	"testing"

	"github.com/idelchi/aggr/internal/gitignore"
)

func TestGitIgnore(t *testing.T) {
	testCases := []struct {
		Group        string
		Description  string
		Patterns     []string
		Path         string
		IsDir        bool
		ShouldIgnore bool
	}{
		// Basic patterns
		{
			Group:        "basic",
			Description:  "simple file pattern",
			Patterns:     []string{"one"},
			Path:         "one",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "basic",
			Description:  "simple file pattern in subdirectory",
			Patterns:     []string{"one"},
			Path:         "a/one",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "basic",
			Description:  "non-matching file",
			Patterns:     []string{"one"},
			Path:         "two",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Wildcard patterns
		{
			Group:        "wildcards",
			Description:  "star wildcard prefix",
			Patterns:     []string{"*.o"},
			Path:         "file.o",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "wildcards",
			Description:  "star wildcard prefix in subdirectory",
			Patterns:     []string{"*.o"},
			Path:         "src/internal.o",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "wildcards",
			Description:  "star wildcard suffix",
			Patterns:     []string{"ignored-*"},
			Path:         "ignored-file",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "wildcards",
			Description:  "star wildcard suffix in subdirectory",
			Patterns:     []string{"ignored-*"},
			Path:         "a/ignored-but-in-index",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "wildcards",
			Description:  "two star prefix matches",
			Patterns:     []string{"two*"},
			Path:         "a/b/twooo",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "wildcards",
			Description:  "star suffix matches",
			Patterns:     []string{"*three"},
			Path:         "a/3-three",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Directory patterns
		{
			Group:        "directories",
			Description:  "directory with trailing slash",
			Patterns:     []string{"top-level-dir/"},
			Path:         "top-level-dir",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "directories",
			Description:  "file not matched by directory pattern",
			Patterns:     []string{"top-level-dir/"},
			Path:         "top-level-dir",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "directories",
			Description:  "nested directory with trailing slash",
			Patterns:     []string{"ignored-dir/"},
			Path:         "a/b/ignored-dir",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "directories",
			Description:  "files inside ignored directory",
			Patterns:     []string{"ignored-dir/"},
			Path:         "a/b/ignored-dir/foo",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Negation patterns
		{
			Group:        "negation",
			Description:  "negated pattern",
			Patterns:     []string{"*", "!important.txt"},
			Path:         "important.txt",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "negation",
			Description:  "negated pattern with wildcards",
			Patterns:     []string{"*.html", "!foo.html"},
			Path:         "foo.html",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "negation",
			Description:  "other html file still ignored",
			Patterns:     []string{"*.html", "!foo.html"},
			Path:         "bar.html",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "negation",
			Description:  "negated directory pattern",
			Patterns:     []string{"/*", "!/foo", "/foo/*", "!/foo/bar"},
			Path:         "foo/bar",
			IsDir:        true,
			ShouldIgnore: false,
		},

		// Rooted patterns (with leading slash)
		{
			Group:        "rooted",
			Description:  "rooted pattern matches at root",
			Patterns:     []string{"/hello.txt"},
			Path:         "hello.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "rooted",
			Description:  "rooted pattern doesn't match in subdirectory",
			Patterns:     []string{"/hello.txt"},
			Path:         "a/hello.txt",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "rooted",
			Description:  "rooted wildcard pattern",
			Patterns:     []string{"/hello.*"},
			Path:         "hello.c",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "rooted",
			Description:  "rooted wildcard pattern doesn't match in subdirectory",
			Patterns:     []string{"/hello.*"},
			Path:         "a/hello.java",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Middle slash patterns
		{
			Group:        "middle-slash",
			Description:  "pattern with middle slash",
			Patterns:     []string{"doc/frotz"},
			Path:         "doc/frotz",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "middle-slash",
			Description:  "pattern with middle slash matches with leading slash",
			Patterns:     []string{"/doc/frotz"},
			Path:         "doc/frotz",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "middle-slash",
			Description:  "pattern with middle slash matches in subdirectory",
			Patterns:     []string{"doc/frotz"},
			Path:         "a/doc/frotz",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "middle-slash",
			Description:  "foo/* matches file in foo",
			Patterns:     []string{"foo/*"},
			Path:         "foo/test.json",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "middle-slash",
			Description:  "foo/* matches directory in foo",
			Patterns:     []string{"foo/*"},
			Path:         "foo/bar",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "middle-slash",
			Description:  "foo/* does NOT match nested files (per Git spec)",
			Patterns:     []string{"foo/*"},
			Path:         "foo/bar/hello.c",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Git spec compliance tests for foo/* pattern
		{
			Group:        "git-spec-compliance",
			Description:  "foo/* matches direct file foo/test.json",
			Patterns:     []string{"foo/*"},
			Path:         "foo/test.json",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "git-spec-compliance",
			Description:  "foo/* matches direct directory foo/bar",
			Patterns:     []string{"foo/*"},
			Path:         "foo/bar",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "git-spec-compliance",
			Description:  "foo/* does NOT match nested foo/bar/hello.c",
			Patterns:     []string{"foo/*"},
			Path:         "foo/bar/hello.c",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Parent directory exclusion rule tests
		{
			Group:        "parent-exclusion",
			Description:  "Cannot re-include file if parent dir excluded",
			Patterns:     []string{"build/", "!build/important.txt"},
			Path:         "build/important.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "parent-exclusion",
			Description:  "Can re-include directory itself",
			Patterns:     []string{"build/", "!build/"},
			Path:         "build",
			IsDir:        true,
			ShouldIgnore: false,
		},

		// Double asterisk patterns
		{
			Group:        "double-asterisk",
			Description:  "**/foo matches foo anywhere",
			Patterns:     []string{"**/foo"},
			Path:         "foo",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "double-asterisk",
			Description:  "**/foo matches foo in subdirectory",
			Patterns:     []string{"**/foo"},
			Path:         "a/b/c/foo",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "double-asterisk",
			Description:  "**/foo/bar matches bar under any foo",
			Patterns:     []string{"**/foo/bar"},
			Path:         "foo/bar",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "double-asterisk",
			Description:  "**/foo/bar matches bar under any foo in subdirectory",
			Patterns:     []string{"**/foo/bar"},
			Path:         "a/b/foo/bar",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "double-asterisk",
			Description:  "abc/** matches everything inside abc",
			Patterns:     []string{"abc/**"},
			Path:         "abc/file.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "double-asterisk",
			Description:  "abc/** matches deeply nested file",
			Patterns:     []string{"abc/**"},
			Path:         "abc/x/y/z/file.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "double-asterisk",
			Description:  "a/**/b matches a/b",
			Patterns:     []string{"a/**/b"},
			Path:         "a/b",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "double-asterisk",
			Description:  "a/**/b matches a/x/b",
			Patterns:     []string{"a/**/b"},
			Path:         "a/x/b",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "double-asterisk",
			Description:  "a/**/b matches a/x/y/b",
			Patterns:     []string{"a/**/b"},
			Path:         "a/x/y/b",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Question mark patterns
		{
			Group:        "question-mark",
			Description:  "? matches single character",
			Patterns:     []string{"file.?"},
			Path:         "file.c",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "question-mark",
			Description:  "? doesn't match multiple characters",
			Patterns:     []string{"file.?"},
			Path:         "file.cc",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Range patterns
		{
			Group:        "range",
			Description:  "[a-z] matches lowercase letter",
			Patterns:     []string{"file[a-z].txt"},
			Path:         "filec.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "range",
			Description:  "[a-z] doesn't match uppercase",
			Patterns:     []string{"file[a-z].txt"},
			Path:         "fileC.txt",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "range",
			Description:  "[a-zA-Z] matches any letter",
			Patterns:     []string{"file[a-zA-Z].txt"},
			Path:         "fileC.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Complex patterns from test suite
		{
			Group:        "complex",
			Description:  "vmlinux* pattern",
			Patterns:     []string{"vmlinux*"},
			Path:         "arch/foo/kernel/vmlinux.lds.S",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "complex",
			Description:  "negation overrides globally with single root gitignore",
			Patterns:     []string{"vmlinux*", "!vmlinux*"},
			Path:         "arch/foo/kernel/vmlinux.lds.S",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Escaped patterns
		{
			Group:        "escaped",
			Description:  "escaped exclamation mark",
			Patterns:     []string{`\!important!.txt`},
			Path:         "!important!.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "escaped",
			Description:  "escaped hash",
			Patterns:     []string{`\#hashtag`},
			Path:         "#hashtag",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Comment and blank line handling
		{
			Group:        "special-lines",
			Description:  "pattern after comment",
			Patterns:     []string{"# this is a comment", "actual-pattern"},
			Path:         "actual-pattern",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "special-lines",
			Description:  "pattern after blank line",
			Patterns:     []string{"pattern1", "", "pattern2"},
			Path:         "pattern2",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Trailing whitespace
		{
			Group:        "whitespace",
			Description:  "trailing spaces are ignored",
			Patterns:     []string{"trailing   "},
			Path:         "trailing",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "whitespace",
			Description:  "escaped trailing spaces",
			Patterns:     []string{`trailing\ \ `},
			Path:         "trailing  ",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// From test suite - nested includes with negation
		{
			Group:        "nested-negation",
			Description:  "multiple negation levels - on* pattern negates globally",
			Patterns:     []string{"four", "five", "six", "ignored-dir/", "!on*", "!two"},
			Path:         "a/b/one",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Edge cases
		{
			Group:        "edge-cases",
			Description:  "dot as path",
			Patterns:     []string{"*"},
			Path:         ".",
			IsDir:        true,
			ShouldIgnore: false,
		},
		{
			Group:        "edge-cases",
			Description:  "empty path",
			Patterns:     []string{"*"},
			Path:         "",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Exact prefix matching tests from suite
		{
			Group:        "prefix-matching",
			Description:  "git/ matches git directory",
			Patterns:     []string{"git/"},
			Path:         "a/git",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "prefix-matching",
			Description:  "git/ matches file in git directory",
			Patterns:     []string{"git/"},
			Path:         "a/git/foo",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "prefix-matching",
			Description:  "git/ doesn't match git-foo directory",
			Patterns:     []string{"git/"},
			Path:         "a/git-foo",
			IsDir:        true,
			ShouldIgnore: false,
		},
		{
			Group:        "prefix-matching",
			Description:  "git/ doesn't match file in git-foo",
			Patterns:     []string{"git/"},
			Path:         "a/git-foo/bar",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "prefix-matching",
			Description:  "/git/ matches git at root",
			Patterns:     []string{"/git/"},
			Path:         "git",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "prefix-matching",
			Description:  "/git/ doesn't match git in subdirectory",
			Patterns:     []string{"/git/"},
			Path:         "a/git",
			IsDir:        true,
			ShouldIgnore: false,
		},

		// Data/** pattern tests from suite
		{
			Group:        "complex-double-asterisk",
			Description:  "data/** ignores file",
			Patterns:     []string{"data/**", "!data/**/", "!data/**/*.txt"},
			Path:         "data/file",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "complex-double-asterisk",
			Description:  "data/** ignores nested file",
			Patterns:     []string{"data/**", "!data/**/", "!data/**/*.txt"},
			Path:         "data/data1/file1",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "complex-double-asterisk",
			Description:  "data/** allows .txt files",
			Patterns:     []string{"data/**", "!data/**/", "!data/**/*.txt"},
			Path:         "data/data1/file1.txt",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "complex-double-asterisk",
			Description:  "data/** allows nested .txt files",
			Patterns:     []string{"data/**", "!data/**/", "!data/**/*.txt"},
			Path:         "data/data2/file2.txt",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Extra tests
		// Ignore all except directories and specific extensions
		{
			Group:        "except-dirs-and-ext",
			Description:  "* !*/ !*.txt - hidden file matched by *",
			Patterns:     []string{"*", "!*/", "!*.txt"},
			Path:         ".file",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "except-dirs-and-ext",
			Description:  "* !*/ !*.txt - go file in subdir ignored",
			Patterns:     []string{"*", "!*/", "!*.txt"},
			Path:         "internal/file.go",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "except-dirs-and-ext",
			Description:  "* !*/ !*.txt - txt file in subdir allowed",
			Patterns:     []string{"*", "!*/", "!*.txt"},
			Path:         "internal/file.txt",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "except-dirs-and-ext",
			Description:  "* !*/ !*.txt - directory allowed",
			Patterns:     []string{"*", "!*/", "!*.txt"},
			Path:         "internal",
			IsDir:        true,
			ShouldIgnore: false,
		},
		{
			Group:        "except-dirs-and-ext",
			Description:  "* !*/ !*.txt - root txt file allowed",
			Patterns:     []string{"*", "!*/", "!*.txt"},
			Path:         "readme.txt",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Same but with specific directory ignored
		{
			Group:        "except-dirs-and-ext-with-ignored",
			Description:  "* !*/ !*.txt internal - hidden file ignored",
			Patterns:     []string{"*", "!*/", "!*.txt", "internal"},
			Path:         ".file",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "except-dirs-and-ext-with-ignored",
			Description:  "* !*/ !*.txt internal - go file in internal ignored",
			Patterns:     []string{"*", "!*/", "!*.txt", "internal"},
			Path:         "internal/file.go",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "except-dirs-and-ext-with-ignored",
			Description:  "* !*/ !*.txt internal - txt file in internal ignored",
			Patterns:     []string{"*", "!*/", "!*.txt", "internal"},
			Path:         "internal/file.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "except-dirs-and-ext-with-ignored",
			Description:  "* !*/ !*.txt internal - internal dir ignored",
			Patterns:     []string{"*", "!*/", "!*.txt", "internal"},
			Path:         "internal",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "except-dirs-and-ext-with-ignored",
			Description:  "* !*/ !*.txt internal - root txt file allowed",
			Patterns:     []string{"*", "!*/", "!*.txt", "internal"},
			Path:         "file.txt",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "except-dirs-and-ext-with-ignored",
			Description:  "* !*/ !*.txt internal - txt in other dir allowed",
			Patterns:     []string{"*", "!*/", "!*.txt", "internal"},
			Path:         "pkg/file.txt",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Multiple extensions allowed
		{
			Group:        "multiple-extensions",
			Description:  "* !*/ !*.go !*.mod !*.sum - go file allowed",
			Patterns:     []string{"*", "!*/", "!*.go", "!*.mod", "!*.sum"},
			Path:         "main.go",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "multiple-extensions",
			Description:  "* !*/ !*.go !*.mod !*.sum - mod file allowed",
			Patterns:     []string{"*", "!*/", "!*.go", "!*.mod", "!*.sum"},
			Path:         "go.mod",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "multiple-extensions",
			Description:  "* !*/ !*.go !*.mod !*.sum - nested go file allowed",
			Patterns:     []string{"*", "!*/", "!*.go", "!*.mod", "!*.sum"},
			Path:         "cmd/app/main.go",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "multiple-extensions",
			Description:  "* !*/ !*.go !*.mod !*.sum - txt file ignored",
			Patterns:     []string{"*", "!*/", "!*.go", "!*.mod", "!*.sum"},
			Path:         "readme.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "multiple-extensions",
			Description:  "* !*/ !*.go !*.mod !*.sum - binary ignored",
			Patterns:     []string{"*", "!*/", "!*.go", "!*.mod", "!*.sum"},
			Path:         "app",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Build artifacts ignored but source allowed
		{
			Group:        "build-artifacts",
			Description:  "* !*/ !*.go build/ - source in build ignored",
			Patterns:     []string{"*", "!*/", "!*.go", "build/"},
			Path:         "build/generated.go",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "build-artifacts",
			Description:  "* !*/ !*.go build/ - build dir ignored",
			Patterns:     []string{"*", "!*/", "!*.go", "build/"},
			Path:         "build",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "build-artifacts",
			Description:  "* !*/ !*.go build/ - source outside build allowed",
			Patterns:     []string{"*", "!*/", "!*.go", "build/"},
			Path:         "src/main.go",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Vendor exception pattern
		{
			Group:        "vendor-exception",
			Description:  "* !*/ !*.go vendor/**/*.go - vendor go file ignored",
			Patterns:     []string{"*", "!*/", "!*.go", "vendor/**/*.go"},
			Path:         "vendor/github.com/pkg/file.go",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "vendor-exception",
			Description:  "* !*/ !*.go vendor/**/*.go - non-vendor go allowed",
			Patterns:     []string{"*", "!*/", "!*.go", "vendor/**/*.go"},
			Path:         "internal/app.go",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "vendor-exception",
			Description:  "* !*/ !*.go vendor/**/*.go - vendor dir allowed",
			Patterns:     []string{"*", "!*/", "!*.go", "vendor/**/*.go"},
			Path:         "vendor",
			IsDir:        true,
			ShouldIgnore: false,
		},

		// Hidden files and directories special handling
		{
			Group:        "hidden-special",
			Description:  "* !*/ !.gitignore - .gitignore allowed",
			Patterns:     []string{"*", "!*/", "!.gitignore"},
			Path:         ".gitignore",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "hidden-special",
			Description:  "* !*/ !.gitignore - other hidden file matched by *",
			Patterns:     []string{"*", "!*/", "!.gitignore"},
			Path:         ".env",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "hidden-special",
			Description:  "* !*/ !.gitignore - nested .gitignore allowed",
			Patterns:     []string{"*", "!*/", "!.gitignore"},
			Path:         "subdir/.gitignore",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "hidden-special",
			Description:  "* !*/ !.git/ - .git dir allowed",
			Patterns:     []string{"*", "!*/", "!.git/"},
			Path:         ".git",
			IsDir:        true,
			ShouldIgnore: false,
		},

		// Test files special case
		{
			Group:        "test-files",
			Description:  "* !*/ !*_test.go - test file allowed",
			Patterns:     []string{"*", "!*/", "!*_test.go"},
			Path:         "main_test.go",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "test-files",
			Description:  "* !*/ !*_test.go - nested test file allowed",
			Patterns:     []string{"*", "!*/", "!*_test.go"},
			Path:         "pkg/utils/helper_test.go",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "test-files",
			Description:  "* !*/ !*_test.go - non-test go file ignored",
			Patterns:     []string{"*", "!*/", "!*_test.go"},
			Path:         "main.go",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Documentation files only
		{
			Group:        "docs-only",
			Description:  "* !*/ !*.md !*.txt !LICENSE - md file allowed",
			Patterns:     []string{"*", "!*/", "!*.md", "!*.txt", "!LICENSE"},
			Path:         "README.md",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "docs-only",
			Description:  "* !*/ !*.md !*.txt !LICENSE - LICENSE allowed",
			Patterns:     []string{"*", "!*/", "!*.md", "!*.txt", "!LICENSE"},
			Path:         "LICENSE",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "docs-only",
			Description:  "* !*/ !*.md !*.txt !LICENSE - nested md allowed",
			Patterns:     []string{"*", "!*/", "!*.md", "!*.txt", "!LICENSE"},
			Path:         "docs/api.md",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "docs-only",
			Description:  "* !*/ !*.md !*.txt !LICENSE - go file ignored",
			Patterns:     []string{"*", "!*/", "!*.md", "!*.txt", "!LICENSE"},
			Path:         "main.go",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Complex: ignore all except go, then re-ignore generated
		{
			Group:        "except-go-ignore-generated",
			Description:  "allow go but ignore generated - regular go allowed",
			Patterns:     []string{"*", "!*/", "!*.go", "*.generated.go", "*.pb.go"},
			Path:         "main.go",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "except-go-ignore-generated",
			Description:  "allow go but ignore generated - generated ignored",
			Patterns:     []string{"*", "!*/", "!*.go", "*.generated.go", "*.pb.go"},
			Path:         "models.generated.go",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "except-go-ignore-generated",
			Description:  "allow go but ignore generated - pb.go ignored",
			Patterns:     []string{"*", "!*/", "!*.go", "*.generated.go", "*.pb.go"},
			Path:         "api/service.pb.go",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Node modules style ignore with exceptions
		{
			Group:        "node-style",
			Description:  "* !*/ !package.json node_modules - package.json in node_modules ignored",
			Patterns:     []string{"*", "!*/", "!package.json", "node_modules"},
			Path:         "node_modules/package.json",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "node-style",
			Description:  "* !*/ !package.json node_modules - root package.json allowed",
			Patterns:     []string{"*", "!*/", "!package.json", "node_modules"},
			Path:         "package.json",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "node-style",
			Description:  "* !*/ !package.json node_modules - node_modules dir ignored",
			Patterns:     []string{"*", "!*/", "!package.json", "node_modules"},
			Path:         "node_modules",
			IsDir:        true,
			ShouldIgnore: true,
		},

		// Test to verify the special case code for foo/* is unnecessary
		{
			Group:        "foo-star-special-case",
			Description:  "foo/* at nested level should work without special code",
			Patterns:     []string{"foo/*"},
			Path:         "deep/nested/foo/bar",
			IsDir:        false,
			ShouldIgnore: true, // This should work even without the special case code
		},
		{
			Group:        "foo-star-special-case",
			Description:  "foo/* should not match foo/bar/baz at any level",
			Patterns:     []string{"foo/*"},
			Path:         "deep/nested/foo/bar/baz",
			IsDir:        false,
			ShouldIgnore: false,
		},

		// Tests to verify the directory matching logic handles these edge cases correctly
		{
			Group:        "dir-pattern-edge",
			Description:  "build pattern should match build directory at any level",
			Patterns:     []string{"build"},
			Path:         "src/build",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "dir-pattern-edge",
			Description:  "build pattern should match files inside build at any level",
			Patterns:     []string{"build"},
			Path:         "src/build/output.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "dir-pattern-edge",
			Description:  "build pattern should match deeply nested build dirs",
			Patterns:     []string{"build"},
			Path:         "a/b/c/d/build/e/f/g.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},

		// Tests to verify wildcard directory negation behavior
		{
			Group:        "wildcard-dir-negation",
			Description:  "!*/ should un-ignore directories only, not files",
			Patterns:     []string{"*", "!*/"},
			Path:         "file.txt",
			IsDir:        false,
			ShouldIgnore: true, // File should remain ignored
		},
		{
			Group:        "wildcard-dir-negation",
			Description:  "!*/ should un-ignore all directories",
			Patterns:     []string{"*", "!*/"},
			Path:         "somedir",
			IsDir:        true,
			ShouldIgnore: false, // Directory should be un-ignored
		},
		{
			Group:        "wildcard-dir-negation",
			Description:  "!**/ should work similarly for nested dirs",
			Patterns:     []string{"**/*", "!**/"},
			Path:         "deep/nested/dir",
			IsDir:        true,
			ShouldIgnore: false,
		},
		{
			Group:        "wildcard-dir-negation",
			Description:  "!**/ should not un-ignore files",
			Patterns:     []string{"**/*", "!**/"},
			Path:         "deep/nested/file.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "wildcard-dir-negation",
			Description:  "!*/ should un-ignore all directories but not files",
			Patterns:     []string{"*", "!*/"},
			Path:         "a/file.txt",
			IsDir:        false,
			ShouldIgnore: true, // File should remain ignored
		},
		{
			Group:        "wildcard-dir-negation",
			Description:  "!*/ should un-ignore all directories but only files explicitly given",
			Patterns:     []string{"*", "!*/", "!*.go"},
			Path:         "a/file.txt",
			IsDir:        false,
			ShouldIgnore: true, // File should remain ignored
		},
		{
			Group:        "wildcard-dir-negation",
			Description:  "!*/ should un-ignore all directories but only files explicitly given",
			Patterns:     []string{"*", "!*/", "!*.go"},
			Path:         "a/b/file.go",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "wildcard-dir-negation",
			Description:  "!*/ should un-ignore all directories but only files explicitly given",
			Patterns:     []string{"*", "!*/", "!*.go"},
			Path:         "file.go",
			IsDir:        false,
			ShouldIgnore: false,
		},

		{
			Group:        "parent-exclusion-non-dir",
			Description:  "cannot re-include file if parent dir excluded by non-slash pattern",
			Patterns:     []string{"build", "!build/important.txt"},
			Path:         "build/important.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "parent-exclusion-non-dir",
			Description:  "cannot re-include directory if parent dir excluded by non-slash pattern",
			Patterns:     []string{"build", "!build/sub/"},
			Path:         "build/sub",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "parent-exclusion-dironly",
			Description:  "cannot re-include subdirectory when parent directory excluded",
			Patterns:     []string{"build/", "!build/sub/"},
			Path:         "build/sub",
			IsDir:        true,
			ShouldIgnore: true,
		},
		{
			Group:        "negated-dir-does-not-include-files",
			Description:  "!build/ should not un-ignore files inside without explicit file rules",
			Patterns:     []string{"*", "!build/"},
			Path:         "build/file.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "negated-dir-does-not-include-files",
			Description:  "!/build/ should not un-ignore files inside without explicit file rules",
			Patterns:     []string{"/", "!/build/"},
			Path:         "build/file.txt",
			IsDir:        false,
			ShouldIgnore: false,
		},
		{
			Group:        "negated-dir-does-not-include-files",
			Description:  "!/build/ should not un-ignore files inside without explicit file rules",
			Patterns:     []string{"/*", "!/build/"},
			Path:         "build/file.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "parent-exclusion-wildcard",
			Description:  "cannot re-include file if parent directory excluded by wildcard name pattern",
			Patterns:     []string{"tmp*", "!tmpcache/keep.txt"},
			Path:         "tmpcache/keep.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "parent-exclusion-non-dir-depth",
			Description:  "cannot re-include deep file if ancestor dir excluded by non-slash pattern",
			Patterns:     []string{"build", "!deep/build/important.txt"},
			Path:         "deep/build/important.txt",
			IsDir:        false,
			ShouldIgnore: true,
		},
		{
			Group:        "parent-exclusion-dironly-wildcard",
			Description:  "cannot re-include nested directory under an ignored parent directory",
			Patterns:     []string{"foo*/", "!foobar/baz/"},
			Path:         "foobar/baz",
			IsDir:        true,
			ShouldIgnore: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Group+"/"+tc.Description, func(t *testing.T) {
			isIgnored := gitignore.Ignore(tc.Patterns, tc.Path, tc.IsDir)
			if isIgnored != tc.ShouldIgnore {
				t.Errorf("Ignore(patterns=%v, path=%q, isDir=%v) = %v, want %v",
					tc.Patterns, tc.Path, tc.IsDir, isIgnored, tc.ShouldIgnore)
			}
		})
	}
}
