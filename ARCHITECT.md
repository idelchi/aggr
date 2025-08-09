# Master Go Architecture: Restructuring aggr for Excellence

## Architectural Philosophy

As a master Go architect, I see the current `aggr` application as functionally sound but architecturally improvable. While it demonstrates good Go practices, it violates key principles of **radical simplicity**, **deep interpretability**, and **clear separation of concerns**. Here's how to restructure it into an exemplary Go application.

## Current Architectural Problems

### 1. **God Package Anti-Pattern**
The `packer` package does everything:
- Orchestrates workflows
- Handles UI interactions  
- Processes patterns
- Manages file I/O
- Controls aggregation logic

This violates the **single responsibility principle** and makes testing nearly impossible.

### 2. **Unclear Domain Boundaries**
```go
// Current: Technical organization
internal/
├── packer/     // Does everything
├── checkers/   // Unclear what it checks
├── walker/     // Just walks files?
├── matcher/    // Matches what?
```

### 3. **Scattered Business Logic**
Domain rules are embedded throughout:
- Pattern normalization in multiple places
- File filtering logic in various packages  
- Archive format handling mixed with business logic

### 4. **Poor Abstraction Layers**
```go
// Current: Direct coupling
CLI -> Packer -> Everything Else
```

No clear interfaces, making components hard to test or replace.

## Master Architecture Design

### Core Principle: Screaming Architecture
The structure should **scream** what the application does. When someone looks at the packages, they should immediately understand: "This aggregates and unpacks files."

### New Structure: Domain-Driven Hexagonal Architecture

```
internal/
├── domain/                 # Pure business logic (zero external deps)
│   ├── archive/           # Archive domain entity
│   ├── file/              # File domain entity  
│   ├── pattern/           # Pattern matching domain
│   └── rule/              # Business rules
├── application/           # Use cases and orchestration
│   ├── pack/              # Pack use case
│   ├── unpack/            # Unpack use case
│   └── ports/             # Interface definitions
├── infrastructure/        # External concerns
│   ├── filesystem/        # File system adapter
│   ├── archive/           # Archive format implementation
│   └── filter/            # File filtering implementation
└── interfaces/           # User interfaces
    └── cli/              # Command-line interface
```

## Domain Layer: Pure Business Logic

### internal/domain/archive/
```go
package archive

// Entry represents a single file entry in an archive
type Entry struct {
    path    string
    content []byte
    size    int64
}

// Archive represents the business concept of a file archive
type Archive struct {
    entries []Entry
    format  Format
}

// Pack creates an archive from entries following business rules
func (a *Archive) Pack(entries []Entry) error {
    // Pure business logic - no I/O, no external dependencies
}

// Unpack extracts entries following business rules  
func (a *Archive) Unpack() ([]Entry, error) {
    // Pure business logic
}
```

### internal/domain/file/
```go
package file

// File represents the business concept of a file to be processed
type File struct {
    path     Path
    size     Size  
    content  Content
    binary   bool
}

// Path is a value object with validation
type Path string

func (p Path) Validate() error {
    // Business rule: no absolute paths, no .. segments
    if filepath.IsAbs(string(p)) {
        return ErrAbsolutePath
    }
    return nil
}
```

### internal/domain/pattern/
```go
package pattern

// Pattern represents file matching patterns
type Pattern struct {
    value      string
    normalized string
}

// MatchSet represents a collection of patterns with business rules
type MatchSet struct {
    includes []Pattern
    excludes []Pattern
}

func (ms MatchSet) Matches(file File) bool {
    // Pure pattern matching logic
}
```

## Application Layer: Use Cases

### internal/application/pack/
```go
package pack

import (
    "gitlab.garfield-labs.com/apps/aggr/internal/application/ports"
    "gitlab.garfield-labs.com/apps/aggr/internal/domain/archive"
)

// Service orchestrates the packing use case
type Service struct {
    filesystem ports.Filesystem
    filter     ports.Filter
    writer     ports.Writer
}

// Pack executes the complete packing workflow
func (s Service) Pack(request PackRequest) error {
    // 1. Validate patterns (domain logic)
    // 2. Find files (via filesystem port)
    // 3. Filter files (via filter port) 
    // 4. Create archive (domain logic)
    // 5. Write output (via writer port)
}
```

