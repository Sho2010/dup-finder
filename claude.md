# dup-finder Development Documentation

## Project Overview

dup-finder is a CLI tool for finding duplicate files across multiple directories. It performs pairwise comparison of directories, identifying files with the same name and optionally verifying content similarity using SHA256 hashing.

## Development History

### Initial Implementation (Phase 1)
- Created basic CLI tool using Cobra framework
- Single directory scanning with hash-based duplicate detection
- Implemented filters: extensions, min-size, max-depth, recursive
- Added comprehensive test coverage with testify

### Major Refactoring (Phase 2)
The tool was completely refactored to support multi-directory comparison:

**Requirements:**
1. Accept 2+ directories as arguments
2. Compare directories pairwise (a-b, a-c, b-c)
3. Find files with same name across directory pairs
4. Optional hash comparison (--compare-hash, default: false)
5. Parallel processing for performance
6. Changed default extensions from `[.zip, .avi, .mp4]` to `[]` (all files)

**Architecture Decision:**
- Moved from monolithic `main.go` to multi-package structure
- Implemented Worker Pool pattern for parallelization
- Separated concerns: scanning, finding, comparison, output

### Interactive Mode & Performance Improvements (Phase 3)

**Hash Algorithm Change (commit 82392d9):**
- Replaced SHA256 with xxHash for significant performance improvement
- xxHash is 10-20x faster for file comparison use cases
- Still maintains high collision resistance for duplicate detection
- Trade-off: xxHash is non-cryptographic (acceptable for duplicate detection)

**Interactive Deletion Mode (commits 5b8a649, 582ec11, 40a782d):**
- Added `--interactive` / `-i` flag for safe, interactive file deletion
- Features:
  - Review each duplicate set before deletion
  - Choose which file to keep (multiple options)
  - Batch deletion mode for 2-directory comparisons
  - On-demand hash verification during interactive session
  - Early finish option to skip remaining files
  - Final confirmation with deletion summary and space freed
  - Comprehensive error handling and rollback
- New package: `internal/interactive/` with session, ui, and deleter components
- See [INTERACTIVE_MODE.md](INTERACTIVE_MODE.md) for detailed usage

## Project Structure

```
dup-finder/
├── main.go                           # Entry point (calls cmd.Execute())
├── cmd/
│   └── root.go                       # Cobra CLI setup and orchestration
├── internal/
│   ├── models/
│   │   └── models.go                 # Data structures (FileInfo, ScanOptions, etc.)
│   ├── scanner/
│   │   ├── scanner.go                # Directory scanning with filtering
│   │   └── worker.go                 # Worker Pool implementation
│   ├── finder/
│   │   ├── finder.go                 # Duplicate detection logic
│   │   └── comparator.go             # xxHash computation
│   ├── interactive/
│   │   ├── session.go                # Interactive session orchestration
│   │   ├── ui.go                     # User interface prompts
│   │   └── deleter.go                # Safe file deletion with validation
│   └── output/
│       └── formatter.go              # Output formatting (Formatter interface)
├── internal/*/test.go files          # Unit tests
├── integration_test.go               # End-to-end tests
├── INTERACTIVE_MODE.md               # Interactive mode documentation
├── Taskfile.yml                      # Build and test automation
├── run-compare.ps1                   # Windows PowerShell helper script
├── go.mod
└── go.sum
```

## Key Components

### 1. Models (`internal/models/models.go`)

**FileInfo**: Represents a scanned file
```go
type FileInfo struct {
    Path      string    // Full path
    Directory string    // Root directory
    Size      int64     // File size in bytes
    ModTime   time.Time // Modification time
    Hash      string    // SHA256 hash (computed lazily)
}
```

**ScanOptions**: Configuration for scanning
```go
type ScanOptions struct {
    Directories []string // Directories to scan
    Recursive   bool     // Search recursively
    MinSize     int64    // Minimum file size
    Extensions  []string // File extensions filter
    MaxDepth    int      // Maximum depth (-1 = unlimited)
    CompareHash bool     // Enable hash comparison
    NumWorkers  int      // Parallel workers
}
```

