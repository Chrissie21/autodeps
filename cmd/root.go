package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	scan   bool
	dryRun bool
)

var rootCmd = &cobra.Command{
	Use:   "autodeps",
	Short: "🔧 autodeps automatically installs project dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		if scan {
			runScanner(dryRun)
		} else {
			fmt.Println("Use --scan to scan and install project dependencies.")
		}
	},
}

func Execute() {
	rootCmd.PersistentFlags().BoolVar(&scan, "scan", false, "Scan for project dependency files and install them")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Simulate the commands without running them")
	rootCmd.Execute()
}
