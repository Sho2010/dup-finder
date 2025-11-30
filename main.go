package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dup-finder [directory]",
		Short: "Find duplicate files in a directory",
		Long:  `dup-finder scans a directory recursively and finds duplicate files based on their content hash (SHA256).`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  runDupFinder,
	}

	recursive   bool
	minSize     int64
	extensions  []string
	maxDepth    int
)

func init() {
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", true, "Search directories recursively")
	rootCmd.Flags().Int64VarP(&minSize, "min-size", "m", 0, "Minimum file size in bytes to consider")
	rootCmd.Flags().StringSliceVarP(&extensions, "extensions", "e", []string{".zip", ".avi", ".mp4"}, "File extensions to consider (e.g., .zip,.avi,.mp4)")
	rootCmd.Flags().IntVarP(&maxDepth, "max-depth", "L", -1, "Maximum directory depth for recursive search (-1 for unlimited)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runDupFinder(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	// Get absolute path of base directory for depth calculation
	baseDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %w", err)
	}

	// Map to store hash -> list of file paths
	hashMap := make(map[string][]string)

	// Walk through directory
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing path %s: %v\n", path, err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			if !recursive && path != dir {
				return filepath.SkipDir
			}

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
				depth := strings.Count(relPath, string(filepath.Separator))
				if depth > maxDepth {
					return filepath.SkipDir
				}
			}

			return nil
		}

		// Skip files smaller than minimum size
		if info.Size() < minSize {
			return nil
		}

		// Check file extension
		if len(extensions) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			matched := false
			for _, allowedExt := range extensions {
				if strings.ToLower(allowedExt) == ext {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		// Calculate file hash
		hash, err := calculateFileHash(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error hashing file %s: %v\n", path, err)
			return nil // Continue walking
		}

		hashMap[hash] = append(hashMap[hash], path)
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	// Print duplicates
	duplicatesFound := false
	for hash, files := range hashMap {
		if len(files) > 1 {
			if !duplicatesFound {
				fmt.Println("Duplicate files found:")
				duplicatesFound = true
			}
			fmt.Printf("\nHash: %s\n", hash[:16]+"...")
			for _, file := range files {
				fmt.Printf("  - %s\n", file)
			}
		}
	}

	if !duplicatesFound {
		fmt.Println("No duplicate files found.")
	}

	return nil
}

func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