**PairComparison**: Result of comparing two directories
```go
type PairComparison struct {
    Dir1    string      // First directory
    Dir2    string      // Second directory
    Matches []FileMatch // Files with same name
}
```

**FileMatch**: Pair of files with same name
```go
type FileMatch struct {
    Filename    string   // Base filename
    File1       FileInfo // File from first directory
    File2       FileInfo // File from second directory
    HashChecked bool     // Whether hash was compared
    HashMatch   bool     // Whether hashes match
}
```

**DuplicateSet**: Group of duplicate files (for interactive mode)
```go
type DuplicateSet struct {
    ID          int        // Set identifier
    Files       []FileInfo // All duplicate file instances
    HashChecked bool       // Whether hashes were verified
}
```

**UserAction**: User decision in interactive mode
```go
type UserAction struct {
    Action     string // "skip", "delete", "keep_dir1", "keep_dir2", "compute_hash", "finish"
    DeleteFile string // Path of file to delete
}
```

**SessionSummary**: Results of interactive session
```go
type SessionSummary struct {
    TotalSets       int         // Number of duplicate sets processed
    FilesDeleted    int         // Number of files deleted
    SpaceFreed      int64       // Total bytes freed
    SuccessfulOps   []string    // Successfully deleted files
    FailedOps       []FailedOp  // Failed deletion attempts
}
```

### 2. Scanner (`internal/scanner/`)

**Scanner**: Scans directories with filtering
- `Scan(directory)`: Scans single directory
- `ScanAll()`: Scans all directories in parallel
- Applies filters: extensions, minSize, maxDepth

**WorkerPool**: Manages parallel file processing
- Channel-based job distribution
- Configurable number of workers
- Used for both scanning and hash computation

### 3. Finder (`internal/finder/`)

**Finder**: Detects duplicate files
- `ComparePair(dir1Files, dir2Files)`: Compares two directories
- Groups files by basename
- Finds intersection of filenames
- Optionally computes and compares hashes

**Comparator**: Hash computation (using xxHash)
- `CalculateFileHash(path)`: Computes xxHash (64-bit) hash
- `ComputeHashesParallel(files, workers)`: Parallel hash computation
- xxHash is ~10-20x faster than SHA256 for duplicate detection
- Non-cryptographic but highly collision-resistant

**Helper Functions:**
- `GeneratePairs(dirs)`: Generates all directory pairs
- `groupByName(files)`: Maps basename → FileInfo
- `findCommonFiles(group1, group2)`: Finds intersection

### 4. Output (`internal/output/`)

**Formatter Interface**: Defines output formatting
```go
type Formatter interface {
    FormatPairComparison(comparison PairComparison) string
}
```

**SimpleFormatter**: Text-based output
- Formats pairwise comparisons
- Shows hash comparison results if enabled
- Example output:
```
=== /path/to/a ↔ /path/to/b ===
photo.jpg:        ✓ [Hash: ✓ Identical]
document.pdf:     ✓ [Hash: ✗ Different]
```

**Design Note**: Formatter interface allows easy addition of new output formats (JSON, CSV, Table) without modifying core logic.

### 5. Interactive (`internal/interactive/`)

**Session**: Orchestrates interactive workflow
- `RunInteractiveSession(comparisons, opts)`: Main entry point
- Converts PairComparison to DuplicateSet
- Manages user interaction loop
- Handles batch deletion mode (for 2-directory comparisons)
- Supports on-demand hash verification
- Early finish option to skip remaining files
- Collects and executes deletion actions

**UI**: User interface components
- `DisplayDuplicateSet(set)`: Shows file information with formatting
- `PromptUserAction(set, allowBatchByDir)`: Gets user choice
- Options: skip, delete specific file, batch by directory, compute hash, finish, quit
- Color-coded output for better readability
- Human-readable file sizes

**Deleter**: Safe file deletion
- `ValidateFileDeletion(path)`: Pre-deletion checks
  - File exists
  - Is a regular file (not directory/symlink)
  - Parent directory is writable
