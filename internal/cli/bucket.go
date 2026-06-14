package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joncombe/tagbackup/internal/config"
	"github.com/joncombe/tagbackup/internal/store"
	"github.com/spf13/cobra"
)

func (g *Runtime) cmdBucket() *cobra.Command {
	b := &cobra.Command{
		Use:   "bucket",
		Short: "Manage configured bucket entries",
	}
	b.AddCommand(g.cmdBucketAdd(), g.cmdBucketList(), g.cmdBucketVerify(), g.cmdBucketDelete(), g.cmdBucketEdit())
	return b
}

func (g *Runtime) cmdBucketList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List bucket aliases in the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.LoadOrEmpty(g.ConfigPath)
			if err != nil {
				return exitConfig("bucket list", err)
			}
			aliases := make([]string, 0, len(cfg.Buckets))
			for a := range cfg.Buckets {
				aliases = append(aliases, a)
			}
			sort.Strings(aliases)
			for _, a := range aliases {
				_, _ = fmt.Fprintln(os.Stdout, a)
			}
			return nil
		},
	}
}

func (g *Runtime) cmdBucketAdd() *cobra.Command {
	var (
		alias, bucketName, endpoint, region, prefix string
		credType, accessKey, secret, profile         string
		forcePath, insecure, noTest                 bool
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a bucket to the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.runBucketAdd(&bucketAddFl{
				Alias:              alias,
				Bucket:             bucketName,
				Endpoint:           endpoint,
				Region:             region,
				Prefix:             prefix,
				CredentialType:     credType,
				AccessKey:          accessKey,
				Secret:             secret,
				Profile:            profile,
				ForcePath:          forcePath,
				Insecure:           insecure,
				NoTest:             noTest,
			})
		},
	}
	cmd.Flags().StringVar(&alias, "alias", "", "bucket alias")
	cmd.Flags().StringVar(&bucketName, "bucket", "", "S3 bucket name")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "S3 endpoint URL")
	cmd.Flags().StringVar(&region, "region", "", "region")
	cmd.Flags().StringVar(&prefix, "prefix", "", "optional key prefix")
	cmd.Flags().StringVar(&credType, "credential-type", "", "static|profile|iam")
	cmd.Flags().StringVar(&accessKey, "access-key-id", "", "for static")
	cmd.Flags().StringVar(&secret, "secret-access-key", "", "for static")
	cmd.Flags().StringVar(&profile, "credentials-profile", "", "for profile")
	cmd.Flags().BoolVar(&forcePath, "force-path-style", false, "path-style addressing")
	cmd.Flags().BoolVar(&insecure, "insecure-skip-verify", false, "skip TLS verify")
	cmd.Flags().BoolVar(&noTest, "no-test", false, "skip post-add connection test")
	return cmd
}

type bucketAddFl struct {
	Alias, Bucket, Endpoint, Region, Prefix      string
	CredentialType, AccessKey, Secret, Profile   string
	ForcePath, Insecure, NoTest                    bool
}

