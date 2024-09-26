package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cleaner [path]",
	Short: "Cleaner is a tool to clean up files.",
	Long:  `A longer description of your command line tool.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runCleaner,
}

var dryRun bool
var force bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Calculate the space that would be cleaned up without actually deleting files.")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Force delete without asking for confirmation.")
}

func runCleaner(cmd *cobra.Command, args []string) {
	var rootPath string
	if len(args) > 0 {
		rootPath = args[0]
	} else {
		var err error
		rootPath, err = os.Getwd()
		if err != nil {
			fmt.Println("Error getting current directory:", err)
			os.Exit(1)
		}
	}

	cleaner := NewCleaner(rootPath, dryRun, force)

	fmt.Printf("Starting cleanup in: %s\n", rootPath)
	if dryRun {
		fmt.Println("Running in dry-run mode. No files will be deleted.")
	}

	cleaner.Clean()
}
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
