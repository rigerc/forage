package cmd

import (
	"github.com/rigerc/ref-repo-fetch/internal/prompt"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactively add repos to config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd()
	},
}

func runAdd() error {
	projectDir := resolveProjectDir()
	configPath := resolveConfigPath(projectDir)
	cfg := loadConfig(configPath)
	return prompt.AddReposInteractive(cfg, configPath)
}
