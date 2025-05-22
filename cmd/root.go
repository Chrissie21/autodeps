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

type InstallSummary struct {
	Successes []string
	Failures  []string
	Skipped   []string
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

	visitedDirs := map[string]bool{}
	summary := InstallSummary{}

	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		}

		dir := filepath.Dir(path)
		file := info.Name()

		if visitedDirs[dir] {
			return nil
		}

		// --- Python .venv + requirements.txt support ---
		venvPath := filepath.Join(dir, ".venv")
		reqsPath := filepath.Join(dir, "requirements.txt")

		venvExists := false
		if stat, err := os.Stat(venvPath); err == nil && stat.IsDir() {
			venvExists = true
		}

		if _, reqExists := os.Stat(reqsPath); reqExists == nil {
			if only == "" || filter["pip"] {
				fmt.Printf("📁 Found: requirements.txt in %s\n", dir)

				if !venvExists {
					fmt.Println("⚙️  .venv not found — creating it with `python3 -m venv .venv`")
					createCmd := exec.Command("python3", "-m", "venv", ".venv")
					createCmd.Dir = dir
					createCmd.Stdout = os.Stdout
					createCmd.Stderr = os.Stderr

					if dryRun {
						fmt.Printf("   🔸 Dry-run: python3 -m venv .venv\n")
					} else {
						if err := createCmd.Run(); err != nil {
							msg := fmt.Sprintf("❌ Failed to create .venv in %s: %v", dir, err)
							fmt.Println(msg)
							summary.Failures = append(summary.Failures, msg)
							return nil
						}
					}
				}

				script := fmt.Sprintf("source .venv/bin/activate && pip install -r requirements.txt")

				if verbose || dryRun {
					fmt.Printf("   🔸 %s: %s\n", map[bool]string{true: "Dry-run", false: "Executing"}[dryRun], script)
				}

				if dryRun {
					summary.Skipped = append(summary.Skipped, fmt.Sprintf("Dry-run: %s", dir))
				} else {
					cmd := exec.Command("bash", "-c", script)
					cmd.Dir = dir
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						msg := fmt.Sprintf("❌ Error installing Python deps in %s: %v", dir, err)
						fmt.Println(msg)
						summary.Failures = append(summary.Failures, msg)
					} else {
						summary.Successes = append(summary.Successes, fmt.Sprintf("✅ Python deps installed in %s", dir))
					}
				}
				visitedDirs[dir] = true
				return nil
			}
		}

		// --- General project file runners ---
		if runner, ok := runners[file]; ok {
			if only != "" && !filter[runner.Type] {
				return nil
			}

			fmt.Printf("📁 Found: %s in %s\n", file, dir)
			fmt.Printf("🔧 %s → %s\n", runner.Name, dir)

			cmdPath, pathErr := exec.LookPath(runner.Command[0])
			if pathErr != nil {
				msg := fmt.Sprintf("⚠️  Command not found: %s", runner.Command[0])
				fmt.Println(msg)
				summary.Failures = append(summary.Failures, msg)
				return nil
			}

			fullCmd := append([]string{cmdPath}, runner.Command[1:]...)
			if verbose || dryRun {
				fmt.Printf("   🔸 %s: %s\n", map[bool]string{true: "Dry-run", false: "Executing"}[dryRun], strings.Join(fullCmd, " "))
			}

			if dryRun {
				summary.Skipped = append(summary.Skipped, fmt.Sprintf("Dry-run: %s", dir))
			} else {
				cmd := exec.Command(fullCmd[0], fullCmd[1:]...)
				cmd.Dir = dir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					msg := fmt.Sprintf("❌ Error installing deps in %s: %v", dir, err)
					fmt.Println(msg)
					summary.Failures = append(summary.Failures, msg)
				} else {
					summary.Successes = append(summary.Successes, fmt.Sprintf("✅ Installed: %s", runner.Name))
				}
			}

			visitedDirs[dir] = true
		}

		return nil
	})

	if err != nil {
		log.Fatalf("❌ Scan failed: %v", err)
	}

	// ✅ Print Summary
	fmt.Println("\n🔚 Install Summary:")
	fmt.Println("---------------------------")

	if len(summary.Successes) > 0 {
		fmt.Println("✅ Successes:")
		for _, msg := range summary.Successes {
			fmt.Println("   -", msg)
		}
	}

	if len(summary.Skipped) > 0 {
		fmt.Println("\n🚫 Skipped (dry-run):")
		for _, msg := range summary.Skipped {
			fmt.Println("   -", msg)
		}
	}

	if len(summary.Failures) > 0 {
		fmt.Println("\n❌ Failures:")
		for _, msg := range summary.Failures {
			fmt.Println("   -", msg)
		}
	} else {
		fmt.Println("\n🎉 No errors encountered.")
	}

	fmt.Println()
}
