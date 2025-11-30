package finder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dup-finder/internal/models"
)

func TestCalculateFileHash(t *testing.T) {
	// Create a temporary file with known content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	err := os.WriteFile(tmpFile, content, 0644)
	require.NoError(t, err)

	// Calculate hash
	hash, err := CalculateFileHash(tmpFile)
	require.NoError(t, err)

	// Verify hash is not empty
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA256 produces 64 hex characters

	// Verify same content produces same hash
	hash2, err := CalculateFileHash(tmpFile)
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

	hash1, err := CalculateFileHash(file1)
	require.NoError(t, err)

	hash2, err := CalculateFileHash(file2)
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

	hash1, err := CalculateFileHash(file1)
	require.NoError(t, err)

	hash2, err := CalculateFileHash(file2)
	require.NoError(t, err)

	// Same content should produce same hash
	assert.Equal(t, hash1, hash2)
}

func TestCalculateFileHash_NonExistentFile(t *testing.T) {
	hash, err := CalculateFileHash("/non/existent/file.txt")
	assert.Error(t, err)
	assert.Empty(t, hash)
}

func TestComputeHashesParallel(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []struct {
		name    string
		content string
	}{
		{"file1.txt", "content1"},
		{"file2.txt", "content2"},
		{"file3.txt", "content3"},
	}

	var fileInfos []*models.FileInfo
	for _, f := range files {
		path := filepath.Join(tmpDir, f.name)
		err := os.WriteFile(path, []byte(f.content), 0644)
		require.NoError(t, err)

		fileInfos = append(fileInfos, &models.FileInfo{
			Path:      path,
			Directory: tmpDir,
			Size:      int64(len(f.content)),
		})
	}

	// Compute hashes in parallel
	err := ComputeHashesParallel(fileInfos, 2)
	require.NoError(t, err)

	// Verify all files have hashes
	for _, fi := range fileInfos {
		assert.NotEmpty(t, fi.Hash)
		assert.Len(t, fi.Hash, 64)
	}

	// Verify hashes are unique (since content is different)
	hashes := make(map[string]bool)
	for _, fi := range fileInfos {
		assert.False(t, hashes[fi.Hash], "Hash should be unique")
		hashes[fi.Hash] = true
	}
}

func TestComputeHashesParallel_EmptyList(t *testing.T) {
	var fileInfos []*models.FileInfo
	err := ComputeHashesParallel(fileInfos, 2)
	require.NoError(t, err)
}
