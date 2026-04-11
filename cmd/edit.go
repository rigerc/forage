package cmd

import (
	"os"

	"github.com/rigerc/ref-repo-fetch/internal/prompt"
	"github.com/rigerc/ref-repo-fetch/internal/ui"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Re-run the interactive config wizard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runEdit()
	},
}

func runEdit() error {
	projectDir := resolveProjectDir()
	configPath := resolveConfigPath(projectDir)

	if _, err := os.Stat(configPath); err == nil {
		ui.LogWarn("existing config will be replaced")
	}

	return prompt.CreateConfigInteractive(configPath)
}
