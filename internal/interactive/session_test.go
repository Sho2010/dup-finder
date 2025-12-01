package interactive

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"dup-finder/internal/models"
)

func TestConvertToDuplicateSets(t *testing.T) {
	t.Run("with matching hashes", func(t *testing.T) {
		comparisons := []models.PairComparison{
			{
				Dir1: "/tmp/dir1",
				Dir2: "/tmp/dir2",
				Matches: []models.FileMatch{
					{
						Filename: "file1.txt",
						File1: models.FileInfo{
							Path:      "/tmp/dir1/file1.txt",
							Directory: "/tmp/dir1",
							Size:      100,
							ModTime:   time.Now(),
							Hash:      "abc123",
						},
						File2: models.FileInfo{
							Path:      "/tmp/dir2/file1.txt",
							Directory: "/tmp/dir2",
							Size:      100,
							ModTime:   time.Now(),
							Hash:      "abc123",
						},
						HashChecked: true,
						HashMatch:   true,
					},
				},
			},
		}

		sets := convertToDuplicateSets(comparisons, 2)

		if len(sets) != 1 {
			t.Errorf("Expected 1 duplicate set, got %d", len(sets))
		}

		if sets[0].Hash != "abc123" {
			t.Errorf("Expected hash abc123, got %s", sets[0].Hash)
		}

		if len(sets[0].Files) != 2 {
			t.Errorf("Expected 2 files in set, got %d", len(sets[0].Files))
		}
	})

	t.Run("with non-matching hashes", func(t *testing.T) {
		comparisons := []models.PairComparison{
			{
				Dir1: "/tmp/dir1",
				Dir2: "/tmp/dir2",
				Matches: []models.FileMatch{
					{
						Filename: "file1.txt",
						File1: models.FileInfo{
							Path:      "/tmp/dir1/file1.txt",
							Directory: "/tmp/dir1",
							Size:      100,
							ModTime:   time.Now(),
							Hash:      "abc123",
						},
						File2: models.FileInfo{
							Path:      "/tmp/dir2/file1.txt",
							Directory: "/tmp/dir2",
							Size:      100,
							ModTime:   time.Now(),
							Hash:      "def456",
						},
						HashChecked: true,
						HashMatch:   false,
					},
				},
			},
		}

		sets := convertToDuplicateSets(comparisons, 2)

		if len(sets) != 0 {
			t.Errorf("Expected 0 duplicate sets, got %d", len(sets))
		}
	})

	t.Run("with empty hash", func(t *testing.T) {
		comparisons := []models.PairComparison{
			{
				Dir1: "/tmp/dir1",
				Dir2: "/tmp/dir2",
				Matches: []models.FileMatch{
					{
						Filename: "file1.txt",
						File1: models.FileInfo{
							Path:      "/tmp/dir1/file1.txt",
							Directory: "/tmp/dir1",
							Size:      100,
							ModTime:   time.Now(),
							Hash:      "",
						},
						File2: models.FileInfo{
							Path:      "/tmp/dir2/file1.txt",
							Directory: "/tmp/dir2",
							Size:      100,
							ModTime:   time.Now(),
							Hash:      "",
						},
					},
				},
			},
		}

		// This will try to calculate hashes, but will fail for non-existent files
		// The test verifies that it doesn't crash
		sets := convertToDuplicateSets(comparisons, 2)

		// Files don't exist, so hash calculation will fail and sets will be empty
		if len(sets) != 0 {
			t.Errorf("Expected 0 duplicate sets for non-existent files, got %d", len(sets))
		}
	})
}

func TestEnsureHashesCalculated(t *testing.T) {
	t.Run("no hashes needed", func(t *testing.T) {
		files := []*models.FileInfo{
			{
				Path: "/tmp/file1.txt",
				Hash: "abc123",
			},
			{
				Path: "/tmp/file2.txt",
				Hash: "def456",
			},
		}

		// Should not panic or error
		ensureHashesCalculated(files, 2)

		// Hashes should remain unchanged
		if files[0].Hash != "abc123" {
			t.Errorf("Hash was modified")
		}
	})

	t.Run("with actual files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test files
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(file2, []byte("content1"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		files := []*models.FileInfo{
			{Path: file1, Hash: ""},
			{Path: file2, Hash: ""},
		}

		ensureHashesCalculated(files, 2)

		// Verify hashes were calculated
		if files[0].Hash == "" {
			t.Errorf("Hash was not calculated for file1")
		}
		if files[1].Hash == "" {
			t.Errorf("Hash was not calculated for file2")
		}

		// Same content should have same hash
		if files[0].Hash != files[1].Hash {
			t.Errorf("Same content should have same hash: %s != %s", files[0].Hash, files[1].Hash)
		}
	})
}

func TestConvertToDuplicateSetsWithRealFiles(t *testing.T) {
	tmpDir := t.TempDir()

	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}

	// Create identical files
	content := []byte("test content for duplicate detection")
	file1 := filepath.Join(dir1, "test.txt")
	file2 := filepath.Join(dir2, "test.txt")

	if err := os.WriteFile(file1, content, 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, content, 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	comparisons := []models.PairComparison{
		{
			Dir1: dir1,
			Dir2: dir2,
			Matches: []models.FileMatch{
				{
					Filename: "test.txt",
					File1: models.FileInfo{
						Path:      file1,
						Directory: dir1,
						Size:      int64(len(content)),
						ModTime:   time.Now(),
						Hash:      "", // Empty - will be calculated
					},
					File2: models.FileInfo{
						Path:      file2,
						Directory: dir2,
						Size:      int64(len(content)),
						ModTime:   time.Now(),
						Hash:      "", // Empty - will be calculated
					},
				},
			},
		},
	}

	sets := convertToDuplicateSets(comparisons, 2)

	if len(sets) != 1 {
		t.Fatalf("Expected 1 duplicate set, got %d", len(sets))
	}

	if sets[0].Hash == "" {
		t.Errorf("Hash was not calculated")
	}

	if len(sets[0].Files) != 2 {
		t.Errorf("Expected 2 files in set, got %d", len(sets[0].Files))
	}

	// Verify both files have the same hash
	if sets[0].Files[0].Hash != sets[0].Files[1].Hash {
		t.Errorf("Files with identical content should have same hash")
	}
}
