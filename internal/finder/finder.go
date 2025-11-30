package finder

import (
	"path/filepath"
	"runtime"
	"sort"

	"dup-finder/internal/models"
)

// Finder handles duplicate file detection
type Finder struct {
	options models.ScanOptions
}

// NewFinder creates a new finder with the given options
func NewFinder(opts models.ScanOptions) *Finder {
	return &Finder{options: opts}
}

// ComparePair compares files from two directories and finds matches by name
func (f *Finder) ComparePair(dir1Files, dir2Files []models.FileInfo) models.PairComparison {
	// Group files by basename
	group1 := groupByName(dir1Files)
	group2 := groupByName(dir2Files)

	// Find common filenames
	matches := findCommonFiles(group1, group2)

	// If hash comparison is enabled, compute hashes
	if f.options.CompareHash && len(matches) > 0 {
		f.computeHashesForMatches(matches)
	}

	// Extract directory names from file lists
	var dir1, dir2 string
	if len(dir1Files) > 0 {
		dir1 = dir1Files[0].Directory
	}
	if len(dir2Files) > 0 {
		dir2 = dir2Files[0].Directory
	}

	// Sort matches by filename for consistent output
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Filename < matches[j].Filename
	})

	return models.PairComparison{
		Dir1:    dir1,
		Dir2:    dir2,
		Matches: matches,
	}
}

// groupByName creates a map of basename -> FileInfo
func groupByName(files []models.FileInfo) map[string]models.FileInfo {
	m := make(map[string]models.FileInfo)
	for _, f := range files {
		basename := filepath.Base(f.Path)
		m[basename] = f
	}
	return m
}

// findCommonFiles finds files that exist in both groups
func findCommonFiles(group1, group2 map[string]models.FileInfo) []models.FileMatch {
	var matches []models.FileMatch
	for name, file1 := range group1 {
		if file2, exists := group2[name]; exists {
			matches = append(matches, models.FileMatch{
				Filename:    name,
				File1:       file1,
				File2:       file2,
				HashChecked: false,
				HashMatch:   false,
			})
		}
	}
	return matches
}

// computeHashesForMatches computes hashes for all matched files and updates HashMatch
func (f *Finder) computeHashesForMatches(matches []models.FileMatch) {
	// Collect all files that need hashing
	var files []*models.FileInfo
	for i := range matches {
		files = append(files, &matches[i].File1)
		files = append(files, &matches[i].File2)
	}

	// Compute hashes in parallel
	numWorkers := runtime.NumCPU() * 2 // I/O bound, so use more workers
	if f.options.NumWorkers > 0 {
		numWorkers = f.options.NumWorkers * 2
	}

	_ = ComputeHashesParallel(files, numWorkers)

	// Update HashMatch for each pair
	for i := range matches {
		matches[i].HashChecked = true
		matches[i].HashMatch = matches[i].File1.Hash == matches[i].File2.Hash &&
			matches[i].File1.Hash != ""
	}
}

// GeneratePairs generates all unique pairs of directories
func GeneratePairs(dirs []string) [][2]string {
	var pairs [][2]string
	for i := 0; i < len(dirs); i++ {
		for j := i + 1; j < len(dirs); j++ {
			pairs = append(pairs, [2]string{dirs[i], dirs[j]})
		}
	}
	return pairs
}
