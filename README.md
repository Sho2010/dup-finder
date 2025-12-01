# dup-finder

A cross-platform CLI tool for finding duplicate files across multiple directories. Performs pairwise comparison of directories, identifying files with the same name and optionally verifying content similarity using xxHash (non-cryptographic, high-performance hashing).

## Features

- **Pairwise comparison** of 2+ directories
- **Fast name-based** duplicate detection
- **Optional xxHash verification** for content matching (10-20x faster than SHA256)
- **Interactive deletion mode** for safe, guided duplicate removal
- **Parallel processing** for performance
- **Flexible filtering** (extensions, file size, depth)
- **Cross-platform support** (Linux, macOS, Windows)

## Installation

### Prerequisites

- Go 1.25.1 or higher

### Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd dup-finder

# Build the executable
go build

# (Optional) Install to your Go bin directory
go install
```

### Building with Taskfile (Recommended)

This project includes a [Taskfile](https://taskfile.dev) for easy building and testing.

#### Install Task

```bash
# Linux/macOS
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Or via Go
go install github.com/go-task/task/v3/cmd/task@latest

# macOS with Homebrew
brew install go-task/tap/go-task

# Ubuntu/Debian
sudo snap install task --classic
```

#### Common Build Commands

```bash
# Show all available tasks
task --list

# Build for current platform
task build

# Build Windows 64-bit binary
task build:windows

# Build for all platforms (Linux, macOS, Windows)
task build:all

# Build with tests
task check

# Clean build artifacts
task clean

# Run tests
task test

# Run tests with coverage
task test:coverage
```

#### WSL: Install to Windows

If you're using WSL, you can directly install the Windows binary to a Windows directory:

```bash
# Install to default location (C:\src\tools)
task install:windows

# Install to custom location
WIN_PATH=/mnt/d/my/tools task install:windows

# Check installation
task install:windows:check
```

This automatically:
- Builds the Windows binary
- Creates the target directory if it doesn't exist
- Copies with overwrite (`-f` flag)
- Shows both WSL and Windows paths

The Taskfile automatically:
- Embeds version information from git
- Optimizes binary size with `-ldflags="-s -w"`
- Creates platform-specific binaries in the `dist/` directory

### Manual Building (Platform-Specific)

#### Linux / macOS
```bash
# Build
go build -o dup-finder

# Run
./dup-finder /path/to/dir1 /path/to/dir2
```

#### Windows

**Using Taskfile (Recommended):**
```bash
task build:windows
```

**Manual Build - Command Prompt / PowerShell:**
```cmd
# Build
go build -o dup-finder.exe

# Run
dup-finder.exe C:\path\to\dir1 C:\path\to\dir2
```

**Manual Build - Git Bash / MSYS2:**
```bash
# Build
go build -o dup-finder.exe

# Run
./dup-finder.exe /c/path/to/dir1 /c/path/to/dir2
```

**Cross-compile from Linux/macOS:**
```bash
# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o dup-finder.exe

# Or use Taskfile
task build:windows
```

## Usage

### Windows PowerShell Script

A PowerShell script is included for easy comparison of `G:\var`, `N:\var`, and `M:\var` directories:

```powershell
# Navigate to the installation directory
cd C:\src\tools

# Basic comparison
.\run-compare.ps1

# With hash verification
.\run-compare.ps1 -CompareHash

# Filter by extensions and minimum size
.\run-compare.ps1 -Extensions '.jpg','.png' -MinSize 1048576

# Custom worker count
.\run-compare.ps1 -CompareHash -Workers 8

# Show help
.\run-compare.ps1 -Help
```

**Features:**
- Automatically checks which directories exist (G:\var, N:\var, M:\var)
- Requires at least 2 directories to compare
- Colored output for better readability
- All dup-finder options supported

### Basic Usage

```bash
# Find files with same name across two directories
dup-finder /path/to/dir1 /path/to/dir2

# Three directory pairwise comparison (a-b, a-c, b-c)
dup-finder /path/to/a /path/to/b /path/to/c

# Automatically skips non-existent directories
# If b doesn't exist, will compare a-c only
dup-finder /path/to/a /path/to/b /path/to/c
# Output: Warning: Skipping /path/to/b: ...
#         Comparing 2 out of 3 directories
```

### With Hash Comparison

Verify that files with the same name have identical content:

```bash
# Enable hash verification
dup-finder --compare-hash /dir1 /dir2

# Short form
dup-finder -H /dir1 /dir2
```

### With Filters

```bash
# Only compare specific file types
dup-finder -e .jpg,.png /photos1 /photos2

# Files larger than 1MB (1048576 bytes)
dup-finder -m 1048576 /dir1 /dir2

# Maximum depth of 2 levels
dup-finder -L 2 /dir1 /dir2

# Non-recursive (only root level)
dup-finder -r=false /dir1 /dir2
```

### Performance Tuning

```bash
# Use 8 workers for parallel processing
dup-finder -w 8 /large/dir1 /large/dir2

# Combine multiple options
dup-finder -H -w 16 -e .zip,.rar -m 1048576 /archives1 /archives2
```

## Interactive Deletion Mode

dup-finder includes an interactive mode for safely deleting duplicate files:

```bash
# Enable interactive mode
dup-finder dir1 dir2 --interactive

