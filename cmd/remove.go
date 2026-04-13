package cmd

import (
	"github.com/rigerc/forage/internal/prompt"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Interactively remove repos from config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRemove()
	},
}

func runRemove() error {
	projectDir := resolveProjectDir()
	configPath := resolveConfigPath(projectDir)
	cfg := loadConfig(configPath)
	return prompt.RemoveReposInteractive(cfg, configPath)
}
