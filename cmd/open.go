package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open .externals.json in $EDITOR",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOpen()
	},
}

func runOpen() error {
	projectDir := resolveProjectDir()
	configPath := resolveConfigPath(projectDir)

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
