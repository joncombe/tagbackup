package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/joncombe/tagbackup/internal/config"
	"github.com/spf13/cobra"
)

// cmdBucketEdit reuses the same flag shape as "add" plus --from-alias for non-interactive rewrites.
func (g *Runtime) cmdBucketEdit() *cobra.Command {
	var (
		from                                        string
		alias, bucketName, endpoint, region, prefix string
		credType, accessKey, secret, profile        string
		forcePath, insecure                         bool
	)
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a configured bucket (alias, endpoint, credentials, …)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.runBucketEdit(cmd, from, &bucketAddFl{
				Alias:            alias,
				Bucket:           bucketName,
				Endpoint:         endpoint,
				Region:           region,
				Prefix:           prefix,
				CredentialType:   credType,
				AccessKey:        accessKey,
				Secret:           secret,
				Profile:          profile,
				ForcePath:        forcePath,
				Insecure:         insecure,
				NoTest:           true, // do not re-run S3 verify on every edit; user can `bucket verify`
			})
		},
	}
	cmd.Flags().StringVar(&from, "from-alias", "", "current alias to edit (required when --non-interactive)")
	cmd.Flags().StringVar(&alias, "alias", "", "new or same alias for this entry")
	cmd.Flags().StringVar(&bucketName, "bucket", "", "S3 bucket name")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "S3 endpoint URL")
	cmd.Flags().StringVar(&region, "region", "", "region")
	cmd.Flags().StringVar(&prefix, "prefix", "", "optional key prefix")
	cmd.Flags().StringVar(&credType, "credential-type", "", "static|profile|iam")
	cmd.Flags().StringVar(&accessKey, "access-key-id", "", "for static")
	cmd.Flags().StringVar(&secret, "secret-access-key", "", "for static (omit to keep existing when non-interactive)")
	cmd.Flags().StringVar(&profile, "credentials-profile", "", "for profile")
	cmd.Flags().BoolVar(&forcePath, "force-path-style", false, "path-style addressing")
	cmd.Flags().BoolVar(&insecure, "insecure-skip-verify", false, "skip TLS verify")
	return cmd
}

