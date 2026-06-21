---
description: tagbackup project conventions and authoritative specification
alwaysApply: true
---

# tagbackup

`tagbackup` is a single-binary Go CLI for uploading, downloading, listing, and
deleting files on any S3-compatible bucket, using tag-in-filename semantics.
There is no always-on server — only the CLI (and an optional local web UI via
`tagbackup serve`) — and no local index; every operation lists the bucket.

## Specification (authoritative)

Treat these documents as the source of truth for behaviour. Align code with
them and update them when behaviour changes:

- @docs/OVERVIEW.md — project description, tech stack, repo layout, CLI
  convention (top-level verbs vs. `noun verb` subcommands), non-goals.
- @docs/FUNCTIONALITY.md — every command's exact behaviour, flags, exit
  codes, output channels, signal handling, tag grammar.
- @docs/CONFIGURATION.md — YAML layout, file location, atomic save, 0600
  permissions, credential resolution order (env → profile → inline → default
  chain), env var name scheme.

## Repo layout

- `cmd/tagbackup/main.go` — entry point only; dispatches to `internal/cli`.
- `internal/cli` — Cobra commands, flag parsing, global flag glue.
- `internal/config` — YAML loading, validation, schema versioning, atomic save.
- `internal/store` — `ObjectStore` interface, AWS-SDK-v2 implementation, and an
  in-memory fake (`Mem`) for tests.
- `internal/tagexpr` — tag-expression parser/evaluator (pure, well tested).
- `internal/server` — `tagbackup serve` HTTP server and embedded web UI.
- `web/` — React + Vite source for the embedded UI (builds into
  `internal/server/dist`).
- `internal/objectkey`, `internal/duration`, `internal/exitc` — small pure
  helpers.

## Conventions

- **Go 1.25+**, standard `log/slog` for structured stderr logging.
- **Error output:** every user-facing error line is
  `tagbackup: <command>: <message>`, optionally followed by a
  `Hint: <remediation>` line. Use `cli.tagErr` / `cli.tagf` and return an
  `*cli.Exit` with the right `exitc.*` code — never call `os.Exit` from
  commands.
- **Exit codes:** `0` OK, `1` generic, `2` usage, `3` config, `4` S3/network,
  `5` no matches, `130`/`143` signals. Missing required flags **must** be
  `exitc.Usage` (2).
- **Output channels:** stdout is data only (file body, listings, JSON);
  stderr carries progress, logs, prompts, hints, errors.
- **TTY-aware:** colour and progress bars only on a TTY (`StderrIsTTY`);
  prompts refused when stdin is not a TTY (`StdinIsTTY`); `--quiet`
  suppresses non-essential stderr output (including progress bars);
  `--no-color` or `NO_COLOR` disables colour.
- **Non-interactive mode:** `--non-interactive` must never prompt; a command
  that would need to prompt exits with `exitc.Usage` and a clear message.
- **S3 layer:** go through `store.ObjectStore` so tests can use `store.Mem`.
  Use `store.NormalizeAPIEndpoint` for any endpoint passed to the SDK.
- **Keys:** build and parse with `internal/objectkey`; never hand-format the
  `<ts>-<tags>-<filename>` shape elsewhere. Tag chars are `[a-zA-Z0-9]` only.
- **Config writes** go through `config.Save` (atomic, 0600 on Unix). Never
  write YAML directly from command code.

## Coding style

- Follow https://go.dev/doc/effective_go — readable over clever.
- No narrating comments (`// increment i`). Comment only non-obvious intent.
- Prefer small helpers over duplicated logic (tag validation, `Exit`
  construction, ANSI colour wrappers).
- Every new behaviour needs at least one test; favour `store.Mem` for
  command-level tests.
