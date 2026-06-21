package store

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joncombe/tagbackup/internal/config"
)

// NewS3Client returns an S3 client configured for the given bucket entry.
// If debug is non-nil, the credential-resolution branch and SDK retry attempts
// are logged at DEBUG level.
func NewS3Client(ctx context.Context, alias string, b config.Bucket, debug *slog.Logger) (*s3.Client, error) {
	cfg, err := loadBaseAWSConfig(ctx, alias, b, debug)
	if err != nil {
		return nil, err
	}
	ep := NormalizeAPIEndpoint(b.Endpoint, b.Bucket)
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(ep)
		o.UsePathStyle = b.ForcePathStyle
		// Multipart uploads (used by push) expose composite checksums the SDK
		// cannot validate on GetObject; suppress the harmless WARN line.
		o.DisableLogOutputChecksumValidationSkipped = true
		if debug != nil {
			o.Retryer = newLoggingRetryer(debug)
		}
	})
	return client, nil
}

func loadBaseAWSConfig(ctx context.Context, alias string, b config.Bucket, debug *slog.Logger) (aws.Config, error) {
	frag := config.EnvKeyFragment(alias)
	envAccess := "TAGBACKUP_BUCKET_" + frag + "_ACCESS_KEY_ID"
	envSecret := "TAGBACKUP_BUCKET_" + frag + "_SECRET_ACCESS_KEY"
	akey := os.Getenv(envAccess)
	asec := os.Getenv(envSecret)

	var opts []func(*awsconfig.LoadOptions) error
	opts = append(opts, awsconfig.WithRegion(b.Region))

	switch {
	case akey != "" && asec != "":
		logResolve(debug, "credentials: resolved via env var "+envAccess)
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(akey, asec, ""),
		))
	case b.CredentialType == "static":
		if b.AccessKeyID == "" || b.SecretAccessKey == "" {
			return aws.Config{}, fmt.Errorf("static credentials: missing access_key_id/secret; set %s or edit config", envAccess)
		}
		logResolve(debug, "credentials: resolved via inline static secret in config.yaml")
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(b.AccessKeyID, b.SecretAccessKey, ""),
		))
	case b.CredentialType == "profile":
		logResolve(debug, "credentials: resolved via shared profile "+b.CredentialsProfile)
		opts = append(opts, awsconfig.WithSharedConfigProfile(b.CredentialsProfile))
	case b.CredentialType == "iam":
		logResolve(debug, "credentials: resolved via AWS SDK default chain")
	default:
		logResolve(debug, "credentials: resolved via AWS SDK default chain (fallback)")
	}

	if b.InsecureSkipVerify {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		if tr == nil {
			tr = &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
			}
		}
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12} //nolint:gosec
		opts = append(opts, awsconfig.WithHTTPClient(&http.Client{Transport: tr}))
	}

	return awsconfig.LoadDefaultConfig(ctx, opts...)
}

func logResolve(debug *slog.Logger, msg string) {
	if debug != nil {
		debug.Debug(msg)
	}
}
