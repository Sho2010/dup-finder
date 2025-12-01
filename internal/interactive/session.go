package interactive

import (
	"fmt"
	"os"

	"github.com/Sho2010/dup-finder/internal/finder"
	"github.com/Sho2010/dup-finder/internal/models"
)

// RunInteractiveSession manages the entire interactive workflow
func RunInteractiveSession(comparisons []models.PairComparison, opts models.ScanOptions) (*models.SessionSummary, error) {
	// 1. Convert PairComparison to DuplicateSet (only for hash-matching pairs)
	sets := convertToDuplicateSets(comparisons, opts.NumWorkers)

	if len(sets) == 0 {
		fmt.Fprintln(os.Stderr, "No duplicate files found (based on size)")
		return &models.SessionSummary{}, nil
	}

	// Check if batch-by-directory option should be available
	// (only when comparing exactly 2 directories)
	allowBatchByDir := len(opts.Directories) == 2

	// 2. Collect user decisions for all duplicate sets
	var actions []models.UserAction
	batchDirAction := "" // Track if user chose batch deletion by directory

	for i, set := range sets {
		set.ID = i + 1

		// If batch directory deletion was chosen, apply it automatically
		if batchDirAction != "" {
			var deleteDir string
			if batchDirAction == "keep_dir1" {
				deleteDir = set.Files[1].Directory
			} else {
				deleteDir = set.Files[0].Directory
			}

			// Find which file to delete based on directory
			for _, file := range set.Files {
				if file.Directory == deleteDir {
					actions = append(actions, models.UserAction{
						Action:     "delete",
						DeleteFile: file.Path,
					})
				}
			}
			continue
		}

		// Display the duplicate set
		if err := DisplayDuplicateSet(set); err != nil {
			return nil, err
		}

		// Get user choice
		action, err := PromptUserAction(set, allowBatchByDir)
		if err != nil {
			if err.Error() == "user finished" {
				// User wants to proceed with selected files
				break
			}
			return nil, err
		}

		// Handle hash computation request
		if action.Action == "compute_hash" {
			fmt.Fprintln(os.Stderr, "Computing hashes...")
			err := computeHashForSet(&set, opts.NumWorkers)
			if err != nil {
				if err.Error() == "hash mismatch" {
					fmt.Fprintln(os.Stderr, "✗ Files are different (hash mismatch). Skipping.")
					fmt.Fprintln(os.Stderr)
					continue
				}
				return nil, err
			}
			fmt.Fprintln(os.Stderr, "✓ Files are identical (hash verified)")
			fmt.Fprintln(os.Stderr)

			// Update the set in the slice
			sets[i] = set

			// Re-display with hash info and prompt again
			if err := DisplayDuplicateSet(set); err != nil {
				return nil, err
			}

			action, err = PromptUserAction(set, allowBatchByDir)
			if err != nil {
				if err.Error() == "user finished" {
					// User wants to proceed with selected files
					break
				}
				return nil, err
			}
		}

		// Handle batch directory deletion
		if action.Action == "batch_delete_by_dir" {
			// Set batch mode for remaining sets
			if action.KeepDirectory == set.Files[0].Directory {
				batchDirAction = "keep_dir1"
			} else {
				batchDirAction = "keep_dir2"
			}

			// Apply to current set
			for _, file := range set.Files {
				if file.Directory == action.DeleteDirectory {
					actions = append(actions, models.UserAction{
						Action:     "delete",
						DeleteFile: file.Path,
					})
				}
			}

			fmt.Fprintf(os.Stderr, "\nBatch mode enabled: All remaining duplicates from %s will be deleted.\n", action.DeleteDirectory)
			fmt.Fprintln(os.Stderr)
			continue
		}

		// Collect individual actions (don't delete yet)
		if action.Action == "delete" {
			actions = append(actions, action)
		}
	}

	// 3. Show final confirmation with list of files to delete
	if len(actions) == 0 {
		fmt.Fprintln(os.Stderr, "\nNo files selected for deletion.")
		return &models.SessionSummary{TotalSets: len(sets)}, nil
	}

	confirmed, err := ConfirmDeletion(actions)
	if err != nil || !confirmed {
		fmt.Fprintln(os.Stderr, "\nDeletion cancelled.")
		return &models.SessionSummary{TotalSets: len(sets)}, nil
	}

	// 4. Execute deletions and collect results
	summary := &models.SessionSummary{
		TotalSets:     len(sets),
		SetsProcessed: len(actions),
	}

	for _, action := range actions {
		result := SafeDelete(action.DeleteFile)
		summary.Results = append(summary.Results, result)

		if result.Success {
			summary.FilesDeleted++
			summary.SpaceFreed += result.SizeFreed
		} else {
			summary.FilesFailed++
		}
	}

	return summary, nil
}

// convertToDuplicateSets converts PairComparison to DuplicateSet (keeps pairwise structure)
// No hash calculation is performed - hashes are computed on-demand
func convertToDuplicateSets(comparisons []models.PairComparison, numWorkers int) []models.DuplicateSet {
	var sets []models.DuplicateSet

	for _, comp := range comparisons {
		// Create DuplicateSet for each match based on size (no hash required)
		for _, match := range comp.Matches {
			sets = append(sets, models.DuplicateSet{
				Hash:         "",    // Empty - not computed yet
				HashComputed: false, // Hash will be computed on-demand
				Files:        []models.FileInfo{match.File1, match.File2},
			})
		}
	}

	return sets
}

// computeHashForSet calculates hashes for files in a specific duplicate set
func computeHashForSet(set *models.DuplicateSet, numWorkers int) error {
	// Collect files that need hashing
	var filesToHash []*models.FileInfo
	for i := range set.Files {
		if set.Files[i].Hash == "" {
			filesToHash = append(filesToHash, &set.Files[i])
		}
	}

	if len(filesToHash) == 0 {
		set.HashComputed = true
		return nil
	}

	// Compute hashes using existing parallel function
	finder.ComputeHashesParallel(filesToHash, numWorkers)

	// Verify all hashes match
	if len(set.Files) > 0 {
		firstHash := set.Files[0].Hash
		for i := 1; i < len(set.Files); i++ {
			if set.Files[i].Hash != firstHash {
				return fmt.Errorf("hash mismatch")
			}
		}
		set.Hash = firstHash
	}

	set.HashComputed = true
	return nil
}
