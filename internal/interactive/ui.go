package interactive

import (
	"fmt"
	"os"

	"dup-finder/internal/models"
)

// DisplayDuplicateSet shows file details for user decision
func DisplayDuplicateSet(set models.DuplicateSet) error {
	fmt.Printf("\n=== Duplicate Set #%d ===\n", set.ID)
	fmt.Printf("Found %d files with same size\n", len(set.Files))

	// Only show hash if computed
	if set.HashComputed {
		fmt.Printf("Hash: %s... (verified)\n", set.Hash[:16])
	}
	fmt.Println()

	for i, file := range set.Files {
		fmt.Printf("[%d] %s\n", i+1, file.Path)
		fmt.Printf("    Size: %s\n", formatSize(file.Size))
		fmt.Printf("    Modified: %s\n", file.ModTime.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

// PromptUserAction gets user's choice for a duplicate set
func PromptUserAction(set models.DuplicateSet, allowBatchByDir bool) (models.UserAction, error) {
	for {
		fmt.Println("Choose an action:")
		fmt.Println("  [s] Skip (do nothing)")
		fmt.Println("  [1] Keep file 1, delete file 2")
		fmt.Println("  [2] Keep file 2, delete file 1")

		// Show hash option only if hash hasn't been computed yet
		if !set.HashComputed {
			fmt.Println("  [h] Compute hash to verify files are identical")
		}

		if allowBatchByDir {
			// Show directory names for batch operations
			dir1 := set.Files[0].Directory
			dir2 := set.Files[1].Directory
			fmt.Printf("  [a] Keep all from %s, delete all from %s\n", dir1, dir2)
			fmt.Printf("  [b] Keep all from %s, delete all from %s\n", dir2, dir1)
		}

		fmt.Println("  [q] Quit interactive mode")
		fmt.Println("  [f] Finish selection and proceed to confirmation")
		fmt.Print("\nYour choice: ")

		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			return models.UserAction{}, fmt.Errorf("failed to read input: %w", err)
		}

		switch input {
		case "s", "S":
			return models.UserAction{Action: "skip"}, nil
		case "q", "Q":
			return models.UserAction{}, fmt.Errorf("user quit")
		case "f", "F":
			return models.UserAction{}, fmt.Errorf("user finished")
		case "h", "H":
			if !set.HashComputed {
				return models.UserAction{Action: "compute_hash"}, nil
			}
			fmt.Println("Hash already computed. Please choose a different option.")
			fmt.Println()
		case "1":
			return models.UserAction{
				Action:     "delete",
				KeepFile:   set.Files[0].Path,
				DeleteFile: set.Files[1].Path,
			}, nil
		case "2":
			return models.UserAction{
				Action:     "delete",
				KeepFile:   set.Files[1].Path,
				DeleteFile: set.Files[0].Path,
			}, nil
		case "a", "A":
			if allowBatchByDir {
				return models.UserAction{
					Action:          "batch_delete_by_dir",
					KeepDirectory:   set.Files[0].Directory,
					DeleteDirectory: set.Files[1].Directory,
				}, nil
			}
			fmt.Println("Invalid choice. Please try again.")
			fmt.Println()
		case "b", "B":
			if allowBatchByDir {
				return models.UserAction{
					Action:          "batch_delete_by_dir",
					KeepDirectory:   set.Files[1].Directory,
					DeleteDirectory: set.Files[0].Directory,
				}, nil
			}
			fmt.Println("Invalid choice. Please try again.")
			fmt.Println()
		default:
			fmt.Println("Invalid choice. Please try again.")
			fmt.Println()
		}
	}
}

// ConfirmDeletion shows list of files to delete and asks for final confirmation
func ConfirmDeletion(actions []models.UserAction) (bool, error) {
	fmt.Println("\n=== Final Confirmation ===")
	fmt.Printf("The following %d file(s) will be deleted:\n\n", len(actions))

	var totalSize int64
	for i, action := range actions {
		info, err := os.Stat(action.DeleteFile)
		if err != nil {
			fmt.Printf("%d. %s (cannot read file info)\n", i+1, action.DeleteFile)
			continue
		}
		totalSize += info.Size()
		fmt.Printf("%d. %s (%s)\n", i+1, action.DeleteFile, formatSize(info.Size()))
	}

	fmt.Printf("\nTotal space to be freed: %s\n", formatSize(totalSize))
	fmt.Println("\nOptions:")
	fmt.Println("  [y] Execute deletions (proceed)")
	fmt.Println("  [n] Cancel all deletions (abort)")
	fmt.Print("\nYour choice [y/N]: ")

	var input string
	fmt.Scanln(&input)

	return input == "y" || input == "Y", nil
}

// DisplaySummary shows final results after session
func DisplaySummary(summary models.SessionSummary) error {
	fmt.Println("\n=== Interactive Session Summary ===")
	fmt.Printf("Duplicate Sets Found: %d\n", summary.TotalSets)
	fmt.Printf("Files Deleted: %d\n", summary.FilesDeleted)
	if summary.FilesFailed > 0 {
		fmt.Printf("Failed Deletions: %d\n", summary.FilesFailed)
	}
	fmt.Printf("Space Freed: %s\n", formatSize(summary.SpaceFreed))

	// Show successful deletions
	if summary.FilesDeleted > 0 {
		fmt.Println("\nSuccessfully Deleted:")
		for _, result := range summary.Results {
			if result.Success {
				fmt.Printf("  ✓ %s (%s freed)\n", result.Path, formatSize(result.SizeFreed))
			}
		}
	}

	// Show errors at the end
	if summary.FilesFailed > 0 {
		fmt.Println("\nFailed Deletions:")
		for _, result := range summary.Results {
			if !result.Success {
				fmt.Printf("  ✗ %s\n     Error: %v\n", result.Path, result.Error)
			}
		}
	}

	return nil
}

// formatSize converts bytes to human-readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
