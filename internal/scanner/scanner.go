package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dup-finder/internal/models"
)

// Scanner handles directory scanning with filtering
type Scanner struct {
	options models.ScanOptions
}

// NewScanner creates a new scanner with the given options
func NewScanner(opts models.ScanOptions) *Scanner {
	return &Scanner{options: opts}
}

// Scan scans a single directory and returns all matching files
func (s *Scanner) Scan(directory string) ([]models.FileInfo, error) {
	baseDir, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path: %w", err)
	}

	var files []models.FileInfo
	pool := NewWorkerPool(s.options.NumWorkers)
	pool.Start()

	// Collect results in a separate goroutine
	done := make(chan bool)
	go func() {
		for result := range pool.Results() {
			if result.Error != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", result.Error)
				continue
			}
			files = append(files, result.FileInfo)
		}
		done <- true
	}()

	// Walk directory and submit jobs
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing path %s: %v\n", path, err)
			return nil
		}

		// Skip directories
		if info.IsDir() {
			if !s.options.Recursive && path != directory {
				return filepath.SkipDir
			}

			// Check max depth
			if s.options.MaxDepth >= 0 {
				absPath, err := filepath.Abs(path)
				if err != nil {
					return nil
				}
				relPath, err := filepath.Rel(baseDir, absPath)
				if err != nil {
					return nil
				}
				depth := strings.Count(relPath, string(filepath.Separator))
				if depth > s.options.MaxDepth {
					return filepath.SkipDir
				}
			}

			return nil
		}

		// Apply filters
		if !s.shouldIncludeFile(path, info) {
			return nil
		}

		// Submit job to worker pool
		pool.Submit(ScanJob{
			Path:      path,
			Directory: directory,
			Info:      info,
		})

		return nil
	})

	pool.Close()
	<-done

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return files, nil
}

// shouldIncludeFile checks if a file should be included based on filters
func (s *Scanner) shouldIncludeFile(path string, info os.FileInfo) bool {
	// Check minimum size
	if info.Size() < s.options.MinSize {
		return false
	}

	// Check file extension
	if len(s.options.Extensions) > 0 {
		ext := strings.ToLower(filepath.Ext(path))
		matched := false
		for _, allowedExt := range s.options.Extensions {
			if strings.ToLower(allowedExt) == ext {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// ScanAll scans all directories in parallel
func (s *Scanner) ScanAll() (map[string][]models.FileInfo, error) {
	results := make(map[string][]models.FileInfo)
	errors := make(chan error, len(s.options.Directories))
	filesChan := make(chan struct {
		dir   string
		files []models.FileInfo
	}, len(s.options.Directories))

	// Scan each directory in parallel
	for _, dir := range s.options.Directories {
		go func(directory string) {
			files, err := s.Scan(directory)
			if err != nil {
				errors <- fmt.Errorf("error scanning %s: %w", directory, err)
				return
			}
			filesChan <- struct {
				dir   string
				files []models.FileInfo
			}{directory, files}
		}(dir)
	}

	// Collect results
	for i := 0; i < len(s.options.Directories); i++ {
		select {
		case result := <-filesChan:
			results[result.dir] = result.files
		case err := <-errors:
			return nil, err
		}
	}

	return results, nil
}
