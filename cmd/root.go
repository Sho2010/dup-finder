package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"dup-finder/internal/finder"
	"dup-finder/internal/interactive"
	"dup-finder/internal/models"
	"dup-finder/internal/output"
	"dup-finder/internal/scanner"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dup-finder [directory1] [directory2] [directory...]",
		Short: "Find duplicate files across multiple directories",
		Long:  `dup-finder scans multiple directories and finds duplicate files based on filename (optionally comparing content hash).`,
		Args:  cobra.MinimumNArgs(2),
		RunE:  runDupFinder,
	}

	recursive        bool
	minSize          int64
	extensions       []string
	maxDepth         int
	compareHash      bool
	numWorkers       int
	interactiveMode  bool
)

func init() {
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", true, "Search directories recursively")
	rootCmd.Flags().Int64VarP(&minSize, "min-size", "m", 0, "Minimum file size in bytes to consider")
	rootCmd.Flags().StringSliceVarP(&extensions, "extensions", "e", []string{}, "File extensions to consider (e.g., .zip,.avi,.mp4)")
	rootCmd.Flags().IntVarP(&maxDepth, "max-depth", "L", -1, "Maximum directory depth for recursive search (-1 for unlimited)")
	rootCmd.Flags().BoolVarP(&compareHash, "compare-hash", "H", false, "Compare file content using SHA256 hash")
	rootCmd.Flags().IntVarP(&numWorkers, "workers", "w", runtime.NumCPU(), "Number of parallel workers")
	rootCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "Enable interactive deletion mode")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func runDupFinder(cmd *cobra.Command, args []string) error {
	// Validate directories exist and filter out non-existent ones
	var validDirs []string
	for _, dir := range args {
		if _, err := os.Stat(dir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping %s: %v\n", dir, err)
			continue
		}
		validDirs = append(validDirs, dir)
	}

	// Check if we have at least 2 valid directories
	if len(validDirs) < 2 {
		return fmt.Errorf("need at least 2 valid directories to compare, found only %d", len(validDirs))
	}

	// Show which directories will be compared
	if len(validDirs) < len(args) {
		fmt.Fprintf(os.Stderr, "Comparing %d out of %d directories:\n", len(validDirs), len(args))
		for _, dir := range validDirs {
			fmt.Fprintf(os.Stderr, "  ✓ %s\n", dir)
		}
		fmt.Fprintln(os.Stderr)
	}

	// Build scan options
	opts := models.ScanOptions{
		Directories: validDirs,
		Recursive:   recursive,
		MinSize:     minSize,
		Extensions:  extensions,
		MaxDepth:    maxDepth,
		CompareHash: compareHash,
		NumWorkers:  numWorkers,
	}

	// Scan all directories
	s := scanner.NewScanner(opts)
	allFiles, err := s.ScanAll()
	if err != nil {
		return fmt.Errorf("error scanning directories: %w", err)
	}

	// Generate directory pairs (only for valid directories)
	pairs := finder.GeneratePairs(validDirs)

	// Compare each pair
	f := finder.NewFinder(opts)
	var comparisons []models.PairComparison

	for _, pair := range pairs {
		dir1Files := allFiles[pair[0]]
		dir2Files := allFiles[pair[1]]

		comparison := f.ComparePair(dir1Files, dir2Files)
		comparisons = append(comparisons, comparison)
	}

	// Format and print output to stdout
	result := output.FormatAllComparisons(comparisons, compareHash)
	fmt.Print(result)

	// Enter interactive mode if requested
	if interactiveMode {
		fmt.Fprintln(os.Stderr, "\n--- Entering Interactive Deletion Mode ---")
		summary, err := interactive.RunInteractiveSession(comparisons, opts)
		if err != nil {
			return fmt.Errorf("interactive session error: %w", err)
		}
		interactive.DisplaySummary(*summary)
	}

	return nil
}
