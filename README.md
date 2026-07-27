# gothub

A read-only GitHub CLI built on [go-gh](https://github.com/cli/go-gh). It only ever issues `GET` requests against the GitHub API — no write operation (create/update/delete) exists anywhere in this codebase, so it cannot perform one, even by mistake or via prompt injection.

## Installation

Download a binary from [GitHub Releases](https://github.com/longkey1/gothub/releases), or build from source:

```sh
go install github.com/longkey1/gothub@latest
```

## Authentication

gothub resolves credentials the same way `gh` does:

1. `GH_TOKEN` or `GITHUB_TOKEN` environment variable
2. `gh`'s own stored auth (`gh auth login`, `~/.config/gh/hosts.yml`)

A fine-grained personal access token with read-only scopes (Contents/Issues/Pull requests) is recommended.

## Usage

```sh
gothub repo view <owner>/<repo>      # show repository details
gothub issue list <owner>/<repo>     # list issues
gothub pr list <owner>/<repo>        # list pull requests
```

Each command accepts:

- `--json` — print the raw JSON response instead of a table
- `--limit <n>` — maximum number of items to fetch (list commands only; default 30)

## Build

```sh
make build   # -> ./bin/gothub
make test
make lint
```

## Release

```sh
make release type=patch|minor|major            # dry run (default)
make release type=patch dryrun=false           # create and push tag
make re-release [tag=vX.Y.Z] dryrun=false      # re-release an existing tag
```

Pushing a `v*` tag triggers `.github/workflows/gorelease.yml`, which builds multi-platform binaries with GoReleaser and uploads them to GitHub Releases.

## License

MIT
