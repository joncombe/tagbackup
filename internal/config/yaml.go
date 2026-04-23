package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidateConfig checks a decoded config. Used after viper.Unmarshal or tests.
func ValidateConfig(c *Cfg) error {
	if c == nil {
		return fmt.Errorf("config: nil")
	}
	if c.Version == 0 {
		return fmt.Errorf("config: missing version")
	}
	if c.Version > SupportedVersion {
		return fmt.Errorf("config schema version %d is newer than this binary supports; please upgrade tagbackup", c.Version)
	}
	if c.Buckets == nil {
		c.Buckets = map[string]Bucket{}
	}
	for alias, b := range c.Buckets {
		if err := ValidateAlias(alias); err != nil {
			return fmt.Errorf("buckets.%s: %w", alias, err)
		}
		if b.Bucket == "" {
			return fmt.Errorf("buckets.%s: bucket name is required", alias)
		}
		if b.Endpoint == "" {
			return fmt.Errorf("buckets.%s: endpoint is required", alias)
		}
		if b.Region == "" {
			return fmt.Errorf("buckets.%s: region is required", alias)
		}
		if err := ValidateBucketCreds(b); err != nil {
			return fmt.Errorf("buckets.%s: %w", alias, err)
		}
	}
	return nil
}

// ValidateBucketCreds checks the credential-type-specific fields of a Bucket.
// Shared by ValidateConfig and the bucket add/edit command paths.
func ValidateBucketCreds(b Bucket) error {
	switch b.CredentialType {
	case "static":
		if b.AccessKeyID == "" || b.SecretAccessKey == "" {
			return fmt.Errorf("static credentials require access_key_id and secret_access_key")
		}
	case "profile":
		if b.CredentialsProfile == "" {
			return fmt.Errorf("credentials_profile is required for profile credentials")
		}
	case "iam":
	case "":
		return fmt.Errorf("credential_type is required")
	default:
		return fmt.Errorf("invalid credential_type %q", b.CredentialType)
	}
	return nil
}

// ClearUnusedCredFields returns a copy of b with credential fields outside the
// active credential_type cleared, so bucket edit never leaves stale secrets
// when the user switches credential types.
func ClearUnusedCredFields(b Bucket) Bucket {
	switch b.CredentialType {
	case "static":
		b.CredentialsProfile = ""
	case "profile":
		b.AccessKeyID = ""
		b.SecretAccessKey = ""
	case "iam":
		b.AccessKeyID = ""
		b.SecretAccessKey = ""
		b.CredentialsProfile = ""
	}
	return b
}

// Save writes cfg to path atomically with 0600 on Unix.
func Save(path string, cfg *Cfg) error {
	if cfg == nil {
		return fmt.Errorf("config: nil cfg")
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tagbackup-config-*.yaml")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// WarnLoosePerms returns true if the file has group/other read bits set (Unix only).
func WarnLoosePerms(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.Mode().Perm()&0o077 != 0
}

// EnvKeyFragment is the upper-cased alias for TAGBACKUP_BUCKET_<U>_... env var names.
func EnvKeyFragment(alias string) string {
	return strings.ToUpper(alias)
}
