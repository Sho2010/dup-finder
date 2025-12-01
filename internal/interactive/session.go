package interactive

import (
	"fmt"
	"os"

	"dup-finder/internal/finder"
	"dup-finder/internal/models"
)

// RunInteractiveSession manages the entire interactive workflow
func RunInteractiveSession(comparisons []models.PairComparison, opts models.ScanOptions) (*models.SessionSummary, error) {
	// 1. Convert PairComparison to DuplicateSet (only for hash-matching pairs)
	sets := convertToDuplicateSets(comparisons, opts.NumWorkers)

	if len(sets) == 0 {
		fmt.Fprintln(os.Stderr, "No duplicate files found (based on content hash)")
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
			return nil, err
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
func convertToDuplicateSets(comparisons []models.PairComparison, numWorkers int) []models.DuplicateSet {
	var sets []models.DuplicateSet

	for _, comp := range comparisons {
		// Collect files from this comparison that need hashing
		var filesToHash []*models.FileInfo
		for i := range comp.Matches {
			if comp.Matches[i].File1.Hash == "" {
				filesToHash = append(filesToHash, &comp.Matches[i].File1)
			}
			if comp.Matches[i].File2.Hash == "" {
				filesToHash = append(filesToHash, &comp.Matches[i].File2)
			}
		}

		// Calculate hashes if needed
		if len(filesToHash) > 0 {
			ensureHashesCalculated(filesToHash, numWorkers)
		}

		// Create DuplicateSet for each match where hashes actually match
		for _, match := range comp.Matches {
			if match.File1.Hash != "" && match.File1.Hash == match.File2.Hash {
				sets = append(sets, models.DuplicateSet{
					Hash:  match.File1.Hash,
					Files: []models.FileInfo{match.File1, match.File2},
				})
			}
		}
	}

	return sets
}

// ensureHashesCalculated computes hashes for files that don't have them
func ensureHashesCalculated(files []*models.FileInfo, numWorkers int) {
	var needHash []*models.FileInfo
	for _, f := range files {
		if f.Hash == "" {
			needHash = append(needHash, f)
		}
	}

	if len(needHash) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "Computing hashes for %d files...\n", len(needHash))
	finder.ComputeHashesParallel(needHash, numWorkers)
	fmt.Fprintln(os.Stderr, "Hash calculation complete.")
	fmt.Fprintln(os.Stderr)
}
