# Configuration

tagbackup keeps all of its configuration in a single YAML file per user. The
file is managed by viper and is created or updated by the `tagbackup bucket`
commands; it is not normally hand-edited, but doing so is supported.

The design goal is to be safe enough for real use while remaining usable from
unattended cron jobs, where no human is available to type a password.

## File location

tagbackup looks for `config.yaml` in the standard per-user config directory
for the operating system:

- Linux / macOS: `$XDG_CONFIG_HOME/tagbackup/config.yaml`, falling back to
  `~/.config/tagbackup/config.yaml` when `XDG_CONFIG_HOME` is unset.
- Windows: `%AppData%\tagbackup\config.yaml`.

Viper resolves the path; the directory is created on first write if it does
not already exist.

## File format

The file is YAML. There is one entry per bucket alias under a top-level
`buckets` map. The alias is the value the user passes to `--bucket` on the
command line.

Alias values must match `[a-zA-Z0-9_]+` (uppercase letters, lowercase
letters, digits, and underscore — no hyphens or other punctuation). This
restriction means the alias is a valid env-var fragment without any
translation, so the credential env-var names below are unambiguous and two
aliases can never silently collide.

Example:

```yaml
version: 1

buckets:
  dbbackup:
    bucket: my-company-db-backups
    endpoint: https://s3.eu-west-1.amazonaws.com
    region: eu-west-1
    prefix: nightly/
    credential_type: static
    access_key_id: AKIAEXAMPLE
    secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

  minio_local:
    bucket: backups
    endpoint: http://127.0.0.1:9000
    region: us-east-1
    force_path_style: true
    insecure_skip_verify: true
    credential_type: static
    access_key_id: minioadmin
    secret_access_key: minioadmin

  ec2_role:
    bucket: another-bucket
    endpoint: https://s3.us-east-1.amazonaws.com
    region: us-east-1
    credential_type: iam
```

Top-level fields:

- `version` (required): the configuration schema version. v1 is `1`. This
  field exists so that future incompatible changes to the file layout can be
  migrated automatically. When tagbackup encounters a `version` value newer
  than it supports, it refuses to load the file and prints an error of the
  form `tagbackup: config schema version 2 is newer than this binary
  supports; please upgrade tagbackup`.

Fields per bucket:

- `bucket` (required): the actual S3 bucket name on the provider.
- `endpoint` (required): the S3 endpoint URL. Required because tagbackup
  supports any S3-compatible provider, not just AWS. The endpoint should be
  the **account root** (e.g. `https://<account>.r2.cloudflarestorage.com` for
  Cloudflare R2, `https://s3.eu-west-1.amazonaws.com` for AWS, or
  `http://127.0.0.1:9000` for a local MinIO). tagbackup normalises the
  endpoint at runtime: if the path of the URL is exactly `/<bucket>` matching
  the configured `bucket` value, the path is stripped before the URL is
  passed to the SDK. This is a convenience for providers whose console UIs
  hand out a bucket-scoped URL that the S3 API does not in fact accept.
- `region` (required): the region string the SDK should use. If your
  S3-compatible provider does not care about region, use `us-east-1`.
- `prefix` (optional): a key prefix prepended to every object written or
  listed. Useful for sharing one bucket between several uses.
- `force_path_style` (optional, default `false`): use path-style addressing
  (`https://endpoint/bucket/key`) instead of virtual-host addressing
  (`https://bucket.endpoint/key`). Required for many S3-compatible providers
  including MinIO. AWS S3 itself supports both but recommends virtual-host.
- `insecure_skip_verify` (optional, default `false`): skip TLS certificate
  verification for HTTPS endpoints. Useful for self-signed certificates in
  development. **Do not enable this in production**; it disables protection
  against MITM attacks.
- `credential_type` (required): one of `static`, `profile`, or `iam`.
- `access_key_id`, `secret_access_key`: present only when
  `credential_type: static` and the user chose to store credentials inline.
- `credentials_profile`: present only when `credential_type: profile`,
  naming an entry in `~/.aws/credentials`.

## File permissions

On first write, tagbackup creates `config.yaml` with mode `0600` (read/write
for the owner only). On startup, if the file exists with looser permissions,
tagbackup prints a non-fatal warning to stderr suggesting `chmod 600`.

No permission check is performed on Windows; the default ACLs on
`%AppData%` are used.

Writes to `config.yaml` are atomic: tagbackup writes to a temporary file in
the same directory (created with mode `0600` on Unix), fsyncs it, and then
renames it over the real path. An interrupted write therefore leaves either
the previous good config or the new good config in place, never a partial
file. `os.Rename` is cross-platform in Go so the same code path applies on
Linux, macOS, and Windows.

## Credential resolution order

For each invocation that needs to talk to S3, credentials for the chosen
bucket are resolved in this order. The first match wins.

1. **Environment variables.** If both
   `TAGBACKUP_BUCKET_<ALIAS>_ACCESS_KEY_ID` and
   `TAGBACKUP_BUCKET_<ALIAS>_SECRET_ACCESS_KEY` are set, they are used.
   `<ALIAS>` is the bucket alias, upper-cased. For example, the alias
   `dbbackup` becomes `DBBACKUP`, so the env vars are
   `TAGBACKUP_BUCKET_DBBACKUP_ACCESS_KEY_ID` and
   `TAGBACKUP_BUCKET_DBBACKUP_SECRET_ACCESS_KEY`.
2. **AWS shared-credentials profile.** If the bucket entry sets
   `credential_type: profile` with a `credentials_profile` name, that
   profile is read from `~/.aws/credentials`.
3. **Inline credentials in `config.yaml`.** Used when
   `credential_type: static` and `access_key_id` / `secret_access_key`
   are present in the file.
4. **AWS SDK default credential chain.** Used when
   `credential_type: iam`, or as a final fallback. This picks up
   instance roles on EC2, task roles on ECS, IRSA on EKS, etc., with no
   static secret stored anywhere.

This order means a cautious operator can keep `config.yaml` free of secrets
and inject them via the environment or an attached IAM role, while a casual
user gets the simple "everything is in one file" experience.

## Default behaviour of `bucket add`

`tagbackup bucket add` stores static access keys inline in `config.yaml` by
default, matching the behaviour of `aws configure` writing to
`~/.aws/credentials`. Users who want one of the other resolution paths can
edit `config.yaml` after the fact, or set the relevant environment variables.

## Security guidance

The config file is only as safe as the user's home directory. The strongest
practical mitigation is to scope the credentials themselves, not to encrypt
the file:

- Create a dedicated IAM user or role per machine.
- Grant only the permissions that machine actually needs. A backup server
  typically needs `s3:PutObject` (and possibly `s3:DeleteObject` for the
  `--older-than` cleanup pattern); a developer laptop typically needs
  `s3:GetObject` and `s3:ListBucket` only.
- Restrict the policy to the specific bucket, and to the configured key
  prefix where applicable.
- Rotate keys on a schedule that matches your normal operational
  practice.

The `tagbackup bucket verify` command is a good place to verify that the
permissions you have granted are the ones you intended.

## Out of scope for v1

The following are deliberately not included in the first release. They can
be added later as opt-in features so that unattended cron use is never
affected by them.

- An OS keyring backend (macOS Keychain, Windows Credential Manager,
  Linux libsecret), gated behind a flag.
- A passphrase-encrypted config file for desktop-only users.
