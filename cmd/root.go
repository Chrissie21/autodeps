package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var scan bool

var rootCmd = &cobra.Command{
	Use:   "autodeps",
	Short: "Auto-install project dependencies",
	Long:  `Scan and install dependencies for Python, Node, and Go projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		if scan {
			fmt.Println("🔍 Scanning for project files...")
			scanAndInstallDependencies()
		} else {
			fmt.Println("ℹ️ Use --help to explore available flags.")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&scan, "scan", "s", false, "Scan project directories for dependencies")
}
