package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

func runList() error {
	projectDir := resolveProjectDir()
	configPath := resolveConfigPath(projectDir)
	cfg := loadConfig(configPath)

	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("Target: %s/\n\n", cfg.TargetDir)

	if len(cfg.Repos) == 0 {
		fmt.Println("  (no repos configured)")
		return nil
	}

	fmt.Printf(
		"  %-3s %-20s %-40s %-12s %s\n",
		"#",
		"NAME",
		"URL",
		"BRANCH",
		"SPARSE",
	)
	fmt.Println("  ─────────────────────────────────────────────────────────────────")

	for i, repo := range cfg.Repos {
		sparseLabel := "no"
		if len(repo.Sparse) > 0 {
			sparseLabel = "yes"
		}
		fmt.Printf(
			"  %-3d %-20s %-40s %-12s %s\n",
			i,
			repo.Name,
			repo.URL,
			repo.Branch,
			sparseLabel,
		)
	}

	return nil
}
