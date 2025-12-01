package interactive

import (
	"fmt"
	"os"

	"dup-finder/internal/models"
)

// SafeDelete performs pre-flight checks and deletes the file
func SafeDelete(path string) models.DeletionResult {
	result := models.DeletionResult{Path: path}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		result.Error = fmt.Errorf("cannot access file: %w", err)
		return result
	}

	// Verify it's a regular file
	if !info.Mode().IsRegular() {
		result.Error = fmt.Errorf("not a regular file")
		return result
	}

	size := info.Size()

	// Attempt deletion
	if err := os.Remove(path); err != nil {
		result.Error = fmt.Errorf("deletion failed: %w", err)
		return result
	}

	result.Success = true
	result.SizeFreed = size
	return result
}
