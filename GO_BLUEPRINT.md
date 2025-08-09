# Go Application Blueprint

## Philosophy & Principles

### Core Values
- **Radical simplicity** - Avoid complexity unless absolutely necessary
- **Deep interpretability** - Code should be readable by both technical and non-technical audiences
- **Production-grade quality** - Every piece of code should be ready for production use
- **Minimal dependencies** - Use standard library where possible, add dependencies only when they provide significant value

### Design Philosophy
- **No partial measures** - Implement features completely, avoid "gradual migrations" or feature flags
- **Clean switches** - Changes should be atomic and complete
- **User-first design** - Focus on intuitive user experience over technical elegance
- **Tools that disappear** - The tool should do its job without getting in the way

## Project Structure

### Directory Layout
```
project/
├── main.go                 # Entry point, minimal logic
├── go.mod                  # Dependencies (keep minimal)
├── go.sum                  # Dependency checksums
├── README.md              # User-facing documentation
├── Taskfile.yml           # Task automation
├── internal/              # Private packages
│   ├── cli/              # CLI command implementations
│   │   ├── root.go       # Root command setup
│   │   ├── common.go     # Shared utilities
│   │   └── *.go          # Subcommands
│   ├── [domain]/         # Domain-specific packages
│   │   ├── doc.go        # Package documentation
│   │   ├── [type].go     # Core types
│   │   ├── [type]_test.go # Tests for each file
│   │   └── testdata/     # Test fixtures
│   └── terminal/         # Terminal-specific utilities
├── .github/              # GitHub configuration
│   └── workflows/        # CI/CD pipelines
├── .devenv/              # Development environment
└── assets/               # Static assets (images, gifs)
```

### Package Organization
- **Single responsibility** - Each package has one clear purpose
- **Internal by default** - Use `internal/` to prevent external imports
- **Domain separation** - Organize by business domain, not technical layers
- **Package documentation** - Every package must have a `doc.go` file

## Code Style

### Naming Conventions

#### Files
- **Descriptive names**: `stringify.go`, `inheritance.go`, `marshal.go`
- **Domain grouping**: Related functionality in same file
- **Test files**: Always `*_test.go` adjacent to source
- **Documentation**: `doc.go` for package-level docs

#### Functions & Methods
```go
// Exported functions - PascalCase
func Execute(version string) error
func New(file file.File) (*Store, error)

// Private functions - camelCase  
func load(files []string) (*Profiles, error)
func needsQuotes(s string) bool

// Constructor pattern - always New()
func New() *Type
func newPrivateType() *privateType
```

#### Variables & Types
```go
// Types - PascalCase
type Profile struct
type InheritanceTracker struct

// Interfaces - PascalCase with -er suffix
type Marshaler interface

// Constants - Context-appropriate
const YAML = "yaml"  // File types
const ErrProfileNotFound = errors.New("profile not found")  // Sentinel errors

// Variables - camelCase
var verbose bool
profileName := "dev"
```

### Struct Tags
```go
type Profile struct {
    // Consistent alignment and spacing
    Env     Env      `toml:"env,omitempty"     yaml:"env,omitempty"`
    DotEnv  []string `toml:"dotenv,omitempty"  yaml:"dotenv,omitempty"`
    Extends []string `toml:"extends,omitempty" yaml:"extends,omitempty"`
}
```

### Comments & Documentation

#### Package Documentation
```go
// Package profile provides profile and store management for environment variable sets.
package profile
```

#### Function Documentation
```go
// Execute runs the root command for the envprof CLI application.
func Execute(version string) error

// Environment returns the merged environment variables for a profile, resolving dependencies.
//
//nolint:gocognit // TODO(Idelchi): Refactor this function to reduce cognitive complexity.
func (p Profiles) Environment(name string) (env.Env, error)
```

#### Inline Comments
- **Sparse usage** - Code should be self-documenting
- **Focus on why** - Explain reasoning, not mechanics
- **NO decorative comments** - Never add unless explicitly asked

#### Nolint Directives
```go
//nolint:wrapcheck // Error does not need additional wrapping.
//nolint:err113 // Dynamic error is acceptable here.
//nolint:mnd // The command takes up to 2 arguments as documented.
```

## Error Handling

### Sentinel Errors
```go
var (
    ErrUnsupportedFileType = errors.New("unsupported file type")
    ErrProfileNotFound     = errors.New("profile not found")
)
```