### internal/application/ports/
```go
package ports

// Filesystem port defines what we need from file system
type Filesystem interface {
    Walk(patterns []string) ([]File, error)
    Read(path string) ([]byte, error)
    Write(path string, content []byte) error
    Exists(path string) bool
}

// Filter port defines file filtering capabilities
type Filter interface {
    Include(file File) bool
    Exclude(file File) bool
}

// Writer port defines output capabilities  
type Writer interface {
    Write(archive Archive) error
}
```

## Infrastructure Layer: External Concerns

### internal/infrastructure/filesystem/
```go
package filesystem

// Adapter implements the Filesystem port
type Adapter struct {
    root string
}

func (a Adapter) Walk(patterns []string) ([]File, error) {
    // Actual file system operations
    // Uses doublestar for glob matching
    // Converts OS files to domain Files
}
```

### internal/infrastructure/filter/
```go
package filter

// CompositeFilter combines multiple filtering strategies
type CompositeFilter struct {
    filters []Filter
}

// SizeFilter filters by file size
type SizeFilter struct {
    maxSize int64
}

// BinaryFilter excludes binary files
type BinaryFilter struct{}

// IgnoreFilter applies .aggrignore patterns
type IgnoreFilter struct {
    patterns []string
}
```

## Interface Layer: CLI

### internal/interfaces/cli/
```go
package cli

// PackHandler handles the pack command
type PackHandler struct {
    packService *pack.Service
}

func (h PackHandler) Handle(cmd *cobra.Command, args []string) error {
    // 1. Parse command line arguments
    // 2. Validate input
    // 3. Create pack request
    // 4. Call pack service
    // 5. Handle response/errors
}
```

## Key Architectural Benefits

### 1. **Radical Simplicity**
Each package has ONE clear responsibility:
- `domain/archive` - Archive business logic only
- `application/pack` - Pack use case only  
- `infrastructure/filesystem` - File operations only
- `interfaces/cli` - CLI concerns only

### 2. **Deep Interpretability**
```go
// Architecture screams the domain
domain/        // "This is what we do"
  archive/     // "We work with archives" 
  file/        // "We work with files"
  pattern/     // "We match patterns"
application/   // "These are our use cases"
  pack/        // "We pack things"
  unpack/      // "We unpack things"
```

### 3. **Dependency Flow**
```
CLI -> Application -> Domain
 |         |           ^
 |         v           |
 +-> Infrastructure ---+
```

Clean, predictable dependency direction. Domain has zero external dependencies.

### 4. **Testability**
```go
// Easy to test - pure functions, injected dependencies
func TestPackService(t *testing.T) {
    // Mock filesystem
    fs := &MockFilesystem{files: testFiles}
    
    // Mock filter
    filter := &MockFilter{allowAll: true}
    
    // Mock writer
    writer := &MockWriter{}
    
    // Test the service
    service := pack.Service{
        filesystem: fs,
        filter:     filter, 
        writer:     writer,
    }
    
    err := service.Pack(packRequest)
    assert.NoError(t, err)
}
```

## Implementation Strategy

### Phase 1: Create Domain Layer (Zero External Dependencies)
```go
// Start with pure business logic
internal/domain/archive/archive.go
internal/domain/file/file.go  
internal/domain/pattern/pattern.go
```

### Phase 2: Define Application Ports
```go
// Define what we need from external world
internal/application/ports/filesystem.go
internal/application/ports/filter.go
internal/application/ports/writer.go
```

### Phase 3: Implement Infrastructure Adapters
```go
// Implement the ports
internal/infrastructure/filesystem/adapter.go
internal/infrastructure/filter/composite.go
```

### Phase 4: Create Application Services
```go
// Orchestrate use cases
internal/application/pack/service.go
internal/application/unpack/service.go
```

### Phase 5: Refactor CLI Layer
```go
// Thin handlers that delegate to services
internal/interfaces/cli/pack_handler.go
internal/interfaces/cli/unpack_handler.go
```

## Advanced Architectural Patterns

### 1. **CQRS for Read/Write Separation**
```go
// Commands (writes)
type PackCommand struct {
    Patterns []string
    Output   string
    Options  PackOptions
}

// Queries (reads)  
type ListFilesQuery struct {
    Patterns []string
    Filters  []string
}
```

### 2. **Specification Pattern for Complex Rules**
```go
package rule

type Specification interface {
    IsSatisfiedBy(file File) bool
}

type SizeSpecification struct {
    maxSize int64
}

func (s SizeSpecification) IsSatisfiedBy(file File) bool {
    return file.Size() <= s.maxSize
}

// Combine specifications
spec := And(
    SizeSpecification{maxSize: 1024 * 1024},
    Not(BinarySpecification{}),
    ExtensionSpecification{extensions: []string{"go", "md"}},
)
```

