package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const NUM_WORKERS = 100
const CHANNEL_BUFFER = 1000

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
	progressChan := make(chan string, CHANNEL_BUFFER)
	deletionTasks, totalScanned := c.WalkConcurrently(progressChan)

	close(progressChan)

	if len(deletionTasks) == 0 {
		fmt.Println("\nNo directories found to clean up.")
		return
	}

	fmt.Printf("\nScanned a total of %d directories.\n", totalScanned)

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

// WalkConcurrently performs a concurrent filesystem walk.
func (c *Cleaner) WalkConcurrently(progressChan chan string) ([]FolderDeletionTask, int) {
	const numWorkers = NUM_WORKERS
	paths := make(chan string, CHANNEL_BUFFER)
	results := make(chan FolderDeletionTask, CHANNEL_BUFFER)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go c.worker(paths, results, progressChan, &wg)
	}

	// Start a goroutine to close the results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	totalScanned := 0
	// Walk the directory tree and send paths to the workers
	go func() {
		err := filepath.Walk(c.rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error accessing path %s: %v\n", path, err)
				return nil
			}
			if info.IsDir() {
				paths <- path
				totalScanned++
				progressChan <- path
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking the file tree: %v\n", err)
		}
		close(paths)
	}()

	// Start a goroutine to print progress
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		var lastPath string
		var maxLength int
		for {
			select {
			case path, ok := <-progressChan:
				if !ok {
					return
				}
				lastPath = path
				if len(path) > maxLength {
					maxLength = len(path)
				}
			case <-ticker.C:
				if lastPath != "" && false {
					paddedPath := fmt.Sprintf("%-*s", maxLength, lastPath)
					fmt.Printf("\rScanning: %s", paddedPath)
				}
			}
		}
	}()

	// Collect results
	var deletionTasks []FolderDeletionTask
	for task := range results {
		deletionTasks = append(deletionTasks, task)
	}

	return deletionTasks, totalScanned
}

// worker processes directories concurrently
func (c *Cleaner) worker(paths <-chan string, results chan<- FolderDeletionTask, progressChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for path := range paths {
		progressChan <- path
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError getting file info for %s: %v\n", path, err)
			continue
		}

		var exists bool
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
			fmt.Fprintf(os.Stderr, "\nError checking folder existence: %v\n", err)
			continue
		}

		if exists {
			size, err := CalculateDirSize(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nError calculating directory size: %v\n", err)
				continue
			}

			fmt.Printf("\nIdentified %s directory for cleanup at %s, will free up %s.\n", info.Name(), path, HumanizeBytes(size))

			results <- FolderDeletionTask{
				Path: path,
				Size: size,
			}
		}
	}
}

// calculateTotalSize calculates the total size of all tasks.
func calculateTotalSize(tasks []FolderDeletionTask) int64 {
	var totalSize int64
	for _, task := range tasks {
		totalSize += task.Size
	}
	return totalSize
}
