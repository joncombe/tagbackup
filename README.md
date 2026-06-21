# tagbackup

Command-line tool to upload, download, list, and delete files in S3-compatible storage (for example Amazon S3, MinIO, or Cloudflare R2). Files are identified and filtered with **tags**; object keys are generated in a consistent, timestamped format. The CLI talks directly to the bucket you configure; optionally, `tagbackup serve` runs a local web UI on `127.0.0.1` for browsing, uploading, and deleting files.

## Install

**Linux and macOS 12+** — run the install script:

```sh
curl -sfL https://tagbackup.com/install.sh | sh
```

Or with `wget`:

```sh
wget -qO- https://tagbackup.com/install.sh | sh
```

The script detects your OS and architecture, downloads the right binary from the [Releases](https://github.com/joncombe/tagbackup/releases) page, and installs it to `/usr/local/bin` (or `~/.local/bin` if that isn't writable).

**Windows** — download the `.zip` for your architecture from the [Releases](https://github.com/joncombe/tagbackup/releases) page and extract `tagbackup.exe` somewhere on your `PATH`.

**Options:**

```sh
# Install a specific version
VERSION=v0.0.3 curl -sfL https://tagbackup.com/install.sh | sh

# Install to a custom directory
INSTALL_DIR=~/bin curl -sfL https://tagbackup.com/install.sh | sh

# Install to a system directory (requires sudo)
curl -sfL https://tagbackup.com/install.sh | sudo sh
```

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

## Quick start

1. Add at least one bucket with `tagbackup bucket add` (interactive), or create the config file at the default per-user path — see [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for the full format and file location. Pass `--config=PATH` to use a different file.
2. Run commands such as `tagbackup push`, `tagbackup pull`, `tagbackup files`, `tagbackup tags`, `tagbackup delete`, and `tagbackup serve` — see [docs/USAGE.md](docs/USAGE.md) for usage and [docs/FUNCTIONALITY.md](docs/FUNCTIONALITY.md) for the full specification.

```sh
./tagbackup --help
./tagbackup push --help
```

## Documentation

| Document                                       | Purpose                             |
| ---------------------------------------------- | ----------------------------------- |
| [docs/OVERVIEW.md](docs/OVERVIEW.md)           | Project summary, tech stack, layout |
| [docs/USAGE.md](docs/USAGE.md)                 | Command-line usage                  |
| [docs/CONFIGURATION.md](docs/CONFIGURATION.md) | `config.yaml` and credentials       |
| [docs/FUNCTIONALITY.md](docs/FUNCTIONALITY.md) | Full functional specification       |

## License

[MIT](LICENSE)
