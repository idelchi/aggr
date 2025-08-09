# Architecture Analysis and Improvement Plan for aggr Application

## Current Architecture Analysis

### Overview
The aggr application is a file aggregation tool that packs multiple files into a single archive and can unpack them back. The codebase follows a multi-package internal structure with various responsibilities distributed across packages.

## Current Issues and Areas for Improvement

### 1. Separation of Concerns Issues

#### Problem Areas:
- **CLI Package Mixed Responsibilities**: The `internal/cli` package contains both command definition and partial business logic
- **Packer Package Overload**: The `internal/packer` package handles too many responsibilities:
  - File collection orchestration
  - Pattern processing 
  - Checker configuration
  - Output management
  - User prompting (UI concern in business logic)
  - Both packing AND unpacking logic
- **Walker-Matcher Overlap**: Both `walker` and `matcher` packages deal with file collection, creating unclear boundaries
- **Config Package Mixing**: Configuration structs contain both runtime options and business rules

### 2. Domain Boundary Issues

#### Current Problems:
- **No Clear Domain Model**: Files are represented as strings/paths throughout, no rich domain types
- **Pattern Processing Scattered**: Pattern logic split between `patterns`, `packer`, and `matcher`
- **Missing Core Domain**: No central domain package defining the core business concepts
- **Technical vs Business Split**: Packages organized by technical role rather than business capability

### 3. Coupling Problems

#### High Coupling Areas:
- **Packer → Everything**: The packer package imports from almost all other packages
- **Circular Conceptual Dependencies**: Matcher uses Walker, which uses Checkers, creating tight coupling
- **Logger Interface Duplication**: Both `walker` and `matcher` define identical Logger interfaces
- **Direct File System Access**: Multiple packages directly access the file system instead of through abstractions

### 4. Package Organization Issues

#### Current Problems:
- **Unclear Package Hierarchy**: No clear layering or dependency direction
- **Missing Abstraction Layers**: Direct jumps from CLI to implementation details
- **Inconsistent Naming**: `packer` does packing AND unpacking, `matcher` does globbing
- **No Clear Entry Points**: Business operations scattered across packages

### 5. Testability Concerns

#### Issues:
- **Hard-coded Dependencies**: Direct instantiation of dependencies within functions
- **File System Coupling**: Direct OS calls make testing difficult
- **Missing Interfaces**: Key components lack interface definitions for mocking
- **Side Effects in Business Logic**: User prompts and file operations mixed with logic

## Proposed Architecture Improvements

### 1. Domain-Driven Package Structure

```
internal/
├── domain/           # Core business domain
│   ├── archive/     # Archive aggregate root
│   │   ├── archive.go       # Archive entity
│   │   ├── entry.go         # Archive entry value object
│   │   ├── format.go        # Archive format specifications
│   │   └── repository.go    # Archive repository interface
│   ├── file/        # File aggregate
│   │   ├── file.go          # File entity
│   │   ├── path.go          # Path value object
│   │   ├── metadata.go      # File metadata
│   │   └── specification.go # File specifications (size, type, etc.)
│   └── rules/       # Business rules
│       ├── filter.go        # Filtering rules interface
│       ├── pattern.go       # Pattern matching rules
│       └── policy.go        # Aggregation policies
│
├── application/     # Application services layer
│   ├── pack/       # Pack use case
│   │   ├── service.go       # Pack service
│   │   ├── command.go       # Pack command (CQRS pattern)
│   │   └── handler.go       # Command handler
│   ├── unpack/     # Unpack use case  
│   │   ├── service.go       # Unpack service
│   │   ├── command.go       # Unpack command
│   │   └── handler.go       # Command handler
│   └── ports/      # Application ports (interfaces)
│       ├── filesystem.go    # File system operations
│       ├── output.go        # Output writer interface
│       └── notification.go  # Progress notifications
│
├── infrastructure/  # Infrastructure implementations
│   ├── filesystem/  # File system adapter
│   │   ├── local.go         # Local file system
│   │   ├── walker.go        # Directory walker
│   │   └── watcher.go       # File change watcher
│   ├── archive/     # Archive format implementations
│   │   ├── text.go          # Text-based archive format
│   │   └── repository.go    # Archive storage
│   ├── filters/     # Filter implementations
│   │   ├── size.go          # Size filter
│   │   ├── binary.go        # Binary detection filter
│   │   ├── pattern.go       # Pattern matching filter
│   │   └── composite.go     # Composite filter
│   └── output/      # Output implementations
│       ├── file.go          # File output
│       ├── stdout.go        # Stdout output
│       └── buffer.go        # Buffer output (for testing)
│
├── interfaces/      # Interface adapters
│   ├── cli/        # CLI interface
│   │   ├── commands/
│   │   │   ├── pack.go
│   │   │   └── unpack.go
│   │   ├── root.go
│   │   └── printer.go      # CLI output formatting
│   └── api/        # Future API interface
│       └── handlers/
│
└── shared/         # Shared kernel
    ├── errors/     # Domain errors
    ├── events/     # Domain events
    └── values/     # Shared value objects
```

### 2. Clear Separation of Concerns

#### Domain Layer (internal/domain/)
- **Pure Business Logic**: No infrastructure dependencies
- **Rich Domain Models**: Entities with behavior, not just data
- **Domain Services**: Complex business operations
- **Specifications**: Business rule implementations

#### Application Layer (internal/application/)
- **Use Case Orchestration**: Coordinates domain operations
- **Command/Query Separation**: CQRS pattern for clarity
- **Port Definitions**: Interfaces for infrastructure
- **Transaction Boundaries**: Manages units of work

