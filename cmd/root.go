package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rigerc/ref-repo-fetch/internal/config"
	"github.com/rigerc/ref-repo-fetch/internal/gitops"
	"github.com/rigerc/ref-repo-fetch/internal/prompt"
	"github.com/rigerc/ref-repo-fetch/internal/ui"
	"github.com/spf13/cobra"
)

var cfgPath string

var rootCmd = &cobra.Command{
	Use:   "fetch-externals",
	Short: "Fetch reference documentation from external repositories",
	Long: `Clone or pull external repositories declared in .externals.json,
keeping project dependencies on reference material up to date.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSync()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.LogError("%s", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file (default is .externals.json)")
	rootCmd.AddCommand(addCmd, removeCmd, listCmd, editCmd, openCmd)
}

func runSync() error {
	projectDir := resolveProjectDir()
	configPath := resolveConfigPath(projectDir)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		ui.LogInfo("no %s found", configPath)
		fmt.Println("  Launching interactive setup...")
		if err := prompt.CreateConfigInteractive(configPath); err != nil {
			return err
		}
		fmt.Println()
	}

	cfg := loadConfig(configPath)
	targetDir := cfg.TargetDir
	if targetDir == "" {
		targetDir = "externals"
	}

	if err := config.ValidateTargetDir(targetDir); err != nil {
		return err
	}

	absTarget := gitops.AbsTarget(projectDir, targetDir)
	absReal := filepath.Clean(filepath.Join(projectDir, targetDir))
	if config.IsUnsafePath(absReal) {
		return fmt.Errorf("resolved target '%s' is unsafe — refusing to operate", absReal)
	}

	if err := os.MkdirAll(absTarget, 0755); err != nil {
		return fmt.Errorf("creating target dir: %w", err)
	}

	if err := config.EnsureGitignore(projectDir, targetDir); err != nil {
		ui.LogWarn("gitignore: %s", err)
	}

	if len(cfg.Repos) == 0 {
		ui.LogWarn("no repositories configured in %s", configPath)
		return nil
	}

	ui.LogInfo("processing %d repos -> %s/", len(cfg.Repos), targetDir)
	fmt.Println()

	var results []ui.Result

	for _, repo := range cfg.Repos {
		dest := filepath.Join(absTarget, repo.Name)
		sparseLabel := "no"
		if len(repo.Sparse) > 0 {
			sparseLabel = "yes"
		}

		if gitops.IsGitRepo(dest) {
			prevHead, _ := gitops.GetHEAD(dest)
			incoming, err := gitops.CountIncoming(dest, repo.Branch)
			if err != nil {
				incoming = 0
			}

			if incoming == 0 {
				ui.LogSuccess("%s: already up to date", repo.Name)
				results = append(results, ui.Result{
					Name: repo.Name, Status: "current",
					Commits: "0", Files: "0", SparseLabel: sparseLabel,
				})
				continue
			}

			ui.LogInfo("%s: pulling %d new commit(s)...", repo.Name, incoming)

			if err := gitops.PullRepo(dest, repo.Branch, repo.Sparse); err != nil {
				ui.LogError("%s: pull failed: %s", repo.Name, err)
				results = append(results, ui.Result{
					Name: repo.Name, Status: "failed",
					Commits: fmt.Sprintf("%d", incoming), Files: "0", SparseLabel: sparseLabel,
				})
				continue
			}

			files, _ := gitops.CountChanges(dest, prevHead)
			ui.LogSuccess("%s: pulled %d commit(s), %d file(s) changed", repo.Name, incoming, files)
			results = append(results, ui.Result{
				Name: repo.Name, Status: "pulled",
				Commits: fmt.Sprintf("%d", incoming), Files: fmt.Sprintf("%d", files), SparseLabel: sparseLabel,
			})
		} else {
			ui.LogInfo("%s: cloning %s (%s)...", repo.Name, repo.URL, repo.Branch)

			if err := gitops.CloneRepo(repo.URL, dest, repo.Branch, repo.Sparse); err != nil {
				ui.LogError("%s: clone failed: %s", repo.Name, err)
				results = append(results, ui.Result{
					Name: repo.Name, Status: "failed",
					Commits: "-", Files: "-", SparseLabel: sparseLabel,
				})
				continue
			}

			ui.LogSuccess("%s: cloned", repo.Name)
			results = append(results, ui.Result{
				Name: repo.Name, Status: "cloned",
				Commits: "-", Files: "-", SparseLabel: sparseLabel,
			})
		}
	}

	ui.PrintSummary(results)
	return nil
}
