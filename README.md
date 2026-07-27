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

A fine-grained personal access token with read-only scopes (Contents/Issues/Pull requests/Actions) is recommended.

## Usage

```sh
# repo
gothub repo view <owner>/<repo>              # show repository details
gothub repo list [<owner>]                   # list repositories (defaults to the authenticated user)

# issue
gothub issue list <owner>/<repo>             # list issues
gothub issue view <owner>/<repo> <number>    # show issue details

# pr
gothub pr list <owner>/<repo>                # list pull requests
gothub pr view <owner>/<repo> <number>       # show pull request details
gothub pr diff <owner>/<repo> <number>       # show changed files/diff
gothub pr checks <owner>/<repo> <number>     # show CI check status

# run
gothub run list <owner>/<repo>               # list workflow runs
gothub run view <owner>/<repo> <run-id>      # show a workflow run and its jobs

# search
gothub search repos <query>...               # search repositories
gothub search issues <query>...              # search issues
gothub search prs <query>...                 # search pull requests
gothub search code <query>...                # search code
```

Every command accepts:

- `--json` — print the raw JSON response instead of a table
- `--limit <n>` — maximum number of items to fetch (list/search commands only; default 30)

Some commands have additional flags:

- `issue view --comments` — include comments
- `pr view --comments` — include review comments
- `pr diff --name-only` — print only the names of changed files
- `run view --log` — print logs for all jobs
- `run view --log-failed` — print logs for failed jobs only

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
