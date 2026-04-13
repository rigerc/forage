package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	TargetDir string      `json:"target_dir"`
	Repos     []RepoEntry `json:"repos"`
}

type RepoEntry struct {
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Branch string   `json:"branch"`
	Sparse []string `json:"sparse"`
}

var ownerRepoPattern = regexp.MustCompile(`^[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+$`)

var unsafePaths = map[string]bool{
	"/":     true,
	"/home": true,
	"/tmp":  true,
}

func IsUnsafePath(p string) bool {
	cleaned := filepath.Clean(p)
	return unsafePaths[cleaned]
}

func ResolveURL(input string) string {
	if ownerRepoPattern.MatchString(input) {
		return "https://github.com/" + input + ".git"
	}
	return input
}

func NameFromURL(rawURL string) string {
	base := rawURL
	if ownerRepoPattern.MatchString(rawURL) {
		base = "https://github.com/" + rawURL + ".git"
	}
	base = strings.TrimSuffix(filepath.Base(base), ".git")
	if base == "" || base == "." || base == ".." {
		return ""
	}
	return base
}

func ValidateTargetDir(dir string) error {
	cleaned := filepath.Clean(dir)
	if cleaned == "/" || cleaned == ".." {
		return fmt.Errorf("target directory %q is unsafe", dir)
	}
	return nil
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.TargetDir == "" {
		cfg.TargetDir = "externals"
	}
	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func EnsureGitignore(projectDir, targetDir string) error {
	gitignorePath := filepath.Join(projectDir, ".gitignore")
	entry := targetDir + "/"

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			content := "# managed by forage\n" + entry + "\n"
			return os.WriteFile(gitignorePath, []byte(content), 0644)
		}
		return fmt.Errorf("reading .gitignore: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == entry || strings.Contains(line, entry) {
			return nil
		}
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening .gitignore: %w", err)
	}
	defer f.Close()

	content := "\n# managed by forage\n" + entry + "\n"
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("writing .gitignore: %w", err)
	}
	return nil
}

func ResolveProjectDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolving working directory: %w", err)
	}

	dir := cwd
	for {
		if fileExists(filepath.Join(dir, ".externals.json")) {
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			dir = cwd
			break
		}
		dir = parent
	}

	if IsUnsafePath(dir) {
		return "", fmt.Errorf("resolved project dir is %q — refusing to operate", dir)
	}

	return dir, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
