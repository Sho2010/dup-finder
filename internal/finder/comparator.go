package finder

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/cespare/xxhash/v2"

	"dup-finder/internal/models"
)

// CalculateFileHash computes the xxHash hash of a file
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := xxhash.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// ComputeHashesParallel computes hashes for multiple files in parallel
func ComputeHashesParallel(files []*models.FileInfo, numWorkers int) error {
	if len(files) == 0 {
		return nil
	}

	jobs := make(chan *models.FileInfo, len(files))
	errors := make(chan error, len(files))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobs {
				hash, err := CalculateFileHash(file.Path)
				if err != nil {
					errors <- fmt.Errorf("error hashing %s: %w", file.Path, err)
					continue
				}
				file.Hash = hash
			}
		}()
	}

	// Submit jobs
	for i := range files {
		jobs <- files[i]
	}
	close(jobs)

	// Wait for completion
	wg.Wait()
	close(errors)

	// Collect errors (if any)
	var firstError error
	for err := range errors {
		if firstError == nil {
			firstError = err
		}
		// Log other errors but don't fail completely
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	return firstError
}