### Error Wrapping Strategy
```go
// Wrap when adding context
return fmt.Errorf("failed to load profile %q: %w", name, err)

// Don't wrap when context is clear
return err //nolint:wrapcheck // Error does not need additional wrapping.
```

### Error Aggregation
```go
var errs []error
// Collect errors...
return result, errors.Join(errs...)
```

### Dynamic Errors
```go
// Use sparingly with justification
return fmt.Errorf("profile file not found: searched for %v", paths) //nolint:err113
```

## Testing Philosophy

### Test Structure
```go
func TestType_Method(t *testing.T) {
    t.Parallel()  // Always use parallel when possible
    
    tests := []struct {
        name    string
        input   string
        want    Result
        wantErr bool
    }{
        // Test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // Test implementation
        })
    }
}
```

### Test Organization
- **Table-driven tests** - Primary testing pattern
- **Parallel execution** - Use `t.Parallel()` liberally
- **Test fixtures** - Organized in `testdata/` with `valid/` and `invalid/` subdirs
- **Integration tests** - Controlled via environment variables
- **Black-box testing** - Use `package_test` for API tests
- **White-box testing** - Same package for internal tests

### Test Naming
```go
// Test functions
func TestNewProfile(t *testing.T)
func TestProfile_BasicFields(t *testing.T)
func TestProfiles_Environment(t *testing.T)

// Subtests - descriptive scenarios
t.Run("simple profile", func(t *testing.T) {})
t.Run("with circular dependency", func(t *testing.T) {})
```

## Patterns & Idioms

### Type Assertions & Switches
```go
switch val := v.(type) {
case nil:
    return "", nil
case string:
    return val, nil
case bool:
    return strconv.FormatBool(val), nil
default:
    // JSON fallback for complex types
    data, err := json.Marshal(val)
    return string(data), err
}
```

### Map Operations
```go
// Existence check
if _, ok := profiles[name]; ok {
    // exists
}

// Initialize with capacity
nodes := make([]string, 0, len(profiles))
```

### Method Receivers
```go
// Pointer receiver - modifies state or large struct
func (s *Store) Load() (*Store, error)

// Value receiver - simple queries, immutable operations
func (p Profiles) Exists(name string) bool
```

### Fluent Interfaces
```go
// Return self for chaining
func (s *Store) Load() (*Store, error) {
    // implementation
    return s, nil
}
```

## CLI Design

### Command Structure
```go
root := &cobra.Command{
    Use:   "tool",
    Short: "Brief description",
    Long: heredoc.Doc(`
        Detailed description with examples
        and usage patterns.
    `),
    Example: heredoc.Doc(`
        # Example usage
        $ tool command args
    `),
}
```

### Help Text
- **Use heredoc** - Clean multi-line strings
- **Provide examples** - Show real usage
- **Be concise** - CLI users want quick answers

### Subcommand Pattern
```go
func List(files *[]string) *cobra.Command {
    cmd := &cobra.Command{
        Use:     "list [profile] [key]",
        Short:   "List profiles and their variables",
        Aliases: []string{"ls"},
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
        },
    }
    return cmd
}
```

## README Structure

### Essential Sections
1. **Header** - Logo, title, tagline
2. **Badges** - Release, docs, build status, license
3. **Brief description** - One-liner explaining what it does
4. **Features** - Bullet points of key capabilities
5. **Installation** - Quick start command
6. **Usage** - Common use cases with examples
7. **Format/Configuration** - How to configure with examples
8. **Demo** - GIF showing the tool in action

### README Style
- **Visual hierarchy** - Use headers, badges, and formatting
- **Copy-paste ready** - All examples should work as-is
- **Progressive disclosure** - Use `<details>` for advanced topics
- **Centered header** - Logo and title aligned center
- **Practical examples** - Show real-world usage

## Dependencies

### Selection Criteria
- **Essential functionality only** - Must provide significant value
- **Well-maintained** - Active development, stable API
- **Minimal transitive deps** - Avoid dependency sprawl
- **Standard choices** - Use community standards when available

### Approved Patterns
```go
// CLI Framework
github.com/spf13/cobra         // Standard CLI framework

// Configuration
github.com/BurntSushi/toml     // TOML parsing
github.com/goccy/go-yaml       // YAML with better features than stdlib

// Documentation
github.com/MakeNowJust/heredoc // Clean multi-line strings

// Custom utilities
github.com/idelchi/*           // Own utility libraries
```

