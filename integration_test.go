package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dup-finder/internal/finder"
	"dup-finder/internal/models"
	"dup-finder/internal/scanner"
)

func TestTwoDirectoryComparison(t *testing.T) {
	// Setup test directories
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	require.NoError(t, os.Mkdir(dir1, 0755))
	require.NoError(t, os.Mkdir(dir2, 0755))

	// Create files
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "common.txt"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "common.txt"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "unique1.txt"), []byte("unique"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "unique2.txt"), []byte("unique"), 0644))

	opts := models.ScanOptions{
		Directories: []string{dir1, dir2},
		Recursive:   true,
		NumWorkers:  runtime.NumCPU(),
	}

	// Scan directories
	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	// Compare
	f := finder.NewFinder(opts)
	comparison := f.ComparePair(allFiles[dir1], allFiles[dir2])

	// Verify results
	assert.Len(t, comparison.Matches, 1)
	assert.Equal(t, "common.txt", comparison.Matches[0].Filename)
}

func TestThreeDirectoryComparison(t *testing.T) {
	// Setup test directories
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	dir3 := filepath.Join(tmpDir, "dir3")

	require.NoError(t, os.Mkdir(dir1, 0755))
	require.NoError(t, os.Mkdir(dir2, 0755))
	require.NoError(t, os.Mkdir(dir3, 0755))

	// Create files with same name, some with same content
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "file.txt"), []byte("same"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "file.txt"), []byte("same"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir3, "file.txt"), []byte("different"), 0644))

	opts := models.ScanOptions{
		Directories: []string{dir1, dir2, dir3},
		Recursive:   true,
		NumWorkers:  runtime.NumCPU(),
	}

	// Scan directories
	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	// Generate pairs
	pairs := finder.GeneratePairs([]string{dir1, dir2, dir3})
	assert.Len(t, pairs, 3) // C(3,2) = 3 pairs

	// Compare all pairs
	f := finder.NewFinder(opts)
	for _, pair := range pairs {
		comparison := f.ComparePair(allFiles[pair[0]], allFiles[pair[1]])
		assert.Len(t, comparison.Matches, 1)
		assert.Equal(t, "file.txt", comparison.Matches[0].Filename)
	}
}

func TestPairwiseWithHashComparison(t *testing.T) {
	// Setup test directories
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	require.NoError(t, os.Mkdir(dir1, 0755))
	require.NoError(t, os.Mkdir(dir2, 0755))

	// Create files with same name but different content
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "file.txt"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "file.txt"), []byte("content2"), 0644))

	opts := models.ScanOptions{
		Directories: []string{dir1, dir2},
		Recursive:   true,
		CompareHash: true,
		NumWorkers:  runtime.NumCPU(),
	}

	// Scan directories
	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	// Compare with hash
	f := finder.NewFinder(opts)
	comparison := f.ComparePair(allFiles[dir1], allFiles[dir2])

	// Verify results
	require.Len(t, comparison.Matches, 1)
	assert.True(t, comparison.Matches[0].HashChecked)
	assert.False(t, comparison.Matches[0].HashMatch) // Content is different
}

func TestNoCommonFiles(t *testing.T) {
	// Setup test directories
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	require.NoError(t, os.Mkdir(dir1, 0755))
	require.NoError(t, os.Mkdir(dir2, 0755))

	// Create unique files
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "file2.txt"), []byte("content2"), 0644))

	opts := models.ScanOptions{
		Directories: []string{dir1, dir2},
		Recursive:   true,
		NumWorkers:  runtime.NumCPU(),
	}

	// Scan directories
	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	// Compare
	f := finder.NewFinder(opts)
	comparison := f.ComparePair(allFiles[dir1], allFiles[dir2])

	// Verify no matches
	assert.Len(t, comparison.Matches, 0)
}

func TestSameNameDifferentContent(t *testing.T) {
	// Setup test directories
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	require.NoError(t, os.Mkdir(dir1, 0755))
	require.NoError(t, os.Mkdir(dir2, 0755))

	// Create files with same name, different content
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "photo.jpg"), []byte("photo data 1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "photo.jpg"), []byte("photo data 2"), 0644))

	opts := models.ScanOptions{
		Directories: []string{dir1, dir2},
		Recursive:   true,
		CompareHash: true,
		NumWorkers:  runtime.NumCPU(),
	}

	// Scan and compare
	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	f := finder.NewFinder(opts)
	comparison := f.ComparePair(allFiles[dir1], allFiles[dir2])

	// Verify: same name but different hash
	require.Len(t, comparison.Matches, 1)
	assert.Equal(t, "photo.jpg", comparison.Matches[0].Filename)
	assert.True(t, comparison.Matches[0].HashChecked)
	assert.False(t, comparison.Matches[0].HashMatch)
}