- `ShowDeletionSummary(actions)`: Final confirmation screen
- `ExecuteDeletions(actions)`: Performs actual deletions
- Returns detailed SessionSummary with success/failure info

### 6. CLI (`cmd/root.go`)

**Orchestration Flow:**
1. Validate directories exist (skip non-existent ones with warning)
2. Build ScanOptions from flags
3. Scan all directories in parallel (Scanner.ScanAll)
4. Generate directory pairs (Finder.GeneratePairs)
5. Compare each pair (Finder.ComparePair)
6. **If interactive mode enabled:**
   - Run interactive session (interactive.RunInteractiveSession)
   - Display summary and exit
7. **If non-interactive:**
   - Format and print output (output.FormatAllComparisons)

**Flags:**
- `-r, --recursive`: Recursive search (default: true)
- `-m, --min-size`: Minimum file size in bytes (default: 0)
- `-e, --extensions`: File extensions filter (default: [])
- `-L, --max-depth`: Maximum depth (default: -1)
- `-H, --compare-hash`: Enable hash comparison (default: false)
- `-w, --workers`: Number of workers (default: NumCPU())
- `-i, --interactive`: Enable interactive deletion mode (default: false)

## Algorithm Details

### Pairwise Comparison
For N directories, generates C(N,2) pairs:
```go
func GeneratePairs(dirs []string) [][2]string {
    var pairs [][2]string
    for i := 0; i < len(dirs); i++ {
        for j := i + 1; j < len(dirs); j++ {
            pairs = append(pairs, [2]string{dirs[i], dirs[j]})
        }
    }
    return pairs
}
```

### Two-Phase Processing

**Phase 1: Name-based scan (always)**
- Scan all directories in parallel
- Group files by basename
- Fast, no hash computation

**Phase 2: Hash verification (optional)**
- Only for files with matching names
- Parallel xxHash computation with I/O-optimized worker count
- Compare hashes to determine content identity
- xxHash provides 10-20x speedup over SHA256 while maintaining reliability

### Interactive Mode Workflow

**Phase 1: Detection and filtering**
- Name-based scan first (same as non-interactive)
- Optionally pre-compute hashes with `--compare-hash`
- Convert PairComparison to DuplicateSet (groups all instances)
- Filter by file size matching

**Phase 2: User interaction**
- Display each duplicate set with details
- User chooses action for each set:
  - Skip (no action)
  - Delete specific file(s)
  - Batch mode: delete all from one directory (2-dir only)
  - On-demand hash verification (if not pre-computed)
  - Early finish (proceed with selected files)
  - Quit without changes

**Phase 3: Execution**
- Show deletion summary with total space to be freed
- Final confirmation prompt
- Validate each file before deletion
- Execute deletions with error handling
- Display detailed summary (success/failures/space freed)

### Parallel Processing Strategy

**Directory Scanning:**
- One goroutine per directory
- Each uses Worker Pool for file processing
- Default workers: `runtime.NumCPU()`

**Hash Computation:**
- Separate Worker Pool with `NumCPU() * 2` workers
- I/O bound, so use more workers than CPU cores
- Process pairs independently

## Testing Strategy

### Unit Tests

**comparator_test.go:**
- xxHash calculation correctness
- Same content produces same hash
- Different content produces different hash
- Parallel hash computation
- Error handling

**formatter_test.go:**
- Output formatting with/without hash
- Empty results handling
- Multiple comparison formatting

**interactive/session_test.go:**
- DuplicateSet conversion
- User action handling
- Batch deletion mode
- On-demand hash verification
- Session summary generation

**interactive/deleter_test.go:**
- File validation (existence, permissions, type)
- Deletion execution
- Error handling
- Summary generation

**interactive/ui_test.go:**
- Human-readable size formatting
- Display formatting

### Integration Tests

**integration_test.go:**
- Two directory comparison
- Three directory pairwise comparison
- Hash comparison accuracy
- Extension filtering
- No common files scenario
- Same name, different content
- Pair generation logic
- Cross-platform path handling (Windows compatibility)
- Unicode filename support

