# dup-finder comparison script
# Compares G:\var, N:\var, and M:\var directories

param(
    [switch]$CompareHash,
    [string[]]$Extensions,
    [int]$MinSize = 0,
    [int]$MaxDepth = -1,
    [switch]$NonRecursive,
    [int]$Workers = 0,
    [switch]$Interactive,
    [switch]$Help
)

# Show help
if ($Help) {
    Write-Host "dup-finder - Compare G:\var, N:\var, and M:\var directories"
    Write-Host ""
    Write-Host "Usage: .\run-compare.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -CompareHash      Enable SHA256 hash comparison"
    Write-Host "  -Extensions       Comma-separated file extensions (e.g., '.txt,.jpg')"
    Write-Host "  -MinSize <bytes>  Minimum file size in bytes (default: 0)"
    Write-Host "  -MaxDepth <n>     Maximum directory depth (default: -1, unlimited)"
    Write-Host "  -NonRecursive     Disable recursive search"
    Write-Host "  -Workers <n>      Number of parallel workers (default: auto)"
    Write-Host "  -Interactive      Enable interactive deletion mode"
    Write-Host "  -Help             Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\run-compare.ps1"
    Write-Host "  .\run-compare.ps1 -CompareHash"
    Write-Host "  .\run-compare.ps1 -Extensions '.jpg,.png' -MinSize 1048576"
    Write-Host "  .\run-compare.ps1 -CompareHash -Workers 8"
    Write-Host "  .\run-compare.ps1 -CompareHash -Interactive"
    exit 0
}

# Target directories
$dir1 = "G:\var"
$dir2 = "N:\var"
$dir3 = "M:\var"

# Check if dup-finder.exe exists
$exePath = "C:\src\tools\dup-finder.exe"
if (-not (Test-Path $exePath)) {
    Write-Host "Error: dup-finder.exe not found at $exePath" -ForegroundColor Red
    Write-Host "Please run 'task install:windows' from WSL to install" -ForegroundColor Yellow
    exit 1
}

# Check if directories exist
$dirsToCheck = @($dir1, $dir2, $dir3)
$existingDirs = @()

foreach ($dir in $dirsToCheck) {
    if (Test-Path $dir) {
        $existingDirs += $dir
        Write-Host "✓ Found: $dir" -ForegroundColor Green
    } else {
        Write-Host "✗ Not found: $dir" -ForegroundColor Yellow
    }
}

if ($existingDirs.Count -lt 2) {
    Write-Host ""
    Write-Host "Error: Need at least 2 directories to compare" -ForegroundColor Red
    Write-Host "Found only $($existingDirs.Count) directories" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Comparing $($existingDirs.Count) directories..." -ForegroundColor Cyan
Write-Host ""

# Build command arguments
$args = @()
$args += $existingDirs

if ($CompareHash) {
    $args += "--compare-hash"
    Write-Host "Option: Hash comparison enabled" -ForegroundColor Cyan
}

if ($Extensions) {
    $args += "--extensions"
    $args += $Extensions -join ","
    Write-Host "Option: Extensions filter = $($Extensions -join ',')" -ForegroundColor Cyan
}

if ($MinSize -gt 0) {
    $args += "--min-size"
    $args += $MinSize.ToString()
    Write-Host "Option: Minimum file size = $MinSize bytes" -ForegroundColor Cyan
}

if ($MaxDepth -ge 0) {
    $args += "--max-depth"
    $args += $MaxDepth.ToString()
    Write-Host "Option: Maximum depth = $MaxDepth" -ForegroundColor Cyan
}

if ($NonRecursive) {
    $args += "--recursive=false"
    Write-Host "Option: Non-recursive mode" -ForegroundColor Cyan
}

if ($Workers -gt 0) {
    $args += "--workers"
    $args += $Workers.ToString()
    Write-Host "Option: Workers = $Workers" -ForegroundColor Cyan
}

if ($Interactive) {
    $args += "--interactive"
    Write-Host "Option: Interactive deletion mode enabled" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Execute dup-finder
& $exePath $args

# Check exit code
if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "✓ Comparison completed successfully" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "✗ Comparison failed with exit code: $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}