## Build & CI/CD

### Task Automation
- **Taskfile.yml** - Primary automation tool
- **Consistent commands** - `task lint`, `task test`, `task build`
- **Docker support** - Development environment containerization

### GitHub Actions
- **Reusable workflows** - Centralized in devenv repository
- **Staged releases** - dev → prerelease → release
- **Automated versioning** - Semantic versioning with tags

### Version Stamping
```go
// Global variable for CI stamping
var version = "unknown - unofficial & generated by unknown"

// Stamped during build
func main() {
    if err := cli.Execute(version); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

## Terminal & User Experience

### Output Philosophy
- **Minimal by default** - Only show what's needed
- **Progressive verbosity** - `-v` flag for details
- **Structured output** - Consistent formatting
- **Error clarity** - Clear, actionable error messages

### Interactive Features
```go
// Subshell detection
if os.Getenv("ENVPROF_ACTIVE_PROFILE") != "" {
    return errors.New("already in an envprof shell")
}

// Shell integration
os.Setenv("ENVPROF_ACTIVE_PROFILE", profileName)
```

### Color & Formatting
- **No unnecessary colors** - Let terminal theme handle it
- **Consistent alignment** - Use formatting for readability
- **No emojis** - Unless explicitly requested

## Security Practices

### Environment Variables
- **Never log secrets** - Sanitize sensitive values
- **No hardcoded credentials** - Always use configuration
- **Secure defaults** - Fail closed, not open

### File Operations
- **Validate paths** - Prevent directory traversal
- **Check permissions** - Ensure proper access rights
- **Atomic operations** - Use temp files and rename

## Performance Considerations

### Optimization Principles
- **Measure first** - Profile before optimizing
- **Batch operations** - Group related work
- **Preallocate slices** - When size is known
- **Lazy evaluation** - Defer expensive operations

### Common Patterns
```go
// Preallocate with capacity
result := make([]string, 0, len(input))

// Early returns
if err != nil {
    return nil, err
}

// Parallel processing
t.Parallel()
```

## Development Workflow

### Branch Strategy
- **main** - Stable releases
- **dev** - Development branch
- **Feature branches** - Short-lived, focused changes

### Commit Messages
- **Descriptive** - Explain why, not what
- **Conventional** - Follow existing patterns
- **Atomic** - One logical change per commit

### Code Review
- **Test coverage** - All new code must be tested
- **Linting passes** - No lint warnings
- **Documentation** - Update docs with code changes

## Vibe & Tone

### Code Personality
- **Professional** - Clean, serious, production-ready
- **Pragmatic** - Solve real problems, avoid over-engineering
- **Respectful** - Clear error messages, helpful documentation
- **Efficient** - Do the job well and get out of the way

### User Communication
- **Direct** - No fluff or unnecessary verbosity
- **Helpful** - Provide examples and clear guidance
- **Respectful** - Assume competence, avoid condescension
- **Action-oriented** - Focus on what to do, not theory

## Anti-Patterns to Avoid

### Code Anti-Patterns
- ❌ Decorative comments
- ❌ Unnecessary abstractions
- ❌ Premature optimization
- ❌ Feature flags for incomplete work
- ❌ Partial implementations
- ❌ Excessive logging
- ❌ Complex inheritance hierarchies
- ❌ Global state

### Design Anti-Patterns
- ❌ "Gradual migrations"
- ❌ Multiple ways to do the same thing
- ❌ Configuration sprawl
- ❌ Breaking changes without major version bump
- ❌ Undocumented behavior
- ❌ Implicit magic

### Testing Anti-Patterns
- ❌ Flaky tests
- ❌ Testing implementation details
- ❌ Excessive mocking
- ❌ Slow tests without reason
- ❌ Missing error cases
- ❌ Hardcoded test data

## Summary

This blueprint represents a commitment to writing Go applications that are:
- **Simple** without being simplistic
- **Complete** without being complex
- **Fast** without being fragile
- **Documented** without being verbose
- **Tested** without being brittle
- **Professional** without being pretentious

Every line of code should serve the user's needs, be ready for production, and be maintainable by future developers. The goal is software that does its job exceptionally well and then gets out of the way.