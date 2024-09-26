package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Cleaner is responsible for finding and deleting reproducible folders.
type Cleaner struct {
	rootPath string
	dryRun   bool
	force    bool
}

// NewCleaner creates a new Cleaner instance.
func NewCleaner(rootPath string, dryRun, force bool) *Cleaner {
	return &Cleaner{
		rootPath: rootPath,
		dryRun:   dryRun,
		force:    force,
	}
}

// Clean initiates the cleaning process.
func (c *Cleaner) Clean() {
	var deletionTasks []FolderDeletionTask
	err := filepath.Walk(c.rootPath, func(path string, info os.FileInfo, err error) error {
		// if err != nil {
		// 	return err
		// }

		if info.IsDir() {
			var exists bool
			var err error
			switch info.Name() {
			case ".venv":
				exists, err = CheckFolderAndConfigExists(filepath.Dir(path), ".venv", "pyproject.toml")
			case "node_modules":
				exists, err = CheckFolderAndConfigExists(filepath.Dir(path), "node_modules", "package.json")
			case "target":
				exists, err = CheckFolderAndConfigExists(filepath.Dir(path), "target", "Cargo.toml")
			case ".pixi":
				exists, err = CheckFolderAndConfigExists(filepath.Dir(path), ".pixi", "pixi.toml", "pyproject.toml")
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking folder existence: %v\n", err)
				return filepath.SkipDir
			}

			if exists {
				size, err := CalculateDirSize(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error calculating directory size: %v\n", err)
					return filepath.SkipDir
				}

				fmt.Printf("Identified %s directory for cleanup at %s, will free up %s.\n", info.Name(), path, HumanizeBytes(size))

				deletionTasks = append(deletionTasks, FolderDeletionTask{
					Path: path,
					Size: size,
				})

				// Skip checking subdirectories since this directory is marked for deletion.
				return filepath.SkipDir
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the file tree: %v\n", err)
		// return
	}

	if len(deletionTasks) == 0 {
		fmt.Println("No directories found to clean up.")
		return
	}

	if !c.dryRun && !c.force {
		fmt.Printf("Found %d directories to clean up. Total size: %s\n", len(deletionTasks), HumanizeBytes(calculateTotalSize(deletionTasks)))
		if !AskForConfirmation() {
			fmt.Println("Aborting cleanup.")
			return
		}
	}

	totalCleanedSize := DeleteFoldersConcurrently(deletionTasks, c.dryRun)
	fmt.Printf("Total cleaned up size: %s\n", HumanizeBytes(totalCleanedSize))
}

// calculateTotalSize calculates the total size of all tasks.
func calculateTotalSize(tasks []FolderDeletionTask) int64 {
	var totalSize int64
	for _, task := range tasks {
		totalSize += task.Size
	}
	return totalSize
}