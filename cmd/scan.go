package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func scanAndInstallDependencies() {
	files := map[string]func(string){
		"requirements.txt": installPythonDeps,
		"package.json":     installNodeDeps,
		"go.mod":           installGoDeps,
	}

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if handler, ok := files[filepath.Base(path)]; ok {
				fmt.Printf("📁 Found: %s\n", path)
				handler(filepath.Dir(path))
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("❌ Scan error: %v\n", err)
	}
}

func installPythonDeps(dir string) {
	fmt.Printf("🐍 Installing Python deps in %s...\n", dir)
	cmd := exec.Command("pip", "install", "-r", "requirements.txt")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	printCmdResult(output, err)
}

func installNodeDeps(dir string) {
	fmt.Printf("🟦 Installing Node deps in %s...\n", dir)
	cmd := exec.Command("npm", "install")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	printCmdResult(output, err)
}

func installGoDeps(dir string) {
	fmt.Printf("🐹 Downloading Go deps in %s...\n", dir)
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	printCmdResult(output, err)
}

func printCmdResult(output []byte, err error) {
	if err != nil {
		fmt.Printf("❌ Command failed: %v\n", err)
	}
	fmt.Println(string(output))
}
