package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	scanFlag    bool
	dryRunFlag  bool
	verboseFlag bool
	onlyFlag    string
)

var rootCmd = &cobra.Command{
	Use:   "autodeps",
	Short: "Automatically install project dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		if scanFlag {
			fmt.Println("🔍 Scanning for dependency files...\n")
			scanAndInstallDependencies(dryRunFlag, verboseFlag, onlyFlag)
		} else {
			fmt.Println("ℹ️  Use --scan to scan for project dependencies.")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&scanFlag, "scan", false, "Scan and install dependencies")
	rootCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Only show what would be done")
	rootCmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Show full command paths")
	rootCmd.Flags().StringVar(&onlyFlag, "only", "", "Comma-separated list of dependency types to run (e.g., go,pip,npm)")
}

// Map of file -> runner
var runners = map[string]struct {
	Name    string
	Command []string
	Type    string
}{
	"go.mod":           {"Go", []string{"go", "mod", "download"}, "go"},
	"package.json":     {"NPM", []string{"npm", "install"}, "npm"},
	"pnpm-lock.yaml":   {"PNPM", []string{"pnpm", "install"}, "pnpm"},
	"yarn.lock":        {"Yarn", []string{"yarn", "install"}, "yarn"},
	"requirements.txt": {"Python (pip)", []string{"pip", "install", "-r", "requirements.txt"}, "pip"},
	"Pipfile":          {"Pipenv", []string{"pipenv", "install"}, "pipenv"},
	"environment.yml":  {"Conda", []string{"conda", "env", "update", "--file", "environment.yml"}, "conda"},
}

func scanAndInstallDependencies(dryRun, verbose bool, only string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ Could not get current directory: %v", err)
	}

	filter := map[string]bool{}
	if only != "" {
		for _, v := range strings.Split(only, ",") {
			filter[strings.TrimSpace(v)] = true
		}
	}

	visitedDirs := map[string]bool{} // Avoid duplicate installs per folder

	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		}

		dir := filepath.Dir(path)
		file := info.Name()

		// Only run once per directory
		if visitedDirs[dir] {
			return nil
		}

		// Check for .venv + requirements.txt combo
		venvPath := filepath.Join(dir, ".venv")
		reqsPath := filepath.Join(dir, "requirements.txt")
		if _, venvExists := os.Stat(venvPath); !os.IsNotExist(venvExists) {
			if _, reqExists := os.Stat(reqsPath); !os.IsNotExist(reqExists) {
				if only == "" || filter["pip"] {
					fmt.Printf("📁 Found: .venv and requirements.txt in %s\n", dir)
					fmt.Println("🐍 Python (.venv) → Virtual Env Detected")

					script := fmt.Sprintf("source %s/bin/activate && pip install -r requirements.txt", filepath.Join(dir, ".venv"))
					fmt.Printf("   🔸 %s: %s\n\n",
						map[bool]string{true: "Dry-run", false: "Executing"}[dryRun],
						script)

					if !dryRun {
						cmd := exec.Command("bash", "-c", script)
						cmd.Dir = dir
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						if err := cmd.Run(); err != nil {
							fmt.Printf("❌ Error installing in .venv: %v\n", err)
						}
					}
					visitedDirs[dir] = true
					return nil
				}
			}
		}

		// Default file-based runners
		if runner, ok := runners[file]; ok {
			if only != "" && !filter[runner.Type] {
				return nil
			}
			fmt.Printf("📁 Found: %s in %s\n", file, dir)
			fmt.Printf("🔧 %s → %s\n", runner.Name, dir)

			cmdPath, pathErr := exec.LookPath(runner.Command[0])
			if pathErr != nil {
				fmt.Printf("⚠️  Command not found: %s\n", runner.Command[0])
				return nil
			}

			fullCmd := append([]string{cmdPath}, runner.Command[1:]...)
			if verbose {
				fmt.Printf("   🔸 %s: %s\n",
					map[bool]string{true: "Dry-run", false: "Executing"}[dryRun],
					strings.Join(fullCmd, " "),
				)
			} else {
				fmt.Printf("   🔸 %s: %s ...\n",
					map[bool]string{true: "Dry-run", false: "Executing"}[dryRun],
					fullCmd[0],
				)
			}

			if !dryRun {
				cmd := exec.Command(cmdPath, runner.Command[1:]...)
				cmd.Dir = dir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Printf("❌ Error running %s: %v\n", runner.Name, err)
				}
			}
			fmt.Println()
			visitedDirs[dir] = true
		}

		return nil
	})

	if err != nil {
		log.Fatalf("❌ Scan failed: %v", err)
	}
}
