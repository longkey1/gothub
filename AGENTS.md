# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, etc.) when working with code in this repository.

## Project Overview

`gothub` is a read-only GitHub CLI built with Cobra on top of [go-gh](https://github.com/cli/go-gh) (`github.com/cli/go-gh/v2`). It is not a wrapper around the `gh` binary — it talks to the GitHub API directly.

Safety is structural, not policy-based: `internal/client` exposes only `Get` and `GetPaginated`. No write method (`Post`/`Patch`/`Put`/`Delete`) is implemented anywhere in this codebase, so gothub cannot perform a write request against the GitHub API by construction. When adding a command, never add a write method to `internal/client` to support it.

Authentication is resolved by go-gh the same way `gh` resolves it: `GH_TOKEN`/`GITHUB_TOKEN`, falling back to `gh`'s own stored auth (`gh auth login`).

## Build Commands

```sh
make build   # Build binary to ./bin/gothub (with version ldflags)
make test    # Run tests
make fmt     # Format code
make vet     # Vet code
make lint    # Run golangci-lint (version managed by go.mod tool directive)
make tidy    # Tidy dependencies
make clean   # Remove build artifacts
```

Binary name is read from `.product_name`.

## Release

```sh
make release type=patch|minor|major            # dry run (default)
make release type=patch dryrun=false           # create and push tag
make re-release [tag=vX.Y.Z] dryrun=false      # re-release an existing tag
```

Pushing a `v*` tag triggers `.github/workflows/gorelease.yml`, which builds multi-platform binaries with GoReleaser and uploads them to GitHub Releases.

## Architecture

- `main.go` — entry point, calls `cmd.Execute()`
- `cmd/` — Cobra command tree
  - `root.go` — `rootCmd` definition, `--version` wiring to `internal/version`, and shared helpers (`splitRepo` for `OWNER/REPO` arguments, `printJSON` for `--json` output)
  - `repo.go` — `repo view <owner>/<repo>`
  - `issue.go` — `issue list <owner>/<repo>` (filters pull requests out of GitHub's issues endpoint)
  - `pr.go` — `pr list <owner>/<repo>`
  - Each list command supports `--json` (raw JSON) and `--limit` (page-following via `client.GetPaginated`, default 30)
- `internal/client/` — thin wrapper around `api.RESTClient` from go-gh. Only `Get` and the generic `GetPaginated[T]` (follows the `Link: rel="next"` header) are exposed
- `internal/version/` — version info injected via ldflags at build time

Key behavior:

- Every command issues `GET` requests only. There is no code path that mutates GitHub state.
- `issue list` fetches `repos/{owner}/{repo}/issues`, which GitHub also uses for pull requests; entries with a non-nil `pull_request` field are filtered out.
- Pagination follows the REST API's `Link` header (`rel="next"`) rather than relying on a dedicated pagination helper, since go-gh does not provide one.
