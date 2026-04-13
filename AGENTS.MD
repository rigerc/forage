# Agents

## Fetch Externals

This project includes a Go CLI (`forage`) that clones and pulls external repositories declared in `.externals.json`. Use it to pull in reference source code or documentation that other skills/tools need to inspect.

### Quick Reference

```sh
# Build (run once after changes)
go build -o forage .

# Sync all repos (clone missing, pull outdated)
./forage

# Add a repo directly (owner/repo shorthand or full URL)
./forage add owner/repo
./forage add https://github.com/user/repo.git

# Add repos interactively
./forage add

# Remove repos (multi-select TUI)
./forage remove

# List configured repos
./forage list

# Re-run the full config wizard
./forage edit

# Open .externals.json in $EDITOR
./forage open

# Use a specific config file
./forage --config path/to/.externals.json list
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