# With hash verification (recommended)
dup-finder dir1 dir2 --compare-hash --interactive
```

**Features:**
- Review each duplicate before deletion
- Choose which file to keep
- Batch deletion mode (for 2-directory comparison)
- Final confirmation before actual deletion
- Detailed summary with freed space

For complete documentation, see [INTERACTIVE_MODE.md](INTERACTIVE_MODE.md).

## Command-Line Options

| Flag | Long Form | Description | Default |
|------|-----------|-------------|---------|
| `-r` | `--recursive` | Search recursively in subdirectories | `true` |
| `-m` | `--min-size` | Minimum file size in bytes | `0` |
| `-e` | `--extensions` | Comma-separated file extensions (e.g., `.jpg,.png`) | `""` (all files) |
| `-L` | `--max-depth` | Maximum directory depth (-1 = unlimited) | `-1` |
| `-H` | `--compare-hash` | Enable SHA256 hash comparison | `false` |
| `-w` | `--workers` | Number of parallel workers | `NumCPU()` |
| `-i` | `--interactive` | Enable interactive deletion mode | `false` |

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
```

## Platform Support

### Supported Operating Systems

- Linux (Ubuntu, Debian, Fedora, Arch, etc.)
- macOS (10.15+)
- Windows (10+, Server 2016+)

### Windows Compatibility

dup-finder is fully compatible with Windows:

- Uses `filepath` package for cross-platform path handling
- Automatically handles backslash (`\`) path separators on Windows
- Works with both Command Prompt and PowerShell
- Supports long paths (260+ characters) on Windows 10 1607+
- UTF-8 unicode output for file names (ensure your terminal supports UTF-8)

**Windows Terminal Configuration:**

For best results with Unicode symbols (✓, ✗, ↔) in output, use:
- Windows Terminal (recommended)
- PowerShell 7+
- Git Bash with proper locale settings

If you see garbled characters, your terminal may not support UTF-8. The tool will still function correctly.

## Examples

### Example 1: Windows PowerShell - Compare Network Drives

```powershell
# Navigate to installation directory
cd C:\src\tools

# Basic comparison of G:\var, N:\var, M:\var
.\run-compare.ps1

# With hash verification
.\run-compare.ps1 -CompareHash

# Filter large media files only
.\run-compare.ps1 -Extensions '.mp4','.avi','.mkv' -MinSize 104857600
```

### Example 2: Compare Photo Directories

```bash
# Find duplicate photos by name
dup-finder ~/Photos/2023 ~/Photos/2024

# Verify content is identical (recommended for critical files)
dup-finder -H ~/Photos/2023 ~/Photos/2024
```

### Example 3: Find Large Duplicate Archives

```bash
# Compare only .zip and .rar files larger than 10MB
dup-finder -e .zip,.rar -m 10485760 /backups/old /backups/new
```

### Example 4: Compare Multiple Backup Directories

```bash
# Three-way comparison
dup-finder /backup/daily /backup/weekly /backup/monthly
```

## Algorithm

### Two-Phase Processing

**Phase 1: Name-based scan (always)**
- Scan all directories in parallel
- Group files by basename
- Fast, no disk I/O for content reading

**Phase 2: Hash verification (optional, `--compare-hash`)**
- Only for files with matching names
- Parallel xxHash computation (10-20x faster than SHA256)
- Determines if content is truly identical
- Non-cryptographic but excellent collision resistance

### Performance

Typical performance on modern hardware (SSD) with xxHash:

- 1,000 files (name-only): < 0.5 seconds
- 1,000 files (with hash): < 2 seconds
- 10,000 files (with hash): < 15 seconds

**Performance Improvement (vs SHA256):**
- Small files (1MB): ~10x faster
- Large files (100MB+): ~15-20x faster

Performance varies based on:
- Disk speed (SSD vs HDD)
- File sizes
- Number of workers
- CPU cores available

## Development

### Running Tests

```bash
# All tests
go test -v ./...

# Specific package
go test -v ./internal/finder

# With coverage
go test -v -cover ./...

# Integration tests only
go test -v -run Integration
```

### Project Structure

```
dup-finder/
├── main.go                           # Entry point
├── cmd/
│   └── root.go                       # CLI orchestration
├── internal/
│   ├── models/
│   │   └── models.go                 # Data structures
│   ├── scanner/
│   │   ├── scanner.go                # Directory scanning
│   │   └── worker.go                 # Worker pool
│   ├── finder/
│   │   ├── finder.go                 # Duplicate detection
│   │   └── comparator.go             # Hash computation
│   ├── interactive/
│   │   ├── session.go                # Interactive mode orchestration
│   │   ├── ui.go                     # User interface
│   │   └── deleter.go                # Safe file deletion
│   └── output/
│       └── formatter.go              # Output formatting
└── integration_test.go               # End-to-end tests
```

For detailed development documentation, see [claude.md](claude.md).

## Troubleshooting

### Common Issues

**"directory does not exist"**
- Ensure all directory paths are valid
- On Windows, use absolute paths or escape backslashes properly
- Example: `C:\Users\Name\Documents` or `C:/Users/Name/Documents`

**Slow performance**
- Increase workers: `--workers 16`
- Use filters to reduce scope: `-e .jpg,.png`
- Skip small files: `--min-size 1048576`

**Memory issues**
- Process fewer directories at once
- Use filters to reduce file count
- Increase available system memory

**Unicode characters not displaying correctly (Windows)**
- Use Windows Terminal or PowerShell 7+
- Set console encoding: `chcp 65001` (UTF-8)
- The tool still works correctly even if symbols don't display

## License

[Add your license information here]

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass: `go test ./...`
2. Code follows Go conventions
3. Test coverage remains >80%
4. Documentation is updated

## Contact

[Add contact information here]
