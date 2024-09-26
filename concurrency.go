package main

import (
	"fmt"
	"os"
	"sync"
)

// FolderDeletionTask represents a task for deleting a folder.
type FolderDeletionTask struct {
	Path string
	Size int64
}

// DeleteFoldersConcurrently deletes the given folders concurrently.
// It returns the total size of all deleted folders.
func DeleteFoldersConcurrently(tasks []FolderDeletionTask, dryRun bool) int64 {
	var wg sync.WaitGroup
	totalSize := int64(0)
	sizeChan := make(chan int64)

	for _, task := range tasks {
		wg.Add(1)
		go func(task FolderDeletionTask) {
			defer wg.Done()
			if dryRun {
				fmt.Printf("Would clean up: %s (Size: %d bytes)\n", task.Path, task.Size)
				sizeChan <- task.Size
			} else {
				err := os.RemoveAll(task.Path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error removing directory: %v\n", err)
					sizeChan <- 0
				} else {
					fmt.Printf("Successfully cleaned up: %s\n", task.Path)
					sizeChan <- task.Size
				}
			}
		}(task)
	}

	go func() {
		wg.Wait()
		close(sizeChan)
	}()

	for size := range sizeChan {
		totalSize += size
	}

	return totalSize
}
