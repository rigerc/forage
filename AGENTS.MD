# Agents

## Fetch Externals

This project includes a Go CLI (`fetch-externals`) that clones and pulls external repositories declared in `.externals.json`. Use it to pull in reference source code or documentation that other skills/tools need to inspect.

### Quick Reference

```sh
# Build (run once after changes)
go build -o fetch-externals .

# Sync all repos (clone missing, pull outdated)
./fetch-externals

# Add a repo directly (owner/repo shorthand or full URL)
./fetch-externals add owner/repo
./fetch-externals add https://github.com/user/repo.git

# Add repos interactively
./fetch-externals add

# Remove repos (multi-select TUI)
./fetch-externals remove

# List configured repos
./fetch-externals list

# Re-run the full config wizard
./fetch-externals edit

# Open .externals.json in $EDITOR
./fetch-externals open

# Use a specific config file
./fetch-externals --config path/to/.externals.json list
```

### Config Format

`.externals.json` lives at the project root:

```json
{
  "target_dir": "externals",
  "repos": [
    {
      "name": "my-lib",
      "url": "https://github.com/org/my-lib.git",
      "branch": "main",
      "sparse": []
    },
    {
      "name": "docs-only",
      "url": "https://github.com/org/big-repo.git",
      "branch": "main",
      "sparse": ["docs/", "README.md"]
    }
  ]
}
```

- `target_dir` — where repos are cloned (relative to project root, auto-added to `.gitignore`)
- `sparse` — list of paths to check out (empty = full clone)

### Source Layout

```
main.go                          # entry point
cmd/                             # cobra commands (root, add, remove, list, edit, open)
internal/config/config.go        # config types, JSON, gitignore, URL/name helpers
internal/gitops/gitops.go        # git operations via go-git
internal/prompt/prompt.go        # interactive flows (prompti input/confirm)
internal/prompt/multiselect.go   # multi-select TUI for remove
internal/ui/summary.go           # colored output, summary table
```