func (g *Runtime) runBucketEdit(c *cobra.Command, from string, f *bucketAddFl) error {
	const name = "bucket edit"
	cfg, path, err := config.LoadOrEmpty(g.ConfigPath)
	if err != nil {
		return exitConfig(name, err)
	}
	if path == "" {
		path, _ = config.DefaultConfigPath()
	}
	if len(cfg.Buckets) == 0 {
		return exitConfigMsg(name, "no buckets in config")
	}

	var oldAlias string
	if g.NonInter {
		if from == "" {
			return exitUsage(name, "--from-alias is required with --non-interactive")
		}
		oldAlias = from
	} else {
		aliases := make([]string, 0, len(cfg.Buckets))
		for a := range cfg.Buckets {
			aliases = append(aliases, a)
		}
		sort.Strings(aliases)
		if from != "" {
			oldAlias = from
		} else {
			if err := askOne(&survey.Select{Message: "Bucket to edit", Options: aliases}, &oldAlias, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
		}
	}

	cur, ok := cfg.Buckets[oldAlias]
	if !ok {
		return exitConfigMsg(name, "unknown bucket alias %q", oldAlias)
	}

	if g.NonInter {
		if f.Alias == "" {
			f.Alias = oldAlias
		}
		if f.Bucket == "" {
			f.Bucket = cur.Bucket
		}
		if f.Endpoint == "" {
			f.Endpoint = cur.Endpoint
		}
		if f.Region == "" {
			f.Region = cur.Region
		}
		if f.CredentialType == "" {
			f.CredentialType = cur.CredentialType
		}
		if f.Prefix == "" {
			f.Prefix = cur.Prefix
		}
		if f.AccessKey == "" {
			f.AccessKey = cur.AccessKeyID
		}
		if f.Profile == "" {
			f.Profile = cur.CredentialsProfile
		}
		if c != nil {
			if !c.Flags().Changed("force-path-style") {
				f.ForcePath = cur.ForcePathStyle
			}
			if !c.Flags().Changed("insecure-skip-verify") {
				f.Insecure = cur.InsecureSkipVerify
			}
		}
	} else {
		*f = bucketAddFl{
			Alias:          curFirst(f.Alias, oldAlias),
			Bucket:         curFirst(f.Bucket, cur.Bucket),
			Endpoint:       curFirst(f.Endpoint, cur.Endpoint),
			Region:         curFirst(f.Region, cur.Region),
			Prefix:         curFirst(f.Prefix, cur.Prefix),
			CredentialType: curFirst(f.CredentialType, cur.CredentialType),
			AccessKey:      curFirst(f.AccessKey, cur.AccessKeyID),
			Secret:         f.Secret,
			Profile:        curFirst(f.Profile, cur.CredentialsProfile),
			ForcePath:      cur.ForcePathStyle,
			Insecure:       cur.InsecureSkipVerify,
			NoTest:         true,
		}
		if err := askOne(&survey.Input{Message: "Alias", Default: f.Alias}, &f.Alias, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		if err := askOne(&survey.Input{Message: "S3 bucket name", Default: f.Bucket}, &f.Bucket, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		if err := askOne(&survey.Input{Message: "Endpoint URL", Default: f.Endpoint}, &f.Endpoint, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		if err := askOne(&survey.Input{Message: "Region", Default: f.Region}, &f.Region, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		if err := askOne(&survey.Input{Message: "Key prefix (optional)", Default: f.Prefix}, &f.Prefix); err != nil {
			return err
		}
		if err := askOne(&survey.Select{Message: "Credential type", Options: []string{"static", "profile", "iam"}, Default: f.CredentialType}, &f.CredentialType); err != nil {
			return err
		}
		switch f.CredentialType {
		case "static":
			if err := askOne(&survey.Input{Message: "Access key ID", Default: f.AccessKey}, &f.AccessKey, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
			if err := askOne(&survey.Password{Message: "Secret access key (empty to keep unchanged)"}, &f.Secret); err != nil {
				return err
			}
		case "profile":
			if err := askOne(&survey.Input{Message: "Profile name", Default: f.Profile}, &f.Profile, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
		}
		if err := askOne(&survey.Confirm{Message: "Use path-style addressing?", Default: f.ForcePath}, &f.ForcePath); err != nil {
			return err
		}
		if err := askOne(&survey.Confirm{Message: "Insecure TLS (skip verify)?", Default: f.Insecure}, &f.Insecure); err != nil {
			return err
		}
	}
	if f.CredentialType == "static" && f.Secret == "" {
		f.Secret = cur.SecretAccessKey
	}
	if err := config.ValidateAlias(f.Alias); err != nil {
		return exitUsageErr(name, err)
	}
	if f.Alias != oldAlias {
		if _, taken := cfg.Buckets[f.Alias]; taken {
			return exitUsage(name, "alias %q already exists", f.Alias)
		}
	}
	nb := config.ClearUnusedCredFields(config.Bucket{
		Bucket:             f.Bucket,
		Endpoint:           f.Endpoint,
		Region:             f.Region,
		Prefix:             f.Prefix,
		ForcePathStyle:     f.ForcePath,
		InsecureSkipVerify: f.Insecure,
		CredentialType:     f.CredentialType,
		AccessKeyID:        f.AccessKey,
		SecretAccessKey:    f.Secret,
		CredentialsProfile: f.Profile,
	})
	if err := config.ValidateBucketCreds(nb); err != nil {
		return exitUsageErr(name, err)
	}
	if oldAlias != f.Alias {
		delete(cfg.Buckets, oldAlias)
	}
	cfg.Buckets[f.Alias] = nb
	cfg.Version = config.SupportedVersion
	if w := config.Save(path, cfg); w != nil {
		return exitConfig(name, w)
	}
	if f.Alias != oldAlias {
		_, _ = fmt.Fprintf(os.Stderr, "alias updated: %q -> %q (update scripts and env var names that reference the alias)\n", oldAlias, f.Alias)
	}
	return nil
}

func curFirst(flag, def string) string {
	if flag != "" {
		return flag
	}
	return def
}