### 3. **Strategy Pattern for Archive Formats**
```go
package archive

type FormatStrategy interface {
    Marshal(entries []Entry) ([]byte, error)
    Unmarshal(data []byte) ([]Entry, error)
}

type AGGRFormat struct{}
type TarFormat struct{}
type ZipFormat struct{}

type Archive struct {
    entries  []Entry
    strategy FormatStrategy
}
```

### 4. **Observer Pattern for Progress Reporting**
```go
package application

type ProgressObserver interface {
    OnFileProcessed(file File, index, total int)
    OnComplete(summary Summary)
    OnError(err error)
}

type Service struct {
    observers []ProgressObserver
}

func (s Service) notifyFileProcessed(file File, index, total int) {
    for _, observer := range s.observers {
        observer.OnFileProcessed(file, index, total)
    }
}
```

## Testing Strategy by Layer

### Domain Layer Tests
```go
func TestArchive_Pack(t *testing.T) {
    // Pure unit tests - no mocks needed
    archive := archive.New()
    entries := []archive.Entry{
        {Path: "file1.go", Content: []byte("package main")},
        {Path: "file2.md", Content: []byte("# Title")},
    }
    
    err := archive.Pack(entries)
    assert.NoError(t, err)
    assert.Len(t, archive.Entries(), 2)
}
```

### Application Layer Tests  
```go
func TestPackService(t *testing.T) {
    // Integration tests with mocked ports
    tests := []struct {
        name     string
        request  PackRequest
        mockFS   MockFilesystem
        want     PackResponse
        wantErr  bool
    }{
        // Test cases...
    }
}
```

### Infrastructure Layer Tests
```go
func TestFilesystemAdapter(t *testing.T) {
    // Test against real file system with temp directories
    tempDir := t.TempDir()
    // Create test files...
    
    adapter := filesystem.New(tempDir)
    files, err := adapter.Walk([]string{"**/*.go"})
    assert.NoError(t, err)
    assert.Len(t, files, expectedCount)
}
```

## Configuration Management

### Clean Configuration Structure
```go
package config

// Domain-specific config groups
type PackConfig struct {
    Output     string
    MaxSize    int64
    MaxFiles   int
    Binary     bool
    StripPrefix string
}

type FilterConfig struct {
    Ignore     []string
    Extensions []string
    Hidden     bool
}

type Config struct {
    Pack   PackConfig
    Filter FilterConfig
    Root   string
    DryRun bool
}
```

## Error Handling Strategy

### Domain Errors
```go
package domain

var (
    ErrInvalidPath     = errors.New("invalid path")
    ErrFileTooBig      = errors.New("file too big")
    ErrArchiveCorrupt  = errors.New("archive corrupt")
)
```

### Application Errors
```go
package application

type PackError struct {
    File  string
    Cause error
}

func (e PackError) Error() string {
    return fmt.Sprintf("packing file %s: %v", e.File, e.Cause)
}

func (e PackError) Unwrap() error {
    return e.Cause
}
```

## Why This Architecture Excels

### 1. **Screaming Architecture**
The structure immediately tells you what the app does and how it's organized.

### 2. **Testability** 
Every component can be tested in isolation with minimal setup.

### 3. **Maintainability**
Changes to one layer don't affect others. Want a new archive format? Just implement the FormatStrategy interface.

### 4. **Understandability**
New developers can navigate from interfaces -> application -> domain and understand the flow.

### 5. **Production Readiness**
Clear error boundaries, dependency injection, comprehensive testing, proper separation of concerns.

### 6. **Blueprint Alignment**
- **Radical simplicity**: Each package has one job
- **Deep interpretability**: Structure maps to business domain
- **No partial measures**: Complete architectural transformation
- **Production-grade**: Proper testing, error handling, separation

## Conclusion

This architecture transforms `aggr` from a functional but coupled codebase into an exemplary Go application that demonstrates mastery of:

- Domain-driven design
- Hexagonal architecture  
- SOLID principles
- Clean architecture
- Testable design
- Go idioms and patterns

The result is code that's not just working, but **excellent** - maintainable, understandable, and ready for any future requirements while staying true to your established blueprint principles.