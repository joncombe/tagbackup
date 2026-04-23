# tagbackup

Command-line tool to upload, download, list, and delete files in S3-compatible storage (for example Amazon S3, MinIO, or Cloudflare R2). Files are identified and filtered with **tags**; object keys are generated in a consistent, timestamped format. There is no separate server—only the CLI talking to the bucket you configure.

## Requirements

- [Go 1.25+](https://go.dev/dl/) (see `go.mod`).

## Build

```sh
go build -o tagbackup ./cmd/tagbackup
```

Or use the Makefile: `make build` (or `make run` to build and run).

## Install from source

```sh
go install github.com/joncombe/tagbackup/cmd/tagbackup@latest
```

Prebuilt binaries for common platforms may be published on the [Releases](https://github.com/joncombe/tagbackup/releases) page when available.

## Quick start

1. Create a `config.yaml` in the current directory (or set `TAGBACKUP_CONFIG` to its path). See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for the full format.
2. Add at least one bucket under `buckets:` with an alias, `endpoint`, `region`, `bucket` name, and credentials.
3. Run commands such as `tagbackup push`, `tagbackup pull`, `tagbackup list`, and `tagbackup delete` — see [docs/USAGE.md](docs/USAGE.md) for usage and [docs/FUNCTIONALITY.md](docs/FUNCTIONALITY.md) for the full specification.

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