**Test Coverage:**
- Overall: >80%
- Critical paths (hash, comparison, deletion): 100%

## Performance Considerations

### Optimization Techniques
1. **xxHash Algorithm**: 10-20x faster than SHA256 for file comparison
2. **Lazy Hash Computation**: Only compute when needed (--compare-hash or interactive on-demand)
3. **Parallel Scanning**: All directories scanned concurrently
4. **Worker Pools**: Efficient parallel file processing
5. **Channel Buffering**: Balanced buffer sizes for throughput
6. **I/O Optimization**: More workers for I/O-bound hash computation
7. **Size-based Pre-filtering**: In interactive mode, filter by size before hash computation

### Performance Targets
- 1000 files across 2 directories: < 0.5s (name-based)
- 1000 files across 2 directories: < 2s (with xxHash)
- 10000 files across 3 directories: < 15s (with xxHash)
- Memory usage: < 100MB for 10,000 files

### Performance Comparison (xxHash vs SHA256)
- Small files (1MB): xxHash ~10x faster
- Large files (100MB+): xxHash ~15-20x faster
- Collision resistance: Both excellent for duplicate detection
- Trade-off: xxHash is non-cryptographic (acceptable for this use case)

## Usage Examples

### Basic Usage
```bash
# Find files with same name across directories
./dup-finder /path/to/dir1 /path/to/dir2

# Three directory comparison
./dup-finder /path/to/a /path/to/b /path/to/c
```

### With Hash Comparison
```bash
# Verify content is actually identical
./dup-finder --compare-hash /dir1 /dir2

# Short form
./dup-finder -H /dir1 /dir2
```

### With Filters
```bash
# Only compare .jpg and .png files
./dup-finder -e .jpg,.png /photos1 /photos2

# Files larger than 1MB
./dup-finder -m 1048576 /dir1 /dir2

# Maximum depth of 2 levels
./dup-finder -L 2 /dir1 /dir2
```

### Performance Tuning
```bash
# Use 8 workers
./dup-finder -w 8 /large/dir1 /large/dir2

# Combine options
./dup-finder -H -w 16 -e .zip,.rar -m 1048576 /archives1 /archives2
```

### Interactive Mode
```bash
# Basic interactive deletion
./dup-finder --interactive /dir1 /dir2

# With pre-computed hashes (recommended)
./dup-finder -H --interactive /dir1 /dir2

# Interactive with filtering
./dup-finder -i -e .jpg,.png -m 1048576 /photos1 /photos2

# Multiple directories with interactive mode
./dup-finder -i /backup1 /backup2 /backup3
```

## Output Format

### Without Hash Comparison
```
=== /path/to/a ↔ /path/to/b ===
photo.jpg:           ✓
document.pdf:        ✓

=== /path/to/a ↔ /path/to/c ===
video.mp4:           ✓

=== /path/to/b ↔ /path/to/c ===
(No duplicates)
```

### With Hash Comparison
```
=== /path/to/a ↔ /path/to/b ===
photo.jpg:           ✓ [Hash: ✓ Identical]
document.pdf:        ✓ [Hash: ✗ Different]

=== /path/to/a ↔ /path/to/c ===
video.mp4:           ✓ [Hash: ✓ Identical]

=== /path/to/b ↔ /path/to/c ===
(No duplicates)
```

### Interactive Mode Output
```
Duplicate Set #1
┌─────────────────────────────────────────────────────
│ File 1: /path/to/dir1/photo.jpg
│   Size: 2.5 MB
│   Modified: 2024-01-15 14:30:00
│   Hash: a1b2c3d4e5f67890
│
│ File 2: /path/to/dir2/photo.jpg
│   Size: 2.5 MB
│   Modified: 2024-01-15 14:30:00
│   Hash: a1b2c3d4e5f67890
└─────────────────────────────────────────────────────

Choose an action:
  [s] Skip
  [1] Keep file 1, delete file 2
  [2] Keep file 2, delete file 1
  [a] Keep all from dir1 (batch mode)
  [b] Keep all from dir2 (batch mode)
  [h] Compute hash (if not already computed)
  [f] Finish and proceed with current selections
  [q] Quit

Action: 1

=== Deletion Summary ===
Files to delete: 5
Total space to free: 125.3 MB

1. /path/to/dir2/photo.jpg (2.5 MB)
2. /path/to/dir2/video.mp4 (50.0 MB)
...

Proceed with deletion? [y/N]: y

Deleting files...
✓ Deleted: /path/to/dir2/photo.jpg
✓ Deleted: /path/to/dir2/video.mp4
...

=== Session Summary ===
Duplicate sets processed: 10
Files deleted: 5
Space freed: 125.3 MB
Successful: 5
Failed: 0
```

