# tagbackup — user guide (usage)

This document starts with **installing a prebuilt binary from GitHub Releases**, then covers **command-line usage**: flags, tag expressions, and behaviour. It does **not** cover compiling from source (use the [README](../README.md) for that). For the configuration file, credentials, and file paths, see [CONFIGURATION.md](CONFIGURATION.md).

## Install from GitHub Releases

Open **[github.com/joncombe/tagbackup/releases](https://github.com/joncombe/tagbackup/releases)** and download the archive for your OS and CPU. Asset names look like `tagbackup_<version>_<os>_<arch>` — for example `tagbackup_1.2.3_linux_amd64.tar.gz` or `tagbackup_1.2.3_windows_arm64.zip`.

| Your machine | Typical asset |
| -------------- | ------------- |
| Linux x86_64 | `…_linux_amd64.tar.gz` |
| Linux ARM64 (e.g. Graviton, many SBCs) | `…_linux_arm64.tar.gz` |
| macOS Intel | `…_darwin_amd64.tar.gz` |
| macOS Apple Silicon | `…_darwin_arm64.tar.gz` |
| Windows x86_64 | `…_windows_amd64.zip` |
| Windows ARM64 | `…_windows_arm64.zip` |

Each release also includes `tagbackup_<version>_checksums.txt`. You can verify a download before extracting (commands below use `VERSION`, `OS`, and `ARCH` as placeholders — substitute the real values from the filename you chose).

### Linux

```sh
# Example: unpack in the current directory
tar -xzf "tagbackup_${VERSION}_linux_${ARCH}.tar.gz"

# Make sure the binary is executable, then move it somewhere on your PATH
chmod +x tagbackup
sudo mv tagbackup /usr/local/bin/
```

To install without `sudo`, use a personal bin directory (ensure it is on `PATH`):

```sh
mkdir -p "$HOME/bin"
mv tagbackup "$HOME/bin/"
```

Optional checksum check (replace the archive name and use the checksums file from the same release):

```sh
sha256sum -c --ignore-missing <(grep "tagbackup_${VERSION}_linux_${ARCH}.tar.gz" tagbackup_${VERSION}_checksums.txt)
```

### macOS

```sh
tar -xzf "tagbackup_${VERSION}_darwin_${ARCH}.tar.gz"
chmod +x tagbackup
sudo mv tagbackup /usr/local/bin/
```

If the binary was quarantined after download in the browser, clear the quarantine attribute so macOS will run it:

```sh
xattr -d com.apple.quarantine /usr/local/bin/tagbackup
```

Optional checksum:

```sh
shasum -a 256 -c --ignore-missing <(grep "tagbackup_${VERSION}_darwin_${ARCH}.tar.gz" tagbackup_${VERSION}_checksums.txt)
```

### Windows

1. Download the matching `…_windows_<arch>.zip` (for example `windows_amd64.zip` on most PCs).
2. Right-click the file → **Extract All…** (or use **Extract** from the ribbon in File Explorer), and choose a folder.
3. Inside the folder you will find `tagbackup.exe` plus bundled `docs`, `README`, and `LICENSE` files.
4. Run `tagbackup.exe` from that folder, **or** add the folder to your user `PATH` so you can run `tagbackup` from any terminal:
   - **Settings** → **System** → **About** → **Advanced system settings** → **Environment Variables** → under *User variables*, select **Path** → **Edit** → **New** → paste the folder path → confirm with **OK**.

To check the version in PowerShell or Command Prompt:

```text
tagbackup.exe --version
```

Optional: verify the zip’s SHA-256 using the line for that `.zip` in `tagbackup_<version>_checksums.txt` together with `Get-FileHash` in PowerShell or a third-party checksum tool.

---

## Concepts

- **Bucket alias** — A short name you choose (letters, digits, underscore only) that refers to a configured S3 connection. You pass it with `--bucket=ALIAS` on file commands.
- **Tags** — Alphanumeric labels stored in the object key. On upload you pass one or more tags (comma-separated). For download, list, and delete you pass a **tag expression** (see [Tag expressions](#tag-expressions)).
- **Object keys** — Files in the bucket use keys shaped like:  
  `<prefix><utc-epoch-ms>-<tags-csv>-<original-filename>`.  
  Only keys that match this pattern are considered tagbackup objects; other keys are ignored when listing and matching.

Run `tagbackup --help` or `tagbackup <command> --help` for the exact flag list your binary supports.

## Configuration

You need at least one bucket defined in the tagbackup config file before file commands will work. Use `tagbackup bucket add` to create entries interactively, or edit the YAML as described in [CONFIGURATION.md](CONFIGURATION.md). Use `--config=PATH` to point at a specific file; otherwise the default per-OS path is used.

## Global flags

These apply to every subcommand:

| Flag | Short | Meaning |
| ---- | ----- | ------- |
| `--config=PATH` | | Config file (default: standard user config location) |
| `--verbose` | `-v` | More diagnostic output (credentials path, SDK retries, skipped keys) |
| `--quiet` | `-q` | Less operational output; errors still print. Mutually exclusive with `--verbose`. |
| `--non-interactive` | | Exit with a usage error instead of prompting (for scripts) |
| `--no-color` | | Disable colour (also when stderr is not a TTY, or if `NO_COLOR` is set) |
| `--version` | | Print version and exit |
| `--help` | `-h` | Help for the current command |

There is **no** default bucket: `--bucket` is **required** for `push`, `pull`, `files`, `tags`, and `delete`.

## Command overview

| Command | Purpose |
| ------- | ------- |
| `tagbackup bucket add` | Add a bucket entry to the config |
| `tagbackup bucket list` | List configured bucket aliases |
| `tagbackup bucket verify` | Check connectivity and list/read/write/delete permissions |
| `tagbackup bucket edit` | Change a bucket’s settings (including alias) |
| `tagbackup bucket delete` | Remove a bucket entry from the config (does not delete the remote bucket) |
| `tagbackup push` | Upload one file with tags |
| `tagbackup pull` | Download one file matching a tag expression |
| `tagbackup files` | List objects matching a tag expression |
| `tagbackup tags` | List all tags in the bucket with file counts and date ranges |
| `tagbackup delete` | Delete objects matching a tag expression (and optional age filter) |
| `tagbackup serve` | Run a local web UI for browsing, uploading, downloading, and deleting files |

---

## `tagbackup bucket add`

**Interactive (default):** prompts for alias, S3 bucket name, endpoint, region, optional prefix, path-style and TLS options, then credential type (`static`, `profile`, or `iam`) and the fields each type needs.

**Non-interactive:** pass everything with flags. Required in combination: `--alias`, `--bucket` (S3 name), `--endpoint`, `--region`, `--credential-type`. Additional requirements:

- `static` — also `--access-key-id` and `--secret-access-key`
- `profile` — also `--credentials-profile`
- `iam` — no extra credential flags

Common optional flags: `--prefix`, `--force-path-style`, `--insecure-skip-verify`, `--no-test` (skip the automatic post-add `verify` probe).

With `--non-interactive`, if any required flag is missing, the command exits with a usage error. After a successful save, a permission check runs against the bucket unless `--no-test` was set.

---

## `tagbackup bucket list`

Prints one bucket alias per line, sorted. Non-interactive.

---

## `tagbackup bucket verify`

Checks list, get, put, and delete (using a small probe object under your configured prefix; see the detailed spec in [FUNCTIONALITY.md](FUNCTIONALITY.md#managing-buckets)). You are prompted to **choose a bucket** from the list unless you pass `--bucket=ALIAS`.

- With **multiple** aliases and `--non-interactive`, you **must** pass `--bucket=ALIAS`.
- With a **single** alias and `--non-interactive`, that alias is used; if you set `--bucket`, it must match.

---

## `tagbackup bucket edit`

Interactive: pick a bucket, then adjust fields. Does not allow alias collisions.

---

## `tagbackup bucket delete`

Interactive: pick an alias, then confirm removal. Only the local config entry is removed (not the remote bucket). Declining confirmation is a no-op.

---

## `tagbackup push`

```text
tagbackup push <path> --bucket=ALIAS --tag=TAG1,TAG2
```

Uploads **one** file. Tags are comma-separated; order does not matter (tags are sorted in the key). Multiple tags use the same rules as a single tag: each tag is alphanumeric only.

- **Stdin:** use `path` `-` and set `--filename=NAME` (required), since there is no path to take a name from. Example:  
  `pg_dump db | tagbackup push - --bucket=db --tag=main --filename=dump.sql`
- **Progress:** for stdin, you get a byte counter instead of a percentage.
- **Limits:** original filename (basename) at most 255 bytes; full key at most 1024 bytes (including prefix, timestamp, and tags) or the command errors.

`--verbose` / `--quiet` affect non-essential messages for this command.

---

## `tagbackup pull`

```text
tagbackup pull --bucket=ALIAS --tag=EXPRESSION [flags]
```

- `--tag` is a [tag expression](#tag-expressions) (not comma-separated like `push`).

| Flag | Role |
| ---- | ---- |
| `--latest` | Download the single newest matching object (by embedded filename timestamp) |
| `--output=PATH` | Destination path. Omit: download to the current directory using the object’s display name. `-`: write the body to **stdout** (progress goes to stderr). If `PATH` is an existing directory, or ends with a path separator, the file is written **inside** that directory under the object’s name. Parent directories are created as needed. |

- **Without** `--latest`: an interactive chooser lists matches (paged, 20 at a time) with column headers (`TIMESTAMP`, `SIZE`, `FILENAME`, `TAGS`) above the list, unless you cannot use interactive mode — then you need `--non-interactive` with `--latest`.
- **With** `--non-interactive` and no `--latest`, the command fails (cannot prompt).

If no object matches, exit code **5** (no matches).

Example interactive chooser (headers on stderr, above the prompt):

```text
  TIMESTAMP                 SIZE  FILENAME  TAGS
? Choose a file
  2026-06-13 22:34:01Z    1.2 MiB  dump.sql  [maindb]
> 2026-06-08 22:34:01Z    1.1 MiB  dump.sql  [maindb]
```

---

## `tagbackup files`

```text
tagbackup files --bucket=ALIAS --tag=EXPRESSION [--json]
```

- `--json`: one JSON object per line on stdout with `key`, `tags`, `size`, `timestamp` (epoch ms from the key). No header row; no paging; suitable for scripts.

Human-readable output (default) prints a header row and one line per match, sorted newest-first:

```text
TIMESTAMP                 SIZE  FILENAME  TAGS
2026-06-13 22:34:01Z    1.2 MiB  dump.sql  [maindb]
2026-06-08 22:34:01Z    1.1 MiB  dump.sql  [maindb]
```

Non-interactive. Empty match set uses exit code **5**.

---

## `tagbackup tags`

```text
tagbackup tags --bucket=ALIAS
```

Lists every tag that appears in at least one object in the bucket. For each tag, shows the tag name, number of files carrying that tag, and the datetime of the oldest and newest file with that tag. Tags are sorted alphabetically. Output is always to stdout; non-interactive.

```text
TAG          FILES  OLDEST                NEWEST
bpmapi           6  2026-06-08 22:34:01Z  2026-06-13 22:34:01Z
joncombeapi      8  2026-06-07 08:19:23Z  2026-06-13 22:33:01Z
```

Empty bucket (no tagbackup objects) uses exit code **5**.

---

## `tagbackup delete`

```text
tagbackup delete --bucket=ALIAS --tag=EXPRESSION [options]
```

| Flag | Role |
| ---- | ---- |
| `--force` | Delete all matches without confirmation |
| `--dry-run` | Show what would be deleted; no deletion |
| `--json` | Machine-readable lines (same shape as `files --json`); in dry-run, lines are “would delete” |
| `--newer-than=DUR` | Only objects **strictly newer** than *now* minus the duration (e.g. `2d`) |
| `--older-than=DUR` | Only objects **strictly older** than *now* minus the duration (e.g. `30d`) |

Use **either** `--newer-than` or `--older-than`, not both. Duration is an integer plus a unit: `s`, `m`, `h`, `d`, `w` (see [FUNCTIONALITY.md](FUNCTIONALITY.md#managing-files) for boundary rules).

- **Without** `--force`: you get an interactive multi-select of candidates **unless** `--dry-run` (preview only) or you use `--non-interactive` — then you must use `--force` or `--dry-run`.
- **With** `--json`, output is the same field set as `files --json` for each deleted (or would-delete) object.

Non-interactive automation typically uses:  
`--force` or `--dry-run`, and often `--older-than=…` for retention.

---

## `tagbackup serve`

```text
tagbackup serve [--port=PORT] [--no-open]
```

Starts a small local web server that hosts a browser for your buckets, then opens it in your default browser. Use it when you want to browse, upload, download, or delete files visually rather than at the command line.

| Flag | Default | Role |
| ---- | ------- | ---- |
| `--port=PORT` | `3000` | Port to listen on (1–65535). Exits with a usage error if out of range, or an error if the port is already in use. |
| `--no-open` | off | Do not open a browser automatically; just print the URL. |

Behaviour:

- The server binds to **`127.0.0.1` only** — it is never exposed on your network. Even so, treat it as you would any local tool that can read and write your bucket.
- The header shows the tagbackup wordmark only; the footer shows the running binary version and a link to tagbackup.com.
- Unlike the file commands, `serve` does **not** take `--bucket`; every configured bucket appears as a tab. The first alias (alphabetically) is selected automatically. With no buckets configured, the page shows a "no buckets" message.
- An **info (ℹ) button** beside the bucket tabs opens a read-only configuration dialog for the selected bucket (endpoint, region, prefix, credential type/source, masked keys). When credentials come from environment variables, values are not shown. You cannot edit configuration from the web UI — use `tagbackup bucket edit` or edit `config.yaml` directly.
- For the selected bucket you get the available tags as filter buttons, and a table of files (filename, size, timestamp with a relative "… ago", and tags). Click column headers to sort. Click a tag filter button to show only files with that tag; click it again to show all files. Only one tag can be active at a time. Results are paginated (50 files per page); a footer shows the total matching file count and combined size, page controls, the binary version, and a link to tagbackup.com.
- The view loads all of a bucket's objects once and does filtering, sorting, and paging in the browser, so a very large bucket may take a moment to load.
- **Upload** — click the Upload button to reveal a drag-and-drop area (or click to browse). Select one or more files, then choose tags in a dialog: toggle existing bucket tags and/or type new ones. The same tags are applied to every file in the batch. Tags must be `[a-zA-Z0-9]` only; invalid tags are highlighted and block submission. Filenames come from the filesystem; renaming is not supported. A progress bar is shown while uploads run; per-file failures are ignored and the list refreshes when done.
- **Download** — each row has a download button. Click it to save that object under its original filename. There is no tag-expression or "latest" picker; you download the specific file shown in the row.
- **Delete** — checkboxes beside each file (and a select-all toggle in the header) let you pick files on the **current page only**. Click Delete; if nothing is selected, a hint dialog explains that you must check files first. With files selected, confirm in a dialog, then watch a progress bar while deletions run. Per-file failures are ignored and the list refreshes when done.
- Tag editing on existing objects is not supported in the web UI.
- Stop the server with **Ctrl+C**.

```sh
tagbackup serve                 # open the UI on http://127.0.0.1:3000
tagbackup serve --port 8080     # use a different port
tagbackup serve --no-open       # print the URL but do not launch a browser
```

The web UI is embedded in the binary; there is nothing extra to install or run separately.

---

## Tag expressions

Used with `pull`, `files`, and `delete` for `--tag`:

| Token | Meaning |
| ----- | ------- |
| `a\|b` | OR: either tag (pipe, not a shell pipe — quote as needed) |
| `a+b` | AND: must have all |
| `-a` | NOT: must not have that tag (unary) |
| `( … )` | Grouping |

**Precedence (highest to lowest):** `()` grouping, unary `-` (NOT), `+` (AND), `|` (OR). **No** spaces in the expression. Tag characters are `[a-zA-Z0-9]` only; `|`, `+`, `-`, and `()` are syntax.

Examples (quote for the shell as needed):

- `prod` — files with tag `prod`
- `db+prod` — both `db` and `prod`
- `a\|b` — `a` or `b`
- `(-scratch)+release` — has `release` and does not have `scratch`
- `(-a+-b)\|c` — (not `a` and not `b`) or `c`

There are no wildcards or regex; only the operators above.

---

## How output is split

- **stdout** — Data you might pipe: file body (`pull --output=-`), `files` / `files --json`, `delete --json` lines.
- **stderr** — Progress, logs, errors, hints, interactive prompts. Piping `files` to `grep` is safe: noise stays on stderr.

## Exit codes (summary)

| Code | Meaning |
| ---- | ------- |
| 0 | Success |
| 1 | Unexpected error (or I/O error on push, etc.) |
| 2 | Usage / bad flags or expression |
| 3 | Config error |
| 4 | S3 / network / auth error |
| 5 | No matching objects (when that is the outcome) |
| 130 / 143 | SIGINT / SIGTERM |

See [FUNCTIONALITY.md](FUNCTIONALITY.md#global-behaviour) for TTY handling, signal behaviour, retries, and pagination details.

## Examples

**Upload a nightly DB dump and keep 30 days of history** (on a server; paths and alias are illustrative):

```sh
mysqldump mydb > dump.sql
tagbackup push dump.sql --bucket=dbbackup --tag=nightly,prod
tagbackup delete --bucket=dbbackup --tag=nightly+prod --older-than=30d --force
```

**Download the latest matching backup:**

```sh
tagbackup pull --bucket=dbbackup --tag=nightly+prod --latest
```

**Restore over a pipe:**

```sh
tagbackup pull --bucket=dbbackup --tag=nightly+prod --latest --output=- | psql mydb
```

**Scripted listing:**

```sh
tagbackup files --bucket=dbbackup --tag=prod --json | jq -r .key
```

For bucket wiring (endpoints, R2 path normalization, env credentials), see [CONFIGURATION.md](CONFIGURATION.md). For the full normative spec, [FUNCTIONALITY.md](FUNCTIONALITY.md) remains the reference.
