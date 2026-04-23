package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/viper"
)

const (
	SupportedVersion = 1
	configSubdir     = "tagbackup"
	defaultFile      = "config.yaml"
)

var aliasRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Cfg is the top-level config file.
// Viper.Unmarshal uses mapstructure, not yaml tags; mapstructure tags must match config keys.
type Cfg struct {
	Version int               `yaml:"version" mapstructure:"version"`
	Buckets map[string]Bucket `yaml:"buckets" mapstructure:"buckets"`
}

// Bucket is one bucket entry.
type Bucket struct {
	Bucket              string `yaml:"bucket" mapstructure:"bucket"`
	Endpoint            string `yaml:"endpoint" mapstructure:"endpoint"`
	Region              string `yaml:"region" mapstructure:"region"`
	Prefix              string `yaml:"prefix,omitempty" mapstructure:"prefix,omitempty"`
	ForcePathStyle      bool   `yaml:"force_path_style,omitempty" mapstructure:"force_path_style,omitempty"`
	InsecureSkipVerify  bool   `yaml:"insecure_skip_verify,omitempty" mapstructure:"insecure_skip_verify,omitempty"`
	CredentialType      string `yaml:"credential_type" mapstructure:"credential_type"`
	AccessKeyID         string `yaml:"access_key_id,omitempty" mapstructure:"access_key_id,omitempty"`
	SecretAccessKey     string `yaml:"secret_access_key,omitempty" mapstructure:"secret_access_key,omitempty"`
	CredentialsProfile  string `yaml:"credentials_profile,omitempty" mapstructure:"credentials_profile,omitempty"`
}

// DefaultConfigPath returns the default config file path for this OS user.
func DefaultConfigPath() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, configSubdir, defaultFile), nil
}

// ResolvePath returns path, or the default if path is empty.
func ResolvePath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	return DefaultConfigPath()
}

// Load reads the configuration from path. If path is empty, the default is used.
func Load(path string) (cfg *Cfg, resolved string, err error) {
	resolved, err = ResolvePath(path)
	if err != nil {
		return nil, "", err
	}
	v := viper.New()
	v.SetConfigFile(resolved)
	if rerr := v.ReadInConfig(); rerr != nil {
		return nil, resolved, rerr
	}
	var c Cfg
	if err := v.Unmarshal(&c); err != nil {
		return nil, resolved, err
	}
	if err := ValidateConfig(&c); err != nil {
		return nil, resolved, err
	}
	return &c, resolved, nil
}

// LoadOrEmpty returns a config: if the file is missing, a valid empty config is returned.
func LoadOrEmpty(path string) (cfg *Cfg, resolved string, err error) {
	resolved, err = ResolvePath(path)
	if err != nil {
		return nil, "", err
	}
	if _, e := os.Stat(resolved); e != nil && os.IsNotExist(e) {
		return &Cfg{Version: SupportedVersion, Buckets: map[string]Bucket{}}, resolved, nil
	}
	v := viper.New()
	v.SetConfigFile(resolved)
	if rerr := v.ReadInConfig(); rerr != nil {
		return nil, resolved, rerr
	}
	var c Cfg
	if err := v.Unmarshal(&c); err != nil {
		return nil, resolved, err
	}
	if err := ValidateConfig(&c); err != nil {
		return nil, resolved, err
	}
	return &c, resolved, nil
}

// GetBucket returns the bucket by alias, or an error.
func (c *Cfg) GetBucket(alias string) (Bucket, error) {
	if c == nil || c.Buckets == nil {
		return Bucket{}, fmt.Errorf("no buckets configured")
	}
	b, ok := c.Buckets[alias]
	if !ok {
		return Bucket{}, fmt.Errorf("unknown bucket alias %q", alias)
	}
	return b, nil
}

// ValidateAlias returns an error if s is not a valid bucket alias.
func ValidateAlias(s string) error {
	if s == "" {
		return fmt.Errorf("alias is required")
	}
	if !aliasRe.MatchString(s) {
		return fmt.Errorf("alias must match [a-zA-Z0-9_]+")
	}
	return nil
}
