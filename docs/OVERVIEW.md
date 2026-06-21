# Description

The tool is called tagbackup and allows the users to upload, download, list and delete files on any s3-compatible bucket. It uses a concept of tags to identify and filter the files.

# Example use case

- a cronjob runs on a server. it creates a backup of a database to a file, then uses tagbackup to upload it with the tag "my-database"
- a developer uses tagbackup to download the latest backup of the database, by requesting the most recent file with the tag "my-database"

The only API calls are between tagbackup and the S3 bucket it is currently working with. There is no telemetry or any other "dial home" functionality.

# License

MIT. tagbackup is a FOSS project; contributions are welcome on GitHub.

# Technology

- Go, with `go 1.25` as the minimum supported version (declared in `go.mod`).
- [Cobra](https://github.com/spf13/cobra) for the command tree.
- [Viper](https://github.com/spf13/viper) for configuration loading.
- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) for all S3-compatible API calls.
- [`survey/v2`](https://github.com/AlecAivazis/survey) for interactive prompts and pickers.
- [`schollz/progressbar/v3`](https://github.com/schollz/progressbar) for upload and download progress bars.
- The standard library's [`log/slog`](https://pkg.go.dev/log/slog) package for structured logging to stderr.

The Go module path is `github.com/joncombe/tagbackup`. The compiled binary is named `tagbackup`.

# Repository layout

- `cmd/tagbackup/main.go` - the binary entry point; only wires cobra and dispatches to the cli package.
- `internal/cli` - cobra command definitions, flag parsing, and the global-flag glue.
- `internal/config` - YAML loading, schema validation, schema versioning, and credential resolution.
- `internal/store` - the S3 layer: an `ObjectStore` interface plus the AWS-SDK-v2-backed implementation. Tests use a fake implementation of the interface.
- `internal/tagexpr` - the tag expression parser and evaluator. Pure code, no external dependencies, fully unit-tested.
- `internal/server` - the `tagbackup serve` HTTP server: a JSON API over `internal/config` and `internal/store`, plus the single-page web UI embedded from `internal/server/dist` via `//go:embed`.
- `web/` - the React + Vite + TypeScript source for the embedded UI. It builds into `internal/server/dist` (committed) so `go build`/`go install` need no Node toolchain.

# CLI conventions

The CLI follows two consistent patterns, mirroring `git`:

- **Top-level verbs** for the common, frequently-used actions: `tagbackup push`, `tagbackup pull`, `tagbackup files`, `tagbackup tags`, `tagbackup delete`, `tagbackup serve`.
- **Noun-then-verb subcommands** for managing configuration objects: `tagbackup bucket add|edit|delete|list|verify`.

New commands introduced after v1 should follow the same convention: a top-level verb if it is a hot-path action operating on files, or `tagbackup <noun> <verb>` if it manages a longer-lived configuration object.

# Coding style

Please follow the Go coding style guide: https://go.dev/doc/effective_go, with an emphasis on code readability and maintainability. I'd prefer something sensible and easy to understand, over something that is complex or clever.

# Configuration

tagbackup stores its configuration in a single YAML file per user. See [CONFIGURATION.md](CONFIGURATION.md) for the full layout, file location, and credential resolution rules.

# Installation

The end result of this project is to be able to generate a stand-alone command line tool usable on MacOS, Windows and Linux via GitHub Releases.

# Non-goals

- TagBackup does not deal with encryption. That is the responsibility of the user.
- TagBackup keeps no local index. Every operation lists the bucket.
- TagBackup is not a sync tool.