func TestExtensionFilter(t *testing.T) {
	// Setup test directories
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	require.NoError(t, os.Mkdir(dir1, 0755))
	require.NoError(t, os.Mkdir(dir2, 0755))

	// Create files with different extensions
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "file.txt"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "file.txt"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "file.jpg"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "file.jpg"), []byte("content"), 0644))

	opts := models.ScanOptions{
		Directories: []string{dir1, dir2},
		Recursive:   true,
		Extensions:  []string{".txt"}, // Only .txt files
		NumWorkers:  runtime.NumCPU(),
	}

	// Scan directories
	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	// Verify only .txt files are scanned
	assert.Equal(t, 1, len(allFiles[dir1]))
	assert.Equal(t, 1, len(allFiles[dir2]))

	// Compare
	f := finder.NewFinder(opts)
	comparison := f.ComparePair(allFiles[dir1], allFiles[dir2])

	// Should only find .txt file
	require.Len(t, comparison.Matches, 1)
	assert.Equal(t, "file.txt", comparison.Matches[0].Filename)
}

func TestGeneratePairs(t *testing.T) {
	tests := []struct {
		name     string
		dirs     []string
		expected int
	}{
		{"2 dirs", []string{"a", "b"}, 1},
		{"3 dirs", []string{"a", "b", "c"}, 3},
		{"4 dirs", []string{"a", "b", "c", "d"}, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pairs := finder.GeneratePairs(tt.dirs)
			assert.Len(t, pairs, tt.expected)

			// Verify pairs are unique
			seen := make(map[[2]string]bool)
			for _, pair := range pairs {
				assert.False(t, seen[pair], "Pairs should be unique")
				seen[pair] = true
			}
		})
	}
}

// TestCrossPlatformPaths verifies that the tool correctly handles paths
// on all platforms (Unix forward slashes, Windows backslashes)
func TestCrossPlatformPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	dir1 := filepath.Join(tmpDir, "dir1")
	subDir1 := filepath.Join(dir1, "subdir")
	dir2 := filepath.Join(tmpDir, "dir2")
	subDir2 := filepath.Join(dir2, "subdir")

	require.NoError(t, os.MkdirAll(subDir1, 0755))
	require.NoError(t, os.MkdirAll(subDir2, 0755))

	// Create files in subdirectories
	file1 := filepath.Join(subDir1, "test.txt")
	file2 := filepath.Join(subDir2, "test.txt")
	require.NoError(t, os.WriteFile(file1, []byte("content"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("content"), 0644))

	opts := models.ScanOptions{
		Directories: []string{dir1, dir2},
		Recursive:   true,
		NumWorkers:  runtime.NumCPU(),
	}

	// Scan directories
	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	// Verify files were found
	assert.Equal(t, 1, len(allFiles[dir1]))
	assert.Equal(t, 1, len(allFiles[dir2]))

	// Verify path separators are correct for platform
	for _, files := range allFiles {
		for _, file := range files {
			// filepath.Join ensures platform-correct separators
			assert.Contains(t, file.Path, string(filepath.Separator))
		}
	}

	// Compare
	f := finder.NewFinder(opts)
	comparison := f.ComparePair(allFiles[dir1], allFiles[dir2])

	// Should find the test.txt in both subdirectories
	require.Len(t, comparison.Matches, 1)
	assert.Equal(t, "test.txt", comparison.Matches[0].Filename)
}

