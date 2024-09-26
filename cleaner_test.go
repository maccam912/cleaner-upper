package main

import (
	"os"
	"path/filepath"
	"testing"
)

// helper function to create a temporary directory structure with a .venv folder
func createTestDirWithVenv(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "cleaner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a .venv directory inside the temp directory
	venvPath := filepath.Join(dir, ".venv")
	if err := os.Mkdir(venvPath, 0755); err != nil {
		t.Fatalf("Failed to create .venv dir: %v", err)
	}

	// Create a blank pyproject.toml file next to the .venv folder
	pyprojectPath := filepath.Join(dir, "pyproject.toml")
	if _, err := os.Create(pyprojectPath); err != nil {
		t.Fatalf("Failed to create pyproject.toml file: %v", err)
	}

	// Cleanup function to delete the temporary directory
	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

// TestCleaner_Clean tests the Cleaner's Clean method with dryRun and force flags
func TestCleaner_Clean(t *testing.T) {
	tests := []struct {
		name   string
		dryRun bool
		force  bool
	}{
		{"DryRun", true, false},
		{"Force", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup := createTestDirWithVenv(t)
			defer cleanup()

			cleaner := NewCleaner(dir, tt.dryRun, tt.force)
			cleaner.Clean()

			// Check if .venv directory still exists based on the flags
			_, err := os.Stat(filepath.Join(dir, ".venv"))
			if tt.dryRun && os.IsNotExist(err) {
				t.Errorf("Dry run should not delete files")
			} else if !tt.dryRun && !tt.force && !os.IsNotExist(err) {
				t.Errorf("Without force, files should not be deleted without confirmation")
			}
			// Note: This test does not handle user confirmation for deletion. You might need to mock AskForConfirmation or refactor it for better testability.
		})
	}
}
