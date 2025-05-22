package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func scanAndInstallDependencies() {
	patterns := map[string]func(string){
		"go.mod": func(dir string) {
			fmt.Println("🐹 Downloading Go deps in", dir)
			cmd := exec.Command("go", "mod", "download")
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println("❌ Go deps failed in", dir, ":", err)
			}
		},
		"package.json": func(dir string) {
			fmt.Println("📦 Installing NPM packages in", dir)
			cmd := exec.Command("npm", "install")
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println("❌ NPM install failed in", dir, ":", err)
			}
		},
		"requirements.txt": func(dir string) {
			fmt.Println("🐍 Installing Python packages in", dir)
			cmd := exec.Command("pip", "install", "-r", "requirements.txt")
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println("❌ Python install failed in", dir, ":", err)
			}
		},
	}

	root, err := os.Getwd()
	if err != nil {
		fmt.Println("❌ Error getting current dir:", err)
		return
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("⚠️ Walk error:", err)
			return nil
		}

		if !info.IsDir() {
			for filename, handler := range patterns {
				if filepath.Base(path) == filename {
					dir := filepath.Dir(path) // ✅ FIXED: get the correct directory
					fmt.Println("📁 Found:", filename)
					handler(dir)
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("❌ Walk failed:", err)
	}
}
