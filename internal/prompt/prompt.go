package prompt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/indaco/prompti"
	"github.com/indaco/prompti/confirm"
	"github.com/indaco/prompti/input"
	"github.com/rigerc/forage/internal/config"
	"github.com/rigerc/forage/internal/gitops"
	"github.com/rigerc/forage/internal/ui"
)

func CollectReposInteractive(existing []config.RepoEntry) ([]config.RepoEntry, error) {
	repos := make([]config.RepoEntry, len(existing))
	copy(repos, existing)

	for {
		fmt.Println()
		fmt.Println(boldStyle.Render("--- Add repository ---"))

		rawURL, err := input.Run(&input.Config{
			Message:     "URL: ",
			Placeholder: "owner/repo or https://github.com/user/repo.git",
		})
		if err != nil {
			if errors.Is(err, prompti.ErrCancelled) {
				return repos, nil
			}
			return repos, err
		}
		if rawURL == "" {
			ui.LogError("repository URL is required")
			continue
		}

		resolvedURL := config.ResolveURL(rawURL)
		inferredName := config.NameFromURL(rawURL)

		namePlaceholder := "my-repo"
		nameInitial := ""
		if inferredName != "" {
			namePlaceholder = inferredName
			nameInitial = inferredName
		}

		name, err := input.Run(&input.Config{
			Message:     "Name: ",
			Placeholder: namePlaceholder,
			Initial:     nameInitial,
		})
		if err != nil {
			if errors.Is(err, prompti.ErrCancelled) {
				return repos, nil
			}
			return repos, err
		}
		if name == "" {
			name = inferredName
		}
		if name == "" {
			ui.LogError("repository name is required")
			continue
		}

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
			if errors.Is(err, prompti.ErrCancelled) {
				return repos, nil
			}
			return repos, err
		}
		if branch == "" {
			branch = "main"
			if detectedBranch != "" {
				branch = detectedBranch
			}
		}

		fullClone, err := confirm.Run(&confirm.Config{
			Mode:     confirm.ModeInline,
			Question: "Fetch full repository?",
		})
		if err != nil {
			if errors.Is(err, prompti.ErrCancelled) {
				return repos, nil
			}
			return repos, err
		}

		sparsePaths := []string{}
		if !fullClone {
			pathsStr, err := input.Run(&input.Config{
				Message:     "Paths: ",
				Placeholder: "src/lib,docs/README.md",
			})
			if err != nil {
				if errors.Is(err, prompti.ErrCancelled) {
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
		if errors.Is(err, prompti.ErrCancelled) {
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

	remaining := []config.RepoEntry{}
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
