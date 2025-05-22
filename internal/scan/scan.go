package scan

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Options struct {
	DryRun  bool
	Verbose bool
	Only    []string
}

func ScanAndInstall(opts Options) {
	files := findProjectFiles(opts.Only)

	if len(files) == 0 {
		fmt.Println("📭 No matching project files found.")
		return
	}

	for _, file := range files {
		fmt.Printf("📁 Found: %s\n", file)
		switch filepath.Base(file) {
		case "requirements.txt":
			handlePython(file, opts)
		case "go.mod":
			handleGo(file, opts)
		case "yarn.lock":
			handleYarn(file, opts)
		case "pnpm-lock.yaml":
			handlePnpm(file, opts)
		}
	}
}

func findProjectFiles(only []string) []string {
	var found []string

	fileTypes := map[string]string{
		"pip":  "requirements.txt",
		"go":   "go.mod",
		"yarn": "yarn.lock",
		"pnpm": "pnpm-lock.yaml",
	}

	for k, v := range fileTypes {
		if len(only) == 1 && only[0] == "" || contains(only, k) {
			if _, err := os.Stat(v); err == nil {
				found = append(found, v)
			}
		}
	}

	return found
}

func handlePython(file string, opts Options) {
	if _, err := os.Stat(".venv"); os.IsNotExist(err) {
		fmt.Println("⚙️  .venv not found. Creating virtual environment...")
		if opts.DryRun {
			fmt.Println("💡 Would run: python3 -m venv .venv")
		} else {
			execCommand("python3", []string{"-m", "venv", ".venv"}, opts)
		}
	}

	if opts.DryRun {
		fmt.Printf("💡 Would install: pip install -r %s\n", file)
	} else {
		fmt.Println("🐍 Installing Python dependencies...")
		execCommand("bash", []string{"-c", "source .venv/bin/activate && pip install -r " + file}, opts)
	}
}

func handleGo(file string, opts Options) {
	if opts.DryRun {
		fmt.Println("💡 Would run: go mod tidy")
	} else {
		fmt.Println("🐹 Installing Go dependencies...")
		execCommand("go", []string{"mod", "tidy"}, opts)
	}
}

func handleYarn(file string, opts Options) {
	if opts.DryRun {
		fmt.Println("💡 Would run: yarn install")
	} else {
		fmt.Println("📦 Installing Yarn packages...")
		execCommand("yarn", []string{"install"}, opts)
	}
}

func handlePnpm(file string, opts Options) {
	if opts.DryRun {
		fmt.Println("💡 Would run: pnpm install")
	} else {
		fmt.Println("📦 Installing pnpm packages...")
		execCommand("pnpm", []string{"install"}, opts)
	}
}

func execCommand(name string, args []string, opts Options) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if opts.Verbose {
		fmt.Printf("🔧 Running: %s %v\n", name, args)
	}
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Command failed:", err)
	}
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
