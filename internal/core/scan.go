package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ScanAndInstall(baseDir string) {
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Detect and handle known dependency files
		switch strings.ToLower(info.Name()) {
		case "requirements.txt":
			fmt.Println("📦 Found Python project in:", filepath.Dir(path))
			runCommand("pip", []string{"install", "-r", path}, filepath.Dir(path))
		case "package.json":
			fmt.Println("📦 Found Node.js project in:", filepath.Dir(path))
			runCommand("npm", []string{"install"}, filepath.Dir(path))
		case "go.mod":
			fmt.Println("📦 Found Go project in:", filepath.Dir(path))
			runCommand("go", []string{"mod", "tidy"}, filepath.Dir(path))
		}
		return nil
	})

	if err != nil {
		fmt.Println("❌ Error scanning:", err)
	}
}

func runCommand(command string, args []string, dir string) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("➡️  Running: %s %s\n", command, strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to run %s: %v\n", command, err)
	}
}
