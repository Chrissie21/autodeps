package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func runScanner(dryRun bool) {
	fmt.Println("🔍 Scanning for dependency files...\n")

	patterns := map[string]func(string, bool){
		"go.mod": func(dir string, dry bool) {
			runCmd("🐹 Go", dir, "go", []string{"mod", "download"}, dry)
		},
		"package.json": func(dir string, dry bool) {
			runCmd("📦 NPM", dir, "npm", []string{"install"}, dry)
		},
		"pnpm-lock.yaml": func(dir string, dry bool) {
			runCmd("⚡ PNPM", dir, "pnpm", []string{"install"}, dry)
		},
		"yarn.lock": func(dir string, dry bool) {
			runCmd("🧶 Yarn", dir, "yarn", []string{"install"}, dry)
		},
		"requirements.txt": func(dir string, dry bool) {
			runCmd("🐍 Python (pip)", dir, "pip", []string{"install", "-r", "requirements.txt"}, dry)
		},
		"Pipfile": func(dir string, dry bool) {
			runCmd("📘 Pipenv", dir, "pipenv", []string{"install"}, dry)
		},
		".venv": func(dir string, dry bool) {
			fmt.Printf("🐍 Found .venv in %s\n", dir)
			// No command for .venv alone, it's an env manager
		},
	}

	root, err := os.Getwd()
	if err != nil {
		fmt.Println("❌ Error getting current dir:", err)
		return
	}

	seen := map[string]bool{}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("⚠️ Walk error:", err)
			return nil
		}

		if !info.IsDir() {
			name := filepath.Base(path)
			handler, ok := patterns[name]
			if ok && !seen[path] {
				dir := filepath.Dir(path)
				fmt.Printf("📁 Found: %s in %s\n", name, dir)
				handler(dir, dryRun)
				seen[path] = true
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("❌ Walk failed:", err)
	}
}

func runCmd(label, dir, command string, args []string, dry bool) {
	fmt.Printf("%s → %s\n", label, dir)
	if dry {
		fmt.Printf("   🔸 Dry-run: %s %v\n\n", command, args)
		return
	}

	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	}
	fmt.Println()
}
