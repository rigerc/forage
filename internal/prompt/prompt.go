package prompt

import (
	"fmt"
	"strings"

	"github.com/indaco/prompti/confirm"
	"github.com/indaco/prompti/input"
	"github.com/rigerc/ref-repo-fetch/internal/config"
	"github.com/rigerc/ref-repo-fetch/internal/gitops"
	"github.com/rigerc/ref-repo-fetch/internal/ui"
)

func CollectReposInteractive(existing []config.RepoEntry) ([]config.RepoEntry, error) {
	repos := make([]config.RepoEntry, len(existing))
	copy(repos, existing)

	for {
		fmt.Println()
		fmt.Println(boldStyle.Render("--- Add repository ---"))

		name, err := input.Run(&input.Config{
			Message:     "Name: ",
			Placeholder: "my-repo",
		})
		if err != nil {
			if strings.Contains(err.Error(), "cancelled") {
				return repos, nil
			}
			return repos, err
		}
		if name == "" {
			ui.LogError("repository name is required")
			continue
		}

		rawURL, err := input.Run(&input.Config{
			Message:     "URL: ",
			Placeholder: "owner/repo or https://github.com/user/repo.git",
		})
		if err != nil {
			if strings.Contains(err.Error(), "cancelled") {
				return repos, nil
			}
			return repos, err
		}
		if rawURL == "" {
			ui.LogError("repository URL is required")
			continue
		}

		resolvedURL := config.ResolveURL(rawURL)

		detectedBranch, _ := gitops.DetectDefaultBranch(resolvedURL)

		branchHeader := "Branch to track"
		branchPlaceholder := "main"
		branchInitial := ""
		if detectedBranch != "" {
			branchHeader = fmt.Sprintf("Branch to track (detected: %s)", detectedBranch)
			branchPlaceholder = detectedBranch
			branchInitial = detectedBranch
		}

		branch, err := input.Run(&input.Config{
			Message:     "Branch: " + branchHeader,
			Placeholder: branchPlaceholder,
			Initial:     branchInitial,
		})
		if err != nil {
			if strings.Contains(err.Error(), "cancelled") {
				return repos, nil
			}
			return repos, err
		}
		if branch == "" {
			if detectedBranch != "" {
				branch = detectedBranch
			} else {
				branch = "main"
			}
		}

		useSparse, err := confirm.Run(&confirm.Config{
			Mode:     confirm.ModeInline,
			Question: "Fetch only specific paths (sparse checkout)?",
		})
		if err != nil {
			if strings.Contains(err.Error(), "cancelled") {
				return repos, nil
			}
			return repos, err
		}

		var sparsePaths []string
		if useSparse {
			pathsStr, err := input.Run(&input.Config{
				Message:     "Paths: ",
				Placeholder: "src/lib,docs/README.md",
			})
			if err != nil {
				if strings.Contains(err.Error(), "cancelled") {
					return repos, nil
				}
				return repos, err
			}
			if pathsStr != "" {
				for _, p := range strings.Split(pathsStr, ",") {
					p = strings.TrimSpace(p)
					if p != "" {
						sparsePaths = append(sparsePaths, p)
					}
				}
			}
		}

		repos = append(repos, config.RepoEntry{
			Name:   name,
			URL:    resolvedURL,
			Branch: branch,
			Sparse: sparsePaths,
		})

		fmt.Println()
		addAnother, err := confirm.Run(&confirm.Config{
			Mode:     confirm.ModeInline,
			Question: "Add another repository?",
		})
		if err != nil || !addAnother {
			break
		}
	}

	return repos, nil
}

func CreateConfigInteractive(configPath string) error {
	targetDir, err := input.Run(&input.Config{
		Message:     "Target directory: ",
		Placeholder: "externals",
		Initial:     "externals",
	})
	if err != nil {
		return err
	}
	if targetDir == "" {
		targetDir = "externals"
	}

	if err := config.ValidateTargetDir(targetDir); err != nil {
		return err
	}

	repos, err := CollectReposInteractive(nil)
	if err != nil {
		return err
	}

	cfg := &config.Config{
		TargetDir: targetDir,
		Repos:     repos,
	}

	if err := config.Save(configPath, cfg); err != nil {
		return err
	}

	ui.LogSuccess("created %s", configPath)
	return nil
}

func RemoveReposInteractive(cfg *config.Config, configPath string) error {
	if len(cfg.Repos) == 0 {
		ui.LogWarn("no repos configured")
		return nil
	}

	names := make([]string, len(cfg.Repos))
	for i, r := range cfg.Repos {
		names[i] = r.Name
	}

	selected, err := RunMultiSelect("Select repos to remove (space to toggle, enter to confirm)", names)
	if err != nil {
		if err.Error() == "cancelled" {
			ui.LogInfo("nothing selected")
			return nil
		}
		return err
	}

	if len(selected) == 0 {
		ui.LogInfo("nothing selected")
		return nil
	}

	removeSet := make(map[string]bool)
	for _, name := range selected {
		removeSet[name] = true
	}

	var remaining []config.RepoEntry
	for _, r := range cfg.Repos {
		if removeSet[r.Name] {
			ui.LogWarn("removed: %s", r.Name)
		} else {
			remaining = append(remaining, r)
		}
	}

	cfg.Repos = remaining
	if err := config.Save(configPath, cfg); err != nil {
		return err
	}

	ui.LogSuccess("updated %s", configPath)
	return nil
}

func AddReposInteractive(cfg *config.Config, configPath string) error {
	repos, err := CollectReposInteractive(cfg.Repos)
	if err != nil {
		return err
	}

	cfg.Repos = repos
	if err := config.Save(configPath, cfg); err != nil {
		return err
	}

	ui.LogSuccess("updated %s", configPath)
	return nil
}
