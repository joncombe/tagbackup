# Development

## Install from source

Requires [Go 1.25+](https://go.dev/dl/).

```sh
go install github.com/joncombe/tagbackup/cmd/tagbackup@latest
```

## Build from source

Requires [Go 1.25+](https://go.dev/dl/).

```sh
go build -o tagbackup ./cmd/tagbackup
```

This works out of the box because the web UI for `tagbackup serve` is built into
`internal/server/dist` and committed to the repository.

If you change anything under `web/`, rebuild the UI (requires Node.js + npm) and
re-build the binary:

```sh
make web          # npm install + vite build into internal/server/dist
make build-go     # go build using the freshly built assets
```

`make build` runs both steps in sequence. Use `make run` to build and run.

## Build a new version

Releases are tag-driven — there is no version number in source. Tag the commit you want to release and push it:

```sh
git tag v0.0.5
git push origin v0.0.5
```

That triggers the GitHub Actions release workflow, which runs GoReleaser to build binaries for all platforms and publish them to [Releases](https://github.com/joncombe/tagbackup/releases).

To dry-run locally before tagging, run `make release-check` (validate config) or `make release-snapshot` (build into `dist/` without publishing).