// TestWindowsStylePaths verifies handling of Windows-style paths
// This test will use the platform's native path separator
func TestWindowsStylePaths(t *testing.T) {
	// This test verifies that paths with backslashes (Windows)
	// or forward slashes (Unix) are handled correctly
	tmpDir := t.TempDir()

	dir1 := filepath.Join(tmpDir, "path", "to", "dir1")
	dir2 := filepath.Join(tmpDir, "path", "to", "dir2")

	require.NoError(t, os.MkdirAll(dir1, 0755))
	require.NoError(t, os.MkdirAll(dir2, 0755))

	// Create files with spaces in names (common on Windows)
	file1 := filepath.Join(dir1, "file with spaces.txt")
	file2 := filepath.Join(dir2, "file with spaces.txt")
	require.NoError(t, os.WriteFile(file1, []byte("content"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("content"), 0644))

	opts := models.ScanOptions{
		Directories: []string{dir1, dir2},
		Recursive:   true,
		CompareHash: true,
		NumWorkers:  runtime.NumCPU(),
	}

	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	// Compare
	f := finder.NewFinder(opts)
	comparison := f.ComparePair(allFiles[dir1], allFiles[dir2])

	// Should find file with spaces in name
	require.Len(t, comparison.Matches, 1)
	assert.Equal(t, "file with spaces.txt", comparison.Matches[0].Filename)
	assert.True(t, comparison.Matches[0].HashChecked)
	assert.True(t, comparison.Matches[0].HashMatch)
}

// TestUnicodeFilenames verifies handling of Unicode characters in filenames
// Important for international users and cross-platform compatibility
func TestUnicodeFilenames(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	require.NoError(t, os.Mkdir(dir1, 0755))
	require.NoError(t, os.Mkdir(dir2, 0755))

	// Create files with Unicode characters (Japanese, emoji, etc.)
	unicodeNames := []string{
		"テスト.txt",           // Japanese
		"文件.txt",             // Chinese
		"файл.txt",            // Russian
		"test 😀.txt",         // Emoji
		"café.txt",            // Accented characters
	}

	for _, name := range unicodeNames {
		file1 := filepath.Join(dir1, name)
		file2 := filepath.Join(dir2, name)
		require.NoError(t, os.WriteFile(file1, []byte("content"), 0644))
		require.NoError(t, os.WriteFile(file2, []byte("content"), 0644))
	}

	opts := models.ScanOptions{
		Directories: []string{dir1, dir2},
		Recursive:   true,
		NumWorkers:  runtime.NumCPU(),
	}

	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	// Should find all files
	assert.Equal(t, len(unicodeNames), len(allFiles[dir1]))
	assert.Equal(t, len(unicodeNames), len(allFiles[dir2]))

	// Compare
	f := finder.NewFinder(opts)
	comparison := f.ComparePair(allFiles[dir1], allFiles[dir2])

	// Should find all Unicode named files
	assert.Len(t, comparison.Matches, len(unicodeNames))
}

// TestCaseInsensitiveExtensions verifies case-insensitive extension filtering
// Important for Windows where filesystem is case-insensitive
func TestCaseInsensitiveExtensions(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")

	require.NoError(t, os.Mkdir(dir1, 0755))

	// Create files with mixed case extensions
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "file.TXT"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "file.txt"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "file.Txt"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "file.jpg"), []byte("content"), 0644))

	opts := models.ScanOptions{
		Directories: []string{dir1},
		Recursive:   true,
		Extensions:  []string{".txt"}, // Lowercase filter
		NumWorkers:  runtime.NumCPU(),
	}

	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	// Should find all .txt files regardless of case
	assert.Equal(t, 3, len(allFiles[dir1]), "Should find TXT, txt, and Txt files")
}

// TestSkipNonExistentDirectories verifies that non-existent directories are skipped
// and comparison continues with remaining valid directories
func TestSkipNonExistentDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2_nonexistent")
	dir3 := filepath.Join(tmpDir, "dir3")

	// Create only dir1 and dir3
	require.NoError(t, os.Mkdir(dir1, 0755))
	require.NoError(t, os.Mkdir(dir3, 0755))

	// Create test files
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "test.txt"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir3, "test.txt"), []byte("content"), 0644))

	// Scan with all three directories, including non-existent dir2
	opts := models.ScanOptions{
		Directories: []string{dir1, dir2, dir3},
		Recursive:   true,
		NumWorkers:  runtime.NumCPU(),
	}

	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	require.NoError(t, err)

	// Should have files from dir1 and dir3, but not dir2
	assert.Equal(t, 1, len(allFiles[dir1]))
	assert.Equal(t, 0, len(allFiles[dir2])) // Non-existent directory
	assert.Equal(t, 1, len(allFiles[dir3]))

	// Generate pairs - should only include valid directories
	validDirs := []string{dir1, dir3}
	pairs := finder.GeneratePairs(validDirs)
	assert.Len(t, pairs, 1) // Only 1 pair: dir1-dir3

	// Compare the valid pair
	f := finder.NewFinder(opts)
	comparison := f.ComparePair(allFiles[dir1], allFiles[dir3])

	// Should find the matching file
	require.Len(t, comparison.Matches, 1)
	assert.Equal(t, "test.txt", comparison.Matches[0].Filename)
}