func (g *Runtime) runBucketAdd(f *bucketAddFl) error {
	const name = "bucket add"
	if f.Alias == "" || f.Bucket == "" || f.Endpoint == "" || f.Region == "" || f.CredentialType == "" {
		if g.NonInter {
			return exitUsage(name, "--alias, --bucket, --endpoint, --region, and --credential-type are required in non-interactive mode")
		}
		if err := askOne(&survey.Input{Message: "Alias"}, &f.Alias, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		if err := askOne(&survey.Input{Message: "S3 bucket name"}, &f.Bucket, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		if err := askOne(&survey.Input{Message: "Endpoint URL"}, &f.Endpoint, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		if err := askOne(&survey.Input{Message: "Region"}, &f.Region, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		if err := askOne(&survey.Input{Message: "Key prefix (optional)"}, &f.Prefix); err != nil {
			return err
		}
		if err := askOne(&survey.Confirm{Message: "Use path-style addressing?", Default: f.ForcePath}, &f.ForcePath); err != nil {
			return err
		}
		if err := askOne(&survey.Confirm{Message: "Insecure TLS (skip verify)?", Default: f.Insecure}, &f.Insecure); err != nil {
			return err
		}
		if err := askOne(&survey.Select{Message: "Credential type", Options: []string{"static", "profile", "iam"}}, &f.CredentialType); err != nil {
			return err
		}
		switch f.CredentialType {
		case "static":
			if err := askOne(&survey.Input{Message: "Access key ID"}, &f.AccessKey, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
			if err := askOne(&survey.Password{Message: "Secret access key"}, &f.Secret, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
		case "profile":
			if err := askOne(&survey.Input{Message: "Profile name"}, &f.Profile, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
		}
	}
	if err := config.ValidateAlias(f.Alias); err != nil {
		return exitUsageErr(name, err)
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
	cfg, path, err := config.LoadOrEmpty(g.ConfigPath)
	if err != nil {
		return exitConfig(name, err)
	}
	if path == "" {
		path, _ = config.DefaultConfigPath()
	}
	if cfg == nil {
		cfg = &config.Cfg{Version: config.SupportedVersion, Buckets: map[string]config.Bucket{}}
	}
	if cfg.Buckets == nil {
		cfg.Buckets = map[string]config.Bucket{}
	}
	if _, ok := cfg.Buckets[f.Alias]; ok {
		return exitUsage(name, "alias %q already exists", f.Alias)
	}
	cfg.Version = config.SupportedVersion
	cfg.Buckets[f.Alias] = nb
	if w := config.Save(path, cfg); w != nil {
		return exitConfig(name, w)
	}
	if !f.NoTest {
		if err := g.runBucketVerifyProbe(f.Alias, cfg); err != nil {
			return err
		}
	}
	return nil
}

func (g *Runtime) cmdBucketVerify() *cobra.Command {
	const name = "bucket verify"
	var flagBucket string
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify S3 permissions for a configured bucket",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.LoadOrEmpty(g.ConfigPath)
			if err != nil {
				return exitConfig(name, err)
			}
			aliases := make([]string, 0, len(cfg.Buckets))
			for a := range cfg.Buckets {
				aliases = append(aliases, a)
			}
			if len(aliases) == 0 {
				return exitConfigMsg(name, "no buckets in config")
			}
			sort.Strings(aliases)

			var alias string
			if g.NonInter {
				if len(aliases) == 1 {
					alias = aliases[0]
					if flagBucket != "" && flagBucket != alias {
						return exitUsage(name, "unknown bucket alias %q (only %q is configured)", flagBucket, alias)
					}
				} else {
					if flagBucket == "" {
						return exitUsage(name, "with --non-interactive and multiple entries, set --bucket=ALIAS")
					}
					if _, ok := cfg.Buckets[flagBucket]; !ok {
						return exitConfigMsg(name, "unknown bucket alias %q", flagBucket)
					}
					alias = flagBucket
				}
			} else {
				if flagBucket != "" {
					if _, ok := cfg.Buckets[flagBucket]; !ok {
						return exitConfigMsg(name, "unknown bucket alias %q", flagBucket)
					}
					alias = flagBucket
				} else {
					if err := askOne(&survey.Select{
						Message:  "Bucket alias to verify",
						Options:  aliases,
						PageSize: 20,
					}, &alias, survey.WithValidator(survey.Required)); err != nil {
						return err
					}
				}
			}
			return g.runBucketVerifyProbe(alias, cfg)
		},
	}
	cmd.Flags().StringVar(&flagBucket, "bucket", "", "bucket alias to verify (with multiple entries, use with --non-interactive; when interactive, optional to skip the picker)")
	return cmd
}

func headBucketHint(err error) string {
	s := err.Error()
	base := "check the bucket name, region, endpoint, and credentials"
	if strings.Contains(s, "404") || strings.Contains(s, "NotFound") {
		return base + "; the bucket may not exist at this endpoint, or the credentials may lack permission to see it. For MinIO and some other providers, try force_path_style: true in config."
	}
	if strings.Contains(s, "403") || strings.Contains(s, "Forbidden") {
		return base + "; the credentials exist but are not authorised. Check the IAM policy for s3:ListBucket on this bucket."
	}
	return base
}

func (g *Runtime) runBucketVerifyProbe(alias string, cfg *config.Cfg) error {
	const name = "bucket verify"
	entr, err := cfg.GetBucket(alias)
	if err != nil {
		return exitConfig(name, err)
	}
	client, err := store.NewS3Client(g.Ctx, alias, entr, g.debugLogger())
	if err != nil {
		return exitConfig(name, err)
	}
	np := entr.Prefix
	if np != "" && !strings.HasSuffix(np, "/") {
		np += "/"
	}
	rand8 := make([]byte, 4)
	_, _ = rand.Read(rand8)
	probe := np + ".tagbackup-permission-check-" + hex.EncodeToString(rand8)

	failed := false
	printOK := func(label string) {
		_, _ = fmt.Fprintln(os.Stderr, "  ", label, g.green("✓"))
	}
	printFail := func(label string, e error, hint string) {
		failed = true
		_, _ = fmt.Fprintln(os.Stderr, "  ", label, g.red("✗"), e)
		if hint != "" {
			_, _ = fmt.Fprintln(os.Stderr, "   ", g.red("Hint:"), hint)
		}
	}

	_, hErr := client.HeadBucket(g.Ctx, &s3.HeadBucketInput{Bucket: aws.String(entr.Bucket)})
	if hErr != nil {
		printFail("Connecting to bucket", hErr, headBucketHint(hErr))
		return exitS3(name, hErr)
	}
	printOK("Connecting to bucket")

	_, lErr := client.ListObjectsV2(g.Ctx, &s3.ListObjectsV2Input{Bucket: aws.String(entr.Bucket), Prefix: aws.String(np), MaxKeys: aws.Int32(1)})
	if lErr != nil {
		printFail("Listing files in bucket", lErr, "check s3:ListBucket on the bucket (and your prefix) for this role")
	} else {
		printOK("Listing files in bucket")
	}

	_, wErr := client.PutObject(g.Ctx, &s3.PutObjectInput{Bucket: aws.String(entr.Bucket), Key: aws.String(probe), Body: strings.NewReader("tagbackup verify")})
	if wErr != nil {
		printFail("Writing file to bucket", wErr, "check s3:PutObject for this key prefix")
	} else {
		printOK("Writing file to bucket")
	}

	_, rErr := client.GetObject(g.Ctx, &s3.GetObjectInput{Bucket: aws.String(entr.Bucket), Key: aws.String(probe)})
	if rErr != nil {
		printFail("Reading file from bucket", rErr, "check s3:GetObject for objects under this key prefix")
	} else {
		printOK("Reading file from bucket")
	}

	_, dErr := client.DeleteObject(g.Ctx, &s3.DeleteObjectInput{Bucket: aws.String(entr.Bucket), Key: aws.String(probe)})
	if dErr != nil {
		failed = true
		_, _ = fmt.Fprintln(os.Stderr, "  Deleting file from bucket", g.red("✗"), dErr)
		_, _ = fmt.Fprintln(os.Stderr, "  warning: could not remove probe; delete manually:", probe)
	} else {
		printOK("Deleting file from bucket")
	}

	if failed {
		return exitS3Msg(name, "one or more steps failed; see above")
	}
	return nil
}

func (g *Runtime) cmdBucketDelete() *cobra.Command {
	const name = "bucket delete"
	return &cobra.Command{
		Use:   "delete",
		Short: "Remove a bucket from the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if g.NonInter {
				return exitUsage(name, "this command is interactive; omit --non-interactive")
			}
			if !StdinIsTTY() {
				return exitUsage(name, "confirmation requires a TTY")
			}
			cfg, path, err := config.LoadOrEmpty(g.ConfigPath)
			if err != nil {
				return exitConfig(name, err)
			}
			if len(cfg.Buckets) == 0 {
				return exitConfigMsg(name, "no buckets in config")
			}
			aliases := make([]string, 0, len(cfg.Buckets))
			for a := range cfg.Buckets {
				aliases = append(aliases, a)
			}
			sort.Strings(aliases)
			var which string
			if err := askOneErr(name, &survey.Select{Message: "Remove alias", Options: aliases}, &which, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
			var sure bool
			if err := askOneErr(name, &survey.Confirm{
				Message: fmt.Sprintf("Remove alias %q from the config? (does not delete the remote bucket)", which),
				Default: false,
			}, &sure); err != nil {
				return err
			}
			if !sure {
				return nil
			}
			delete(cfg.Buckets, which)
			if e := config.Save(path, cfg); e != nil {
				return exitConfig(name, e)
			}
			return nil
		},
	}
}
