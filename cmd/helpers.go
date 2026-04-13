package cmd

import (
	"os"
	"path/filepath"

	"github.com/rigerc/forage/internal/config"
	"github.com/rigerc/forage/internal/ui"
)

func resolveProjectDir() string {
	dir, err := config.ResolveProjectDir()
	if err != nil {
		ui.LogError("%s", err)
		os.Exit(1)
	}
	return dir
}

func resolveConfigPath(projectDir string) string {
	if cfgPath != "" {
		return cfgPath
	}
	return filepath.Join(projectDir, ".externals.json")
}

func loadConfig(configPath string) *config.Config {
	cfg, err := config.Load(configPath)
	if err != nil {
		ui.LogError("%s", err)
		os.Exit(1)
	}
	return cfg
}