#### Infrastructure Layer (internal/infrastructure/)
- **Adapter Implementations**: Implements application ports
- **External System Integration**: File system, archives
- **Technical Concerns**: Performance, caching, persistence

#### Interface Layer (internal/interfaces/)
- **User Interface Adapters**: CLI commands
- **Input Validation**: User input validation
- **Output Formatting**: Presentation logic
- **Error Translation**: User-friendly error messages

### 3. Reduced Coupling Strategies

#### Dependency Inversion
```go
// Domain defines interface
package domain

type FileSystem interface {
    Walk(root Path, pattern Pattern) ([]File, error)
    Read(file File) (io.Reader, error)
    Write(file File, content io.Reader) error
}

// Infrastructure implements
package filesystem

type LocalFileSystem struct {
    // implementation
}

func (fs *LocalFileSystem) Walk(root domain.Path, pattern domain.Pattern) ([]domain.File, error) {
    // implementation
}
```

#### Event-Driven Communication
```go
// Domain events for decoupling
package events

type FileProcessed struct {
    File   domain.File
    Result ProcessResult
}

type ArchiveCreated struct {
    Archive domain.Archive
    Files   []domain.File
}
```

#### Interface Segregation
```go
// Specific interfaces for specific needs
type FileReader interface {
    Read(path Path) (io.Reader, error)
}

type FileWriter interface {
    Write(path Path, content io.Reader) error
}

type FileWalker interface {
    Walk(root Path, pattern Pattern) ([]File, error)
}
```

### 4. Improved Package Organization

#### Clear Dependency Rules
- **Domain → Nothing**: Domain has no external dependencies
- **Application → Domain**: Application uses domain
- **Infrastructure → Application, Domain**: Implements interfaces
- **Interfaces → Application**: Uses application services

#### Single Responsibility Packages
```go
// Each package has ONE clear purpose
package pack      // Handles packing operations ONLY
package unpack    // Handles unpacking operations ONLY  
package filter    // Handles filtering logic ONLY
package archive   // Handles archive format ONLY
```

#### Intuitive Naming
- `archive` instead of `packer` (clearer purpose)
- `filter` instead of `checkers` (more intuitive)
- `filesystem` instead of `walker` (describes what, not how)

### 5. Enhanced Testability

#### Dependency Injection
```go
// Service with injected dependencies
type PackService struct {
    fs         FileSystem
    archive    ArchiveRepository
    filters    FilterChain
    notifier   ProgressNotifier
}

func NewPackService(
    fs FileSystem,
    archive ArchiveRepository, 
    filters FilterChain,
    notifier ProgressNotifier,
) *PackService {
    return &PackService{
        fs:       fs,
        archive:  archive,
        filters:  filters,
        notifier: notifier,
    }
}
```

#### Mock-Friendly Interfaces
```go
// Easy to mock for testing
type FileSystem interface {
    Walk(root Path, pattern Pattern) ([]File, error)
}

// Test implementation
type MockFileSystem struct {
    WalkFunc func(Path, Pattern) ([]File, error)
}
```

#### Pure Functions
```go
// Pure function - easy to test
func ApplyFilter(files []File, filter Filter) []File {
    var result []File
    for _, file := range files {
        if filter.Matches(file) {
            result = append(result, file)
        }
    }
    return result
}
```

## Implementation Strategy

### Phase 1: Domain Model Creation
1. Create `internal/domain` package structure
2. Define core entities: Archive, File, Entry
3. Define value objects: Path, Pattern, Metadata
4. Create domain services and specifications

### Phase 2: Application Layer
1. Create `internal/application` package
2. Implement Pack and Unpack services
3. Define port interfaces
4. Implement command handlers

### Phase 3: Infrastructure Adapters
1. Create `internal/infrastructure` package
2. Implement file system adapter
3. Implement archive formats
4. Create filter implementations

### Phase 4: Interface Layer
1. Refactor CLI to use application services
2. Remove business logic from CLI
3. Implement proper error handling
4. Add progress notifications

### Phase 5: Migration and Cleanup
1. Migrate existing functionality
2. Remove old packages
3. Update tests
4. Update documentation

## Benefits of Proposed Architecture

### 1. Better Separation of Concerns
- Clear responsibility boundaries
- Single purpose packages
- No mixed concerns

### 2. Clearer Domain Boundaries
- Rich domain model
- Business logic centralized
- Technical details isolated

### 3. Reduced Coupling
- Dependency inversion
- Interface-based design
- Event-driven options

### 4. Intuitive Organization
- Logical package structure
- Clear naming conventions
- Obvious dependency flow

### 5. Improved Testability
- Mockable interfaces
- Pure functions
- Isolated components

## Risk Mitigation

### Risks
1. **Large Refactoring**: Significant code changes required
2. **Breaking Changes**: May affect existing functionality
3. **Learning Curve**: New architecture patterns

### Mitigation Strategies
1. **Incremental Migration**: Build new alongside old
2. **Comprehensive Testing**: Test at each phase
3. **Documentation**: Clear architecture documentation
4. **Parallel Structure**: Keep old code working during transition

## Success Criteria

1. **No package imports more than 3 other internal packages**
2. **Domain package has zero external dependencies**
3. **All business logic in domain/application layers**
4. **100% test coverage for domain logic**
5. **Clear, single-purpose packages**
6. **No circular dependencies**
7. **All infrastructure concerns isolated**

## Conclusion

The proposed architecture follows GO_BLUEPRINT principles of radical simplicity and deep interpretability while providing:
- Clear separation of concerns
- Intuitive package organization  
- Reduced coupling
- Enhanced testability
- Maintainable structure

This architecture will make the codebase more understandable and maintainable while preserving the established coding style and patterns.