package cmd

import (
	"fmt"

	"github.com/rigerc/ref-repo-fetch/internal/config"
	"github.com/rigerc/ref-repo-fetch/internal/gitops"
	"github.com/rigerc/ref-repo-fetch/internal/prompt"
	"github.com/rigerc/ref-repo-fetch/internal/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [repo]",
	Short: "Add a repo to config",
	Long: `Add a repository to .externals.json.

With no arguments, launches the interactive wizard.
With a repo argument (owner/repo or URL), adds it directly.
If the config file does not exist, it is created.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return runAddDirect(args[0])
		}
		return runAddInteractive()
	},
}

func runAddInteractive() error {
	projectDir := resolveProjectDir()
	configPath := resolveConfigPath(projectDir)
	cfg := loadOrCreateConfig(configPath)
	return prompt.AddReposInteractive(cfg, configPath)
}

func runAddDirect(repo string) error {
	projectDir := resolveProjectDir()
	configPath := resolveConfigPath(projectDir)
	cfg := loadOrCreateConfig(configPath)

	resolvedURL := config.ResolveURL(repo)
	name := config.NameFromURL(repo)
	if name == "" {
		return fmt.Errorf("could not infer repo name from %q", repo)
	}

	for _, r := range cfg.Repos {
		if r.Name == name {
			return fmt.Errorf("repo %q already exists in config", name)
		}
	}

	detectedBranch, _ := gitops.DetectDefaultBranch(resolvedURL)
	branch := detectedBranch
	if branch == "" {
		branch = "main"
	}

	cfg.Repos = append(cfg.Repos, config.RepoEntry{
		Name:   name,
		URL:    resolvedURL,
		Branch: branch,
		Sparse: []string{},
	})

	if err := config.Save(configPath, cfg); err != nil {
		return err
	}

	ui.LogSuccess(
		"added %s (%s, branch: %s) to %s",
		name,
		resolvedURL,
		branch,
		configPath,
	)
	return nil
}

func loadOrCreateConfig(configPath string) *config.Config {
	cfg, err := config.Load(configPath)
	if err == nil {
		return cfg
	}

	cfg = &config.Config{
		TargetDir: "externals",
		Repos:     []config.RepoEntry{},
	}
	return cfg
}
