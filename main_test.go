package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateFileHash(t *testing.T) {
	// Create a temporary file with known content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	err := os.WriteFile(tmpFile, content, 0644)
	require.NoError(t, err)

	// Calculate hash
	hash, err := calculateFileHash(tmpFile)
	require.NoError(t, err)

	// Verify hash is not empty
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA256 produces 64 hex characters

	// Verify same content produces same hash
	hash2, err := calculateFileHash(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, hash, hash2)
}

func TestCalculateFileHash_DifferentContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two files with different content
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	err := os.WriteFile(file1, []byte("content1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("content2"), 0644)
	require.NoError(t, err)

	hash1, err := calculateFileHash(file1)
	require.NoError(t, err)

	hash2, err := calculateFileHash(file2)
	require.NoError(t, err)

	// Different content should produce different hashes
	assert.NotEqual(t, hash1, hash2)
}

func TestCalculateFileHash_SameContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two files with same content
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	content := []byte("identical content")
	err := os.WriteFile(file1, content, 0644)
	require.NoError(t, err)

	err = os.WriteFile(file2, content, 0644)
	require.NoError(t, err)

	hash1, err := calculateFileHash(file1)
	require.NoError(t, err)

	hash2, err := calculateFileHash(file2)
	require.NoError(t, err)

	// Same content should produce same hash
	assert.Equal(t, hash1, hash2)
}

func TestCalculateFileHash_NonExistentFile(t *testing.T) {
	hash, err := calculateFileHash("/non/existent/file.txt")
	assert.Error(t, err)
	assert.Empty(t, hash)
}

func TestDupFinderWithExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"file1.zip": "duplicate content",
		"file2.zip": "duplicate content",
		"file3.avi": "different content",
		"file4.txt": "duplicate content", // Same content but wrong extension
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Set extensions filter
	extensions = []string{".zip", ".avi"}

	// Map to store hash -> list of file paths
	hashMap := make(map[string][]string)

	// Walk through directory
	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Check file extension
		if len(extensions) > 0 {
			ext := filepath.Ext(path)
			matched := false
			for _, allowedExt := range extensions {
				if allowedExt == ext {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		hash, err := calculateFileHash(path)
		if err != nil {
			return err
		}

		hashMap[hash] = append(hashMap[hash], path)
		return nil
	})

	require.NoError(t, err)

	// Verify results
	// Should find 2 .zip files with duplicate content
	duplicateCount := 0
	for _, files := range hashMap {
		if len(files) > 1 {
			duplicateCount++
		}
	}

	assert.Equal(t, 1, duplicateCount, "Should find one set of duplicates")

	// Verify .txt file was not included
	for _, files := range hashMap {
		for _, file := range files {
			assert.NotContains(t, file, ".txt", "Should not include .txt files")
		}
	}
}

func TestDupFinderWithMinSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files with different sizes
	smallContent := []byte("small")
	largeContent := []byte("this is a much larger content for testing")

	files := map[string][]byte{
		"small1.zip": smallContent,
		"small2.zip": smallContent,
		"large1.zip": largeContent,
		"large2.zip": largeContent,
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, content, 0644)
		require.NoError(t, err)
	}

	// Set minimum size to filter out small files
	minSize = int64(10)
	extensions = []string{".zip"}

	hashMap := make(map[string][]string)

	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Skip files smaller than minimum size
		if info.Size() < minSize {
			return nil
		}

		// Check extension
		if len(extensions) > 0 {
			ext := filepath.Ext(path)
			matched := false
			for _, allowedExt := range extensions {
				if allowedExt == ext {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		hash, err := calculateFileHash(path)
		if err != nil {
			return err
		}

		hashMap[hash] = append(hashMap[hash], path)
		return nil
	})

	require.NoError(t, err)

	// Should only find large files
	totalFiles := 0
	for _, files := range hashMap {
		totalFiles += len(files)
	}

	assert.Equal(t, 2, totalFiles, "Should only include 2 large files")

	// Verify large files are duplicates
	duplicateCount := 0
	for _, files := range hashMap {
		if len(files) > 1 {
			duplicateCount++
		}
	}

	assert.Equal(t, 1, duplicateCount, "Should find one set of large duplicates")

	// Reset for other tests
	minSize = 0
}

func TestDupFinderWithMaxDepth(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	// tmpDir/
	//   file1.zip
	//   dir1/
	//     file2.zip
	//     dir2/
	//       file3.zip

	content := []byte("test content")

	// Root level
	err := os.WriteFile(filepath.Join(tmpDir, "file1.zip"), content, 0644)
	require.NoError(t, err)

	// First level
	dir1 := filepath.Join(tmpDir, "dir1")
	err = os.Mkdir(dir1, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir1, "file2.zip"), content, 0644)
	require.NoError(t, err)

	// Second level
	dir2 := filepath.Join(dir1, "dir2")
	err = os.Mkdir(dir2, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir2, "file3.zip"), content, 0644)
	require.NoError(t, err)

	// Test with maxDepth = 1 (should find file1.zip and file2.zip, but not file3.zip)
	maxDepth = 1
	baseDir, err := filepath.Abs(tmpDir)
	require.NoError(t, err)

	foundFiles := []string{}

	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Check max depth
			if maxDepth >= 0 {
				absPath, err := filepath.Abs(path)
				if err != nil {
					return nil
				}
				relPath, err := filepath.Rel(baseDir, absPath)
				if err != nil {
					return nil
				}
				depth := len(filepath.SplitList(relPath)) - 1
				if relPath != "." {
					depth = len(filepath.SplitList(filepath.ToSlash(relPath)))
				}
				if depth > maxDepth {
					return filepath.SkipDir
				}
			}
			return nil
		}

		foundFiles = append(foundFiles, filepath.Base(path))
		return nil
	})

	require.NoError(t, err)

	// Should find file1.zip and file2.zip but not file3.zip
	assert.Contains(t, foundFiles, "file1.zip")
	assert.Contains(t, foundFiles, "file2.zip")

	// Reset for other tests
	maxDepth = -1
}