## Development Notes

### Breaking Changes from V1
1. **Minimum arguments**: Now requires 2+ directories (was 0-1)
2. **Output format**: Changed from hash-grouped list to pairwise comparison
3. **Default extensions**: Changed from `[.zip,.avi,.mp4]` to `[]` (all files)

### Breaking Changes in Phase 3
1. **Hash algorithm**: Changed from SHA256 to xxHash (non-cryptographic)
   - Old hashes are not compatible
   - Significantly faster but non-cryptographic
   - Still excellent collision resistance for duplicate detection

### Design Decisions

**Why pairwise comparison?**
- Clear visualization of which directories share files
- Easier to identify unique files per directory
- Scales well with 3-4 directories
- Better than N-way matrix for readability

**Why lazy hash computation?**
- Significant performance improvement for name-only use case
- Users can quickly identify potential duplicates
- Hash verification as opt-in for confirmation
- Reduces unnecessary disk I/O

**Why Formatter interface?**
- Easy to add new output formats (JSON, CSV, table)
- Separation of concerns
- Testable in isolation
- Future extensibility without modifying core logic

**Why xxHash over SHA256?**
- Duplicate detection doesn't require cryptographic security
- 10-20x performance improvement for large files
- Excellent collision resistance for practical use
- Lower CPU usage, faster user feedback
- Trade-off: Cannot be used for security verification

**Why separate interactive package?**
- Clear separation between detection and deletion logic
- Interactive mode has different data flow (DuplicateSet vs PairComparison)
- Independent testing of UI, deletion, and session management
- Non-interactive mode remains simple and fast
- Easy to maintain and extend independently

**Why size-based pre-filtering in interactive mode?**
- Files with different sizes cannot be identical
- Reduces unnecessary hash computation
- Improves interactive session startup time
- User sees only actual duplicates (content-wise)

### Future Enhancements

**V3.1 - Interactive Mode Improvements:**
- Preview mode (dry-run without actual deletion)
- Undo last deletion operation
- Save/load decision sets for later execution
- Smart suggestions (older files, smaller quality, etc.)
- Regex filtering in interactive session

**V3.2 - Performance:**
- Incremental scanning with cache
- Progress bar for large directories
- Rate limiting for disk I/O
- Resume interrupted scans

**V3.3 - Features:**
- Regex pattern matching for filenames
- Exclude patterns (like .gitignore)
- Checksum file output/input
- Symlink handling options
- Duplicate file move (instead of delete)

**V3.4 - Advanced:**
- Network directory support (SMB, NFS)
- Compression-aware comparison
- Binary comparison mode (beyond hash)
- Plugin system for custom comparators
- Similarity detection (fuzzy matching)

**V3.5 - Output Formats:**
- JSON output: `--format json`
- CSV output: `--format csv`
- Table output with borders: `--format table`
- HTML report: `--format html`

## Platform Support

### Cross-Platform Compatibility

dup-finder is fully compatible with:
- **Linux** (all major distributions)
- **macOS** (10.15+)
- **Windows** (10+, Server 2016+)

### Windows Compatibility Details

