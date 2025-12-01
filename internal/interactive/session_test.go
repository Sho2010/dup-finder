package interactive

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sho2010/dup-finder/internal/models"
)

func TestConvertToDuplicateSets(t *testing.T) {
	t.Run("creates sets without hash computation", func(t *testing.T) {
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
						HashChecked: false,
						HashMatch:   false,
					},
				},
			},
		}

		sets := convertToDuplicateSets(comparisons, 2)

		// Should create set regardless of hash
		if len(sets) != 1 {
			t.Errorf("Expected 1 duplicate set, got %d", len(sets))
		}

		// Hash should not be computed
		if sets[0].Hash != "" {
			t.Errorf("Expected empty hash, got %s", sets[0].Hash)
		}

		// HashComputed should be false
		if sets[0].HashComputed {
			t.Errorf("Expected HashComputed to be false")
		}

		if len(sets[0].Files) != 2 {
			t.Errorf("Expected 2 files in set, got %d", len(sets[0].Files))
		}
	})

	t.Run("creates multiple sets", func(t *testing.T) {
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
					{
						Filename: "file2.txt",
						File1: models.FileInfo{
							Path:      "/tmp/dir1/file2.txt",
							Directory: "/tmp/dir1",
							Size:      200,
							ModTime:   time.Now(),
							Hash:      "",
						},
						File2: models.FileInfo{
							Path:      "/tmp/dir2/file2.txt",
							Directory: "/tmp/dir2",
							Size:      200,
							ModTime:   time.Now(),
							Hash:      "",
						},
					},
				},
			},
		}

		sets := convertToDuplicateSets(comparisons, 2)

		if len(sets) != 2 {
			t.Errorf("Expected 2 duplicate sets, got %d", len(sets))
		}
	})
}

func TestComputeHashForSet(t *testing.T) {
	t.Run("computes hash for matching files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test files with identical content
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		content := []byte("identical content")
		if err := os.WriteFile(file1, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(file2, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		set := models.DuplicateSet{
			Files: []models.FileInfo{
				{Path: file1, Hash: ""},
				{Path: file2, Hash: ""},
			},
			Hash:         "",
			HashComputed: false,
		}

		err := computeHashForSet(&set, 2)

		// Should succeed
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify hashes were calculated
		if set.Files[0].Hash == "" {
			t.Errorf("Hash was not calculated for file1")
		}
		if set.Files[1].Hash == "" {
			t.Errorf("Hash was not calculated for file2")
		}

		// Same content should have same hash
		if set.Files[0].Hash != set.Files[1].Hash {
			t.Errorf("Same content should have same hash: %s != %s", set.Files[0].Hash, set.Files[1].Hash)
		}

		// Set hash should be populated
		if set.Hash == "" {
			t.Errorf("Set hash was not populated")
		}

		// HashComputed should be true
		if !set.HashComputed {
			t.Errorf("HashComputed should be true")
		}
	})

	t.Run("detects hash mismatch", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test files with different content
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		if err := os.WriteFile(file1, []byte("content A"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(file2, []byte("content B"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		set := models.DuplicateSet{
			Files: []models.FileInfo{
				{Path: file1, Hash: ""},
				{Path: file2, Hash: ""},
			},
			Hash:         "",
			HashComputed: false,
		}

		err := computeHashForSet(&set, 2)

		// Should return hash mismatch error
		if err == nil {
			t.Errorf("Expected hash mismatch error")
		}

		if err != nil && err.Error() != "hash mismatch" {
			t.Errorf("Expected 'hash mismatch' error, got: %v", err)
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
						Hash:      "", // Empty - hash not computed yet
					},
					File2: models.FileInfo{
						Path:      file2,
						Directory: dir2,
						Size:      int64(len(content)),
						ModTime:   time.Now(),
						Hash:      "", // Empty - hash not computed yet
					},
				},
			},
		},
	}

	sets := convertToDuplicateSets(comparisons, 2)

	if len(sets) != 1 {
		t.Fatalf("Expected 1 duplicate set, got %d", len(sets))
	}

	// Hash should NOT be calculated by convertToDuplicateSets
	if sets[0].Hash != "" {
		t.Errorf("Hash should not be calculated upfront, got: %s", sets[0].Hash)
	}

	// HashComputed should be false
	if sets[0].HashComputed {
		t.Errorf("HashComputed should be false initially")
	}

	if len(sets[0].Files) != 2 {
		t.Errorf("Expected 2 files in set, got %d", len(sets[0].Files))
	}

	// Now test on-demand hash computation
	err := computeHashForSet(&sets[0], 2)
	if err != nil {
		t.Fatalf("Failed to compute hash: %v", err)
	}

	// After computing, hash should be set
	if sets[0].Hash == "" {
		t.Errorf("Hash should be calculated after computeHashForSet")
	}

	// HashComputed should be true
	if !sets[0].HashComputed {
		t.Errorf("HashComputed should be true after computeHashForSet")
	}

	// Verify both files have the same hash
	if sets[0].Files[0].Hash != sets[0].Files[1].Hash {
		t.Errorf("Files with identical content should have same hash")
	}
}
