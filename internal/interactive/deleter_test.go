package interactive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeDelete(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	t.Run("successful deletion", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tmpDir, "test_file.txt")
		content := []byte("test content")
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Delete the file
		result := SafeDelete(testFile)

		// Verify success
		if !result.Success {
			t.Errorf("Expected success, got failure: %v", result.Error)
		}

		if result.SizeFreed != int64(len(content)) {
			t.Errorf("Expected size freed %d, got %d", len(content), result.SizeFreed)
		}

		// Verify file is deleted
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Errorf("File still exists after deletion")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		nonExistentFile := filepath.Join(tmpDir, "nonexistent.txt")

		result := SafeDelete(nonExistentFile)

		if result.Success {
			t.Errorf("Expected failure for non-existent file")
		}

		if result.Error == nil {
			t.Errorf("Expected error for non-existent file")
		}
	})

	t.Run("directory instead of file", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "test_dir")
		if err := os.Mkdir(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		result := SafeDelete(testDir)

		if result.Success {
			t.Errorf("Expected failure for directory")
		}

		if result.Error == nil {
			t.Errorf("Expected error for directory")
		}
	})

	t.Run("read-only file", func(t *testing.T) {
		// Create a read-only file in a writable directory
		testFile := filepath.Join(tmpDir, "readonly.txt")
		if err := os.WriteFile(testFile, []byte("readonly"), 0444); err != nil {
			t.Fatalf("Failed to create read-only file: %v", err)
		}

		result := SafeDelete(testFile)

		// Should succeed because parent directory is writable
		if !result.Success {
			t.Errorf("Expected success for read-only file in writable directory, got: %v", result.Error)
		}
	})
}

func TestSafeDeleteResult(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "size_test.txt")
	content := []byte("1234567890")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result := SafeDelete(testFile)

	if result.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, result.Path)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %v", result.Error)
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}

	if result.SizeFreed != int64(len(content)) {
		t.Errorf("Expected size freed %d, got %d", len(content), result.SizeFreed)
	}
}