**Path Handling:**
- Uses `path/filepath` package throughout for cross-platform path handling
- Automatically handles backslash (`\`) and forward slash (`/`) separators
- Correctly uses `filepath.Separator` for depth calculation
- Supports long paths (260+ characters) on Windows 10 1607+

**File Operations:**
- All file I/O uses standard library (`os`, `io`, `crypto`)
- No platform-specific system calls or shell commands
- File permissions handled correctly across platforms
- UTF-8 unicode support for filenames (Japanese, Chinese, emoji, etc.)

**Testing:**
The integration test suite includes Windows-specific tests:
- `TestCrossPlatformPaths`: Verifies correct path separator handling
- `TestWindowsStylePaths`: Tests nested paths and filenames with spaces
- `TestUnicodeFilenames`: Validates international character support
- `TestCaseInsensitiveExtensions`: Ensures case-insensitive extension filtering

**Windows Terminal Recommendations:**
- Use Windows Terminal or PowerShell 7+ for best UTF-8 symbol display (✓, ✗, ↔)
- If symbols don't display correctly, the tool still functions properly
- Set console to UTF-8: `chcp 65001`

### Building on Windows

```cmd
# Command Prompt or PowerShell
go build -o dup-finder.exe

# Run
dup-finder.exe C:\path\to\dir1 C:\path\to\dir2
```

```bash
# Git Bash or MSYS2
go build -o dup-finder.exe

# Run with forward slashes
./dup-finder.exe /c/path/to/dir1 /c/path/to/dir2
```

## Dependencies

```
github.com/spf13/cobra v1.10.1       # CLI framework
github.com/stretchr/testify v1.11.1  # Testing utilities
github.com/cespare/xxhash/v2 v2.3.0  # Fast non-cryptographic hash function
```

### Dependency Rationale

**cobra**: Industry-standard CLI framework
- Rich flag handling and subcommand support
- Auto-generated help and usage documentation
- POSIX-compliant flag parsing

**testify**: Comprehensive testing toolkit
- Assertion helpers for cleaner tests
- Mock support for future extensibility
- Suite support for integration tests

**xxhash**: High-performance hash algorithm
- 10-20x faster than SHA256 for large files
- Excellent collision resistance (64-bit)
- Production-ready (used by many databases and tools)
- Zero dependencies itself

## Building and Testing

### Using Taskfile (Recommended)

The project includes a [Taskfile](https://taskfile.dev) for streamlined building and testing:

```bash
# Build for current platform
task build

# Build Windows binary
task build:windows

# Build for all platforms (Linux, macOS Intel/ARM, Windows)
task build:all

# Run all tests
task test

# Run tests with coverage report
task test:coverage

# Run tests with race detector
task test:race

# Run integration tests only
task test:integration

# Format code
task fmt

# Run static analysis
task vet

# Clean build artifacts
task clean

# Show all available tasks
task --list
```

**Taskfile Features:**
- Automatic version embedding from git tags/commits
- Build optimization with `-ldflags="-s -w"` (reduces binary size)
- Incremental builds based on source file changes
- Cross-platform build support
- Integrated testing and linting

### Manual Build Commands

#### Build
```bash
# Standard build
go build

# Optimized build with version info
go build -ldflags="-s -w" -o dup-finder
```

#### Cross-Platform Builds
```bash
# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o dup-finder.exe

# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o dup-finder-linux

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o dup-finder-darwin-amd64

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o dup-finder-darwin-arm64
```

### Run Tests
```bash
# All tests
go test -v ./...

# Specific package
go test -v ./internal/finder

# With coverage
go test -v -cover ./...

# Benchmarks
go test -bench=. ./...
```

### Install
```bash
go install
```

## Troubleshooting

### Common Issues

**"directory does not exist"**
- Ensure all directory paths are valid
- Use absolute paths or correct relative paths

**Slow performance**
- Increase workers: `--workers 16`
- Reduce scope: use `-e` to filter extensions
- Use `--min-size` to skip small files

**Memory issues**
- Reduce number of directories processed at once
- Increase system memory
- Use filters to reduce file count

## Contributing

When adding new features:
1. Update models if new data structures needed
2. Add tests with >80% coverage
3. Update this documentation
4. Ensure all existing tests pass
5. Follow Go conventions and best practices

## License

MIT License - see [LICENSE](LICENSE) file for details
