# tagbackup

Command-line tool to upload, download, list, and delete files in S3-compatible storage (for example Amazon S3, MinIO, or Cloudflare R2). Files are identified and filtered with **tags**; object keys are generated in a consistent, timestamped format. There is no separate server—only the CLI talking to the bucket you configure.

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

Or use the Makefile: `make build` (or `make run` to build and run).

## Quick start

1. Create a `config.yaml` in the current directory (or set `TAGBACKUP_CONFIG` to its path). See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for the full format.
2. Add at least one bucket under `buckets:` with an alias, `endpoint`, `region`, `bucket` name, and credentials.
3. Run commands such as `tagbackup push`, `tagbackup pull`, `tagbackup files`, `tagbackup tags`, and `tagbackup delete` — see [docs/USAGE.md](docs/USAGE.md) for usage and [docs/FUNCTIONALITY.md](docs/FUNCTIONALITY.md) for the full specification.

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
