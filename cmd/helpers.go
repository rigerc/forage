package cmd

import (
	"fmt"
	"os"

	"github.com/rigerc/ref-repo-fetch/internal/config"
	"github.com/rigerc/ref-repo-fetch/internal/ui"
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
	return fmt.Sprintf("%s/.externals.json", projectDir)
}

func loadConfig(configPath string) *config.Config {
	cfg, err := config.Load(configPath)
	if err != nil {
		ui.LogError("%s", err)
		os.Exit(1)
	}
	return cfg
}

func exitOnError(msg string, err error) {
	if err != nil {
		ui.LogError("%s: %s", msg, err)
		os.Exit(1)
	}
}
