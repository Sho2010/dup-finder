package models

import "time"

// FileInfo represents information about a scanned file
type FileInfo struct {
	Path      string    // Full path to the file
	Directory string    // Root directory this file belongs to
	Size      int64     // File size in bytes
	ModTime   time.Time // Modification time
	Hash      string    // xxHash hash (computed lazily)
}

// ScanOptions contains configuration for file scanning
type ScanOptions struct {
	Directories []string // Directories to scan
	Recursive   bool     // Search directories recursively
	MinSize     int64    // Minimum file size in bytes to consider
	Extensions  []string // File extensions to filter (empty = all files)
	MaxDepth    int      // Maximum directory depth (-1 = unlimited)
	CompareHash bool     // Whether to compare file content using hash
	NumWorkers  int      // Number of parallel workers
}

// PairComparison represents the result of comparing two directories
type PairComparison struct {
	Dir1    string      // First directory path
	Dir2    string      // Second directory path
	Matches []FileMatch // Files that match by name
}

// FileMatch represents a pair of files with the same name
type FileMatch struct {
	Filename    string   // Base filename
	File1       FileInfo // File from first directory
	File2       FileInfo // File from second directory
	HashChecked bool     // Whether hash comparison was performed
	HashMatch   bool     // Whether hashes match (only meaningful if HashChecked)
}

// DuplicateSet represents files that are duplicates based on hash
type DuplicateSet struct {
	ID    int        // Sequential ID for display
	Files []FileInfo // All duplicate files
	Hash  string     // Common hash value
}

// UserAction represents the user's decision
type UserAction struct {
	Action          string // "skip", "delete", or "batch_delete_by_dir"
	KeepFile        string // Path of file to keep (for delete action)
	DeleteFile      string // Path of file to delete (for delete action)
	KeepDirectory   string // Directory to keep (for batch_delete_by_dir)
	DeleteDirectory string // Directory to delete from (for batch_delete_by_dir)
}

// DeletionResult tracks deletion outcome
type DeletionResult struct {
	Path      string
	Success   bool
	Error     error
	SizeFreed int64
}

// SessionSummary provides final report
type SessionSummary struct {
	TotalSets     int
	SetsProcessed int
	FilesDeleted  int
	FilesFailed   int
	SpaceFreed    int64
	Results       []DeletionResult
}
