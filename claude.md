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
│   │   └── comparator.go             # SHA256 hash computation
│   └── output/
│       └── formatter.go              # Output formatting (Formatter interface)
├── internal/finder/comparator_test.go
├── internal/output/formatter_test.go
├── integration_test.go               # End-to-end tests
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

**Comparator**: Hash computation
- `CalculateFileHash(path)`: Computes SHA256 hash
- `ComputeHashesParallel(files, workers)`: Parallel hash computation

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

### 5. CLI (`cmd/root.go`)

**Orchestration Flow:**
1. Validate directories exist
2. Build ScanOptions from flags
3. Scan all directories in parallel (Scanner.ScanAll)
4. Generate directory pairs (Finder.GeneratePairs)
5. Compare each pair (Finder.ComparePair)
6. Format and print output (output.FormatAllComparisons)

**Flags:**
- `-r, --recursive`: Recursive search (default: true)
- `-m, --min-size`: Minimum file size in bytes (default: 0)
- `-e, --extensions`: File extensions filter (default: [])
- `-L, --max-depth`: Maximum depth (default: -1)
- `-H, --compare-hash`: Enable hash comparison (default: false)
- `-w, --workers`: Number of workers (default: NumCPU())

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
- Parallel hash computation with I/O-optimized worker count
- Compare hashes to determine content identity

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
- Hash calculation correctness
- Same content produces same hash
- Different content produces different hash
- Parallel hash computation
- Error handling

**formatter_test.go:**
- Output formatting with/without hash
- Empty results handling
- Multiple comparison formatting

### Integration Tests

**integration_test.go:**
- Two directory comparison
- Three directory pairwise comparison
- Hash comparison accuracy
- Extension filtering
- No common files scenario
- Same name, different content
- Pair generation logic

**Test Coverage:**
- Overall: >80%
- Critical paths (hash, comparison): 100%

## Performance Considerations

### Optimization Techniques
1. **Lazy Hash Computation**: Only compute when needed (--compare-hash)
2. **Parallel Scanning**: All directories scanned concurrently
3. **Worker Pools**: Efficient parallel file processing
4. **Channel Buffering**: Balanced buffer sizes for throughput
5. **I/O Optimization**: More workers for I/O-bound hash computation

### Performance Targets
- 1000 files across 2 directories: < 1s (name-based)
- 1000 files across 2 directories: < 5s (with hash)
- 10000 files across 3 directories: < 30s (with hash)
- Memory usage: < 100MB for 10,000 files

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

## Development Notes

### Breaking Changes from V1
1. **Minimum arguments**: Now requires 2+ directories (was 0-1)
2. **Output format**: Changed from hash-grouped list to pairwise comparison
3. **Default extensions**: Changed from `[.zip,.avi,.mp4]` to `[]` (all files)

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

### Future Enhancements

**V2.1 - Performance:**
- Incremental scanning with cache
- Progress bar for large directories
- Rate limiting for disk I/O

**V2.2 - Features:**
- Regex pattern matching for filenames
- Exclude patterns (like .gitignore)
- Interactive mode for duplicate resolution
- Checksum file output/input

**V2.3 - Advanced:**
- Network directory support (SMB, NFS)
- Compression-aware comparison
- Binary comparison mode (beyond hash)
- Plugin system for custom comparators

**V2.4 - Output Formats:**
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
```

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

[Add your license information here]

## Contact

[Add contact information here]
