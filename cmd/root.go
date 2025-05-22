package cmd

import (
	"fmt"
	"os"

	"github.com/Chrissie21/autodeps/internal/core"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "autodeps",
	Short: "Auto-install dependencies in common project types (Python, Node.js, Go)",
	Run: func(cmd *cobra.Command, args []string) {
		dir, _ := os.Getwd()
		fmt.Println("🔍 Scanning:", dir)
		core.ScanAndInstall(dir)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	}
}
