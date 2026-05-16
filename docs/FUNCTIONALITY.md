# Managing buckets

- `tagbackup bucket add` - add a new bucket. The user is prompted for the bucket's alias, the S3 bucket name, and the other connection parameters (endpoint URL, region, optional key prefix, whether to use path-style addressing, whether to skip TLS verification). The user is then asked to choose one of the three credential types and is prompted for the corresponding fields:

  - `static` - prompts for the access key ID and secret access key (both stored inline in `config.yaml`).
  - `profile` - prompts for the AWS shared-credentials profile name (stored as `credentials_profile` in `config.yaml`; the secret never enters tagbackup's config file).
  - `iam` - no further prompts; tagbackup will use the AWS SDK default credential chain at runtime (instance role, task role, IRSA, etc.).

This information is stored in the configuration file on that machine and not transmitted anywhere. See [CONFIGURATION.md](CONFIGURATION.md) for the on-disk schema.

Once the bucket has been created, run the `tagbackup bucket verify` functionality described below, except do not present a list of buckets to choose from as we already know we are working with the bucket just added.

Do not allow alias collisions. The alias must match `[a-zA-Z0-9_]+` (uppercase letters, lowercase letters, digits, and underscore — no hyphens or other punctuation).

For unattended provisioning (Ansible, Terraform `local-exec`, Dockerfile setup, etc.), `bucket add` also supports a fully-flag-based form that performs no prompting. The required credential flags depend on `--credential-type`:

- `--credential-type=static` - also requires `--access-key-id=KEY --secret-access-key=SECRET`.
- `--credential-type=profile` - also requires `--credentials-profile=NAME`.
- `--credential-type=iam` - no further credential flags.

Common flags for all three: `--alias=NAME --bucket=S3NAME --endpoint=URL --region=REGION [--prefix=PREFIX] [--force-path-style] [--insecure-skip-verify] [--no-test]`.

When all flags required for the chosen credential type are present, the command runs non-interactively. If any required flag is missing, the command falls back to interactive prompting unless the global `--non-interactive` flag is also set, in which case the command exits with a usage error (exit code 2). The post-add `bucket verify` step is still run automatically; pass `--no-test` to skip it (useful when provisioning a bucket whose credentials have not yet been granted full read/write/delete permission).

- `tagbackup bucket verify` - list the available buckets and allow the user to pick one to verify (pass `--bucket=ALIAS` to skip the picker and verify a specific alias; required when `--non-interactive` is set together with a multi-bucket config). Validate connectivity with that bucket and the read/write/list/delete permissions independently. The software should output a check list such as:

  - Connecting to bucket ✓
  - Listing files in bucket ✓
  - Writing file to bucket ✓
  - Reading file from bucket ✓
  - Deleting file from bucket ✓

The list step is run as its own check so that the common IAM mistake of granting `s3:GetObject` without `s3:ListBucket` is detected. Show errors in red (subject to the global `--no-color` rule and TTY detection), with a friendly pointer of what to do, e.g. "check your read permissions on this bucket." A failed verify run must not be destructive: it is perfectly possible that the user may have intentionally NOT granted one of the permissions, e.g. a server that only pushes files, or a developer laptop that only pulls them.

The probe object used by the write/read/delete steps is named `<prefix>.tagbackup-permission-check-<short-random>`, where `<prefix>` is the bucket entry's configured key prefix (or empty if none) and `<short-random>` is a freshly-generated 8-character hex string. The leading dot keeps it out of normal `tagbackup files` output (which only matches the documented `<timestamp>-<tags>-<filename>` shape).

Cleanup is best-effort: tagbackup attempts to delete the probe object when the verify run finishes (whether earlier steps passed or failed). If the delete step itself fails or was never reached because of a permission denial earlier in the sequence, tagbackup prints a warning identifying the residue object key so the operator can remove it by hand.

- `tagbackup bucket delete` - remove a bucket. The user chooses the bucket from a list.

- `tagbackup bucket list` - list all buckets.

- `tagbackup bucket edit` - edit a bucket. The user chooses the bucket from a list and is then shown a form to edit all information for that bucket, including the alias. Do not allow alias collisions.

Bucket commands are normally interactive. `bucket add` additionally supports a fully-flag-based form for unattended provisioning (described above); the others (`bucket verify`, `bucket delete`, `bucket list`, `bucket edit`) are interactive only.

# Managing files

- `tagbackup push myfile.txt --bucket=mybucket --tag=mytag` - upload a file to the specified bucket alias with the specified tag. The user can only upload one file at a time. The user can enter one or more tags separated by commas without spaces. Tags must be `[a-zA-Z0-9]` only and contain a minimum of one character. Tags are sorted before being written into the filename, so `--tag=b,a` produces the same key segment as `--tag=a,b`. To prevent file overwrites, tagbackup stores the file on the server with the following naming convention:

`<utc-timestamp-with-milliseconds>-<tags-csv>-<original-filename>`
Example: `1776831788343-mytag1,mytag2-myfile.txt`

The timestamp should be 13-digit Unix epoch in milliseconds, UTC, no separators.

Note that filenames can contain hyphens. We know that the timestamp is the first segment, the tag(s) are the second, and anything after that is the filename.

- `--filename=NAME` (required only when reading from stdin) - the original filename to embed in the stored S3 key.

When the file argument is `-`, push reads from standard input. `--filename` is then required because there is no real path to take a name from. Example:

`pg_dump mydb | tagbackup push - --bucket=db --tag=maindb --filename=dump.sql`

When reading from stdin, the upload progress is shown as a bytes-transferred counter rather than a percentage, since the total size is not known up front.

**Filename and key length limits.** The original filename (the basename of the path supplied on the command line, or the value of `--filename` when reading from stdin) must be at most 255 bytes, matching POSIX `NAME_MAX`. The full assembled S3 key (`<prefix><timestamp>-<tags>-<filename>`) must be at most 1024 bytes, the S3 key limit. Push refuses to upload and exits with a usage error (exit code 2) if either limit would be exceeded.

**Sub-millisecond collisions.** The timestamp is millisecond-precision. Two pushes from the same machine in the same millisecond produce the same key, and the second push will overwrite the first. This is a deliberate trade-off in favour of a simple, parseable filename format; in practice push is invoked from human commands or cron tasks separated by seconds, not from tight loops, so collisions are not expected to occur.

Display the upload progress and a success or fail message on completion (subject to the global `--quiet` flag, see Global behaviour). This command is non-interactive.

- `tagbackup pull --bucket=mybucket --tag=mytag --latest` - download a file from the specified bucket alias matching this tag if it exists. --bucket and --tag are required parameters. There are a few caveats here:
  - `--bucket` (required)
  - `--tag` (required)
    - Look at the tags in the filename only. No meta information about the file is used.
    - The files that could possibly be downloaded must match the provided tag. Files without this tag are ignored.
    - The tag parameter follows the "tag grammar" rules (see below).
  - `--latest` (optional)
    - if the --latest flag is provided, download the newest file from the matching list (if it exists).
    - if the --latest flag is omitted, show a list of files, with a human readable datetime and filesize, and the user can select from the list. paginate the list if needbe.
  - `--output=PATH` (optional)
    - write the downloaded file to `PATH` instead of the current working directory. Pass `-` to write the file body to standard output. Example: `tagbackup pull --bucket=db --tag=maindb --latest --output=- | psql mydb`.

Display the download progress and a success or fail message on completion (subject to the global `--quiet` flag, see Global behaviour). When `--output=-` is used, progress output is forced to stderr so it does not contaminate the piped file body. This command is non-interactive only with the inclusion of the --latest flag.

When `--output` is not given, the file is downloaded to the current working directory using the original filename. In the case of a name collision the downloaded file overwrites the existing file. The download is written to a temporary file and atomically renamed on success, so an interrupted download never leaves a half-written file in place.

- `tagbackup files --bucket=mybucket --tag=mytag` - list all the files in the bucket matching the specified bucket and tag(s). Use the tag grammar rules (see below), and output human readable output and pagination as in `tagbackup pull`.
  - `--json` (optional) - emit machine-readable output to stdout, one JSON object per line with `key`, `tags` (an array of strings), `size` (bytes), and `timestamp` (the 13-digit epoch-ms value embedded in the filename). Disables pagination and any interactive prompts; suitable for scripting.

This command is non-interactive.

- `tagbackup tags --bucket=mybucket` - list every tag that appears across all objects in the bucket. For each tag, display:
  - the tag name
  - the number of files carrying that tag
  - the datetime of the oldest file with that tag (derived from the embedded timestamp)
  - the datetime of the newest file with that tag (derived from the embedded timestamp)

Tags are sorted alphabetically. Output is tabular, non-interactive, and always goes to stdout. If the bucket contains no tagbackup objects, the command exits with code 5 (no matches).

- `tagbackup delete --bucket=mybucket --tag=mytag` - delete all the files in the bucket matching the specified bucket and tag(s). Use the tag grammar rules (see below), and human readable output and pagination as in `tagbackup pull`. This command has the following optional parameters:
  - `--force` - delete the files without confirming with the user. If this flag is omitted, the user is shown a list of the files which will be deleted and they must confirm which files from that list they want to delete.
  - `--dry-run` - show which files would be deleted but do not delete them. Combine with `--force` for non-interactive previews and with `--json` for machine-readable output. Exits with code 0 even if the matching list is empty.
  - `--json` - emit machine-readable output to stdout, one JSON object per line for each file deleted (or, under `--dry-run`, for each file that would be deleted). Same field shape as `tagbackup files --json`.
  - `--newer-than` - delete all files that match the above rules and have a UTC timestamp younger than the provided value. The provided value must be an integer immediately followed by a character, e.g. `--newer-than=2d` deletes files newer than 2 days ago, computed from now in UTC. Character options are `s` for seconds, `m` for minutes, `h` for hours, `d` for days and `w` for weeks. The boundary is strictly less-than: a file with a timestamp exactly 2 days old is not matched by `--newer-than=2d`.
  - `--older-than` - the same as `--newer-than` but deletes files older than the provided value, with the same strictly-greater-than boundary.
  - the user can only use one of `--newer-than` or `--older-than`, not both.

Display a list of the deleted files and a success or fail message on completion (subject to the global `--quiet` flag, see Global behaviour). This command is non-interactive only with the inclusion of the `--force` flag.

# Global behaviour

The flags and conventions in this section apply to every command unless explicitly overridden.

## Global flags

These flags can be passed to any command:

- `--config=PATH` - use the configuration file at `PATH` instead of the default location (see [CONFIGURATION.md](CONFIGURATION.md)).
- `--verbose` (`-v`) - print extra diagnostic output to stderr (e.g. SDK retry attempts, skipped non-conforming objects in the bucket).
- `--quiet` (`-q`) - suppress non-essential output. Errors and the final exit code are unaffected. `--verbose` and `--quiet` are mutually exclusive.
- `--non-interactive` - refuse to prompt for input. If a command would normally need to prompt, it exits with a usage error instead. Recommended for cron jobs and other unattended use.
- `--no-color` - disable ANSI colour codes in output. Colour is also disabled automatically when stdout is not a TTY, or when the `NO_COLOR` environment variable is set.
- `--version` - print the tagbackup version and exit.
- `--help` (`-h`) - print help for the current command and exit.

There is no default bucket. The `--bucket` flag is required for every file command (`push`, `pull`, `files`, `tags`, `delete`).

## Exit codes

| Code | Meaning                                                                                                                                                                                 |
| ---- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 0    | Success.                                                                                                                                                                                |
| 1    | Any other unexpected error.                                                                                                                                                             |
| 2    | Usage error: invalid flags, missing required arguments, malformed tag expression, etc.                                                                                                  |
| 3    | Configuration error: missing or unreadable config file, unknown bucket alias, malformed config.                                                                                         |
| 4    | S3 / network error: unable to reach the endpoint, authentication failure, permission denied, request error.                                                                             |
| 5    | No matches: the operation completed successfully but no files matched the supplied tag expression. Useful for cron scripts that need to distinguish "nothing to do" from real failures. |
| 130  | Interrupted by SIGINT (Ctrl-C).                                                                                                                                                         |
| 143  | Interrupted by SIGTERM.                                                                                                                                                                 |

## Output channels

- **stdout** carries data only: the file body for `pull --output=-`, the listing for `files`, the JSON payload for `--json`. Anything you might want to pipe into another command goes here.
- **stderr** carries everything else: progress bars, log messages, prompts, error messages, hints. This means `tagbackup files --bucket=x --tag=y | grep foo` works correctly even when progress output is on.

## TTY detection

Interactive features (colour codes, progress bars, the survey-style prompts in `bucket add`/`bucket edit`/`pull` without `--latest`) are auto-disabled when the relevant stream is not a TTY: colour and progress bars are disabled when stderr is not a TTY, and prompts are refused when stdin is not a TTY. This means cron logs do not fill with escape sequences and spinner garbage. The `NO_COLOR` environment variable also disables colour.

## Signal handling

tagbackup installs a handler for SIGINT (Ctrl-C) and SIGTERM. On receipt:

- An in-progress multipart upload is aborted with `AbortMultipartUpload` so that S3 does not retain partially-uploaded parts (which would otherwise accrue storage charges until a bucket lifecycle policy cleaned them up).
- An in-progress download is aborted and any partial output file is removed, so a half-downloaded file is never left on disk.
- The process exits with code `130` (SIGINT) or `143` (SIGTERM), following standard shell convention.

## Retry policy

Network and S3 calls use the AWS SDK's default retry behaviour (currently up to three attempts per request with exponential backoff and jitter). This applies to listing, uploading (per multipart part), downloading, and deletion. tagbackup does not add its own retry layer on top. With `--verbose`, each retry attempt is logged to stderr.

## Bucket scanning

Every command that operates on existing files (`pull`, `files`, `delete`) issues `ListObjectsV2` calls against the bucket, scoped to the configured `prefix:` (or the whole bucket if no prefix is set). The AWS SDK paginator returns objects 1000 at a time; tagbackup iterates until the bucket is exhausted, parses each key against the `<timestamp>-<tags>-<filename>` format (silently skipping non-conforming keys, or logging each at DEBUG when `--verbose` is set), evaluates the supplied tag expression, and buffers the matched set in memory. For `pull --latest` only the maximum-by-timestamp match is retained. Per-match memory cost is a few hundred bytes (key, parsed tags, size, timestamp), so even buckets with hundreds of thousands of matches are comfortably handled.

## Display pagination

The interactive selector used by `pull` when `--latest` is omitted, and the confirmation list shown by `delete` when `--force` is omitted, page through the matched set 20 entries at a time using survey/v2's built-in scrolling controls. The non-interactive `tagbackup files` command prints all matches in one go without paging — pipe through `less` (or any other pager) if you want one — and `--json` output always streams without paging, one JSON object per line.

## Logging

tagbackup uses the standard library's [`log/slog`](https://pkg.go.dev/log/slog) package for structured logging to stderr. The default level is INFO. `--verbose` raises the level to DEBUG, surfacing SDK retry attempts, skipped non-conforming bucket objects, and the credential resolution path used for the current command. `--quiet` lowers the threshold to WARN, suppressing INFO-level operational messages while leaving warnings and errors intact.

## Configuration sources

Global flags (`--config`, `--verbose`, `--quiet`, `--non-interactive`, `--no-color`, `--version`, `--help`) and command-specific flags are flag-only; tagbackup does not honour environment variables for them. Environment variables are reserved for credentials, as described in [CONFIGURATION.md](CONFIGURATION.md). This keeps the precedence rules predictable and avoids surprising overrides at runtime.

# Tag grammar

When using the `tagbackup pull`, `tagbackup files` and `tagbackup delete` commands, the user can make rules with one or more tags:

- `|` = OR, `+` = AND, `-` = NOT (unary), `()` = grouping.
- Precedence `()` > `-` > `+` > `|`
- Whitespace is not allowed
- Tag characters are [a-zA-Z0-9] only, so |+-() are safely reserved.
- A standalone `-foo` is valid (means "all files lacking foo")

EBNF:

```
expr    = orExpr ;
orExpr  = andExpr { "|" andExpr } ;
andExpr = unary { "+" unary } ;
unary   = [ "-" ] atom ;
atom    = tag | "(" expr ")" ;
tag     = letter { letter | digit } ;
```

Examples:

- The user can enter a pipe to match files with `tag1` OR `tag2`, e.g. `tag1|tag2` would match any file with either tag
- The user can enter a plus sign to match files with a combination of tags, e.g. `tag1+tag2+tag3` would only match files that have all 3 tags
- The user can enter a minus sign to exclude files, e.g. `-tag1|-tag2` would exclude files with either of those tags.
- The user can use brackets to make complex rules, e.g. `(-tag1+-tag2)|tag3`, which finds all files that do have tag3 OR do not have tag1 and tag2.

v1 supports only the operators listed above. There are no wildcards, no regex, and no other special characters. New operators may be added in later versions but will not silently change the meaning of existing valid expressions.

# Notes

- Always use the timestamp in the filename, and not any S3 meta information such as `LastModified`.
- The human-readable datetime shown by `tagbackup files` (and the interactive picker in `tagbackup pull` when `--latest` is omitted) is derived from the embedded filename timestamp, not from S3 `LastModified`.
- tagbackup should ignore all files in the bucket that do not match its own naming convention.

# Handling errors

All error messages are formatted as `tagbackup: <command>: <message>` on stderr, optionally followed by a `Hint: ...` line offering concrete remediation advice. The process then exits with the appropriate code from the [Exit codes](#exit-codes) table in Global behaviour. tagbackup never prints stack traces or unwrapped Go error strings to the user.

The following conditions are explicitly surfaced as friendly errors:

- the syntax of a command is invalid (exit code 2)
- required parameters are missing, e.g. `--bucket` or `--tag` with `pull`, `push`, `files` or `delete` (exit code 2)
- the tag expression is malformed under the [Tag grammar](#tag-grammar) rules (exit code 2)
- the configuration file is missing, unreadable, or names an unknown bucket alias (exit code 3)
- the bucket credentials do not work (exit code 4)
- the bucket credentials do not have permission to perform the operation (exit code 4)
- the endpoint cannot be reached (exit code 4)
- using `push` and the file is not present (exit code 1)
- using `pull` and there is no file to download (exit code 5)
- any other unexpected error, e.g. unable to communicate with the bucket (exit code 1)

# Example usage

Imagine a scenario where production data is managed on a server, and a software engineer uses their development machine.

On both machines:

- install tagbackup (instruction to come later)
- use `tagbackup bucket add` interactive to setup access to the same S3 bucket

On the server, set up a script run by cron. This script performs the following daily:

- (pseudo code) create a backup by dumping the database to `dump.sql`
- `tagbackup push dump.sql --bucket=dbbackup --tag=maindb`
- `tagbackup delete --bucket=dbbackup --tag=maindb --older-than=30d --force`

The `delete` command at the end restricts the number of files to keep in s3.

On the development machine, whenever the engineer wants to see the latest copy of the database they run:

- `tagbackup pull --bucket=dbbackup --tag=maindb --latest`
