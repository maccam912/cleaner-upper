package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// CheckFolderAndConfigExists checks if a folder exists and at least one of the config files exists in its parent directory.
func CheckFolderAndConfigExists(parentPath, folderName string, configFiles ...string) (bool, error) {
	folderPath := filepath.Join(parentPath, folderName)
	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	for _, configFile := range configFiles {
		configPath := filepath.Join(parentPath, configFile)
		_, err := os.Stat(configPath)
		if err == nil {
			return true, nil
		}
		if !os.IsNotExist(err) {
			return false, err
		}
	}

	return false, nil
}

// CalculateDirSize calculates the total size of a directory.
func CalculateDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return size, nil
}

// AskForConfirmation prompts the user for confirmation before proceeding.
func AskForConfirmation() bool {
	var response string
	_, err := os.Stdout.WriteString("Are you sure you want to delete? (y/n): ")
	if err != nil {
		return false
	}
	_, err = os.Stdin.Read([]byte(response))
	if err != nil {
		return false
	}
	return response == "y" || response == "Y"
}

// HumanizeBytes converts byte sizes into a human-readable format.
func HumanizeBytes(bytes int64) string {
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
