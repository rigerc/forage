# forage

Gather external repositories into your project.

`forage` clones and pulls external Git repositories declared in `.externals.json`, keeping reference source code and documentation up to date.

## Install

```sh
go install github.com/rigerc/forage@latest
```

Or build from source:

```sh
go build -o forage .
```

## Usage

```sh
forage                          # sync all repos (clone missing, pull outdated)
forage add owner/repo           # add a repo directly (GitHub shorthand)
forage add https://git.example.com/repo.git
forage add                      # interactive add
forage remove                   # interactive multi-select remove
forage list                     # list configured repos
forage edit                     # re-run config wizard
forage open                     # open .externals.json in $EDITOR
```

## Config

`.externals.json` lives at your project root:

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

## License

[MIT](LICENSE)
