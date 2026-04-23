package store

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joncombe/tagbackup/internal/config"
	"github.com/joncombe/tagbackup/internal/objectkey"
	"github.com/joncombe/tagbackup/internal/tagexpr"
)

// S3 is the AWS-backed [ObjectStore].
type S3 struct {
	Bucket   string
	Prefix   string
	Client   *s3.Client
	LogDebug *slog.Logger
}

// SetDebugLog sets the logger used for DEBUG output when listing (non-conforming keys).
func (s *S3) SetDebugLog(l *slog.Logger) { s.LogDebug = l }

// NewS3FromAlias loads config and the AWS S3 client. If debug is non-nil, the
// credential resolution branch and SDK retry attempts are logged at DEBUG.
func NewS3FromAlias(ctx context.Context, c *config.Cfg, alias string, debug *slog.Logger) (*S3, error) {
	if c == nil {
		return nil, fmt.Errorf("config is nil")
	}
	b, err := c.GetBucket(alias)
	if err != nil {
		return nil, err
	}
	client, err := NewS3Client(ctx, alias, b, debug)
	if err != nil {
		return nil, err
	}
	return &S3{
		Bucket:   b.Bucket,
		Prefix:   b.Prefix,
		Client:   client,
		LogDebug: debug,
	}, nil
}

// NewObjectStoreFromAlias is the [ObjectStore] entry point (same as
// [NewS3FromAlias] at runtime; useful for tests and stubs). If debug is
// non-nil, credential resolution and SDK retries are logged at DEBUG.
func NewObjectStoreFromAlias(ctx context.Context, c *config.Cfg, alias string, debug *slog.Logger) (ObjectStore, error) {
	return NewS3FromAlias(ctx, c, alias, debug)
}

// normalizedPrefix has trailing / if non-empty and no path weirdness.
func normalizedPrefix(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if !strings.HasSuffix(p, "/") {
		return p + "/"
	}
	return p
}

// relKey returns the part after the bucket key prefix, or an error if key does not start with prefix.
func (s *S3) relKey(full string) (string, error) {
	np := normalizedPrefix(s.Prefix)
	if np == "" {
		return full, nil
	}
	if !strings.HasPrefix(full, np) {
		return "", fmt.Errorf("key %q does not have expected prefix", full)
	}
	return strings.TrimPrefix(full, np), nil
}

// ListObjectsAll lists under prefix, parses tagbackup keys, and filters by ev (if non-nil) tag expression.
func (s *S3) ListObjectsAll(ctx context.Context, ev func(map[string]struct{}) bool) ([]Object, error) {
	np := normalizedPrefix(s.Prefix)
	pag := s3.NewListObjectsV2Paginator(s.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(np),
	})
	var out []Object
	for pag.HasMorePages() {
		page, err := pag.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, o := range page.Contents {
			if o.Key == nil {
				continue
			}
			rel, err := s.relKey(*o.Key)
			if err != nil {
				continue
			}
			p, err := objectkey.Parse(rel)
			if err != nil {
				if s.LogDebug != nil {
					s.LogDebug.Debug("skip non-conforming key", "key", *o.Key)
				}
				continue
			}
			if ev != nil {
				if !ev(p.TagSet()) {
					continue
				}
			}
			sz := int64(0)
			if o.Size != nil {
				sz = *o.Size
			}
			out = append(out, Object{Key: *o.Key, Parsed: p, Size: sz})
		}
	}
	return out, nil
}

// ParseTagExpr is a small helper to compile a tag expression.
func ParseTagExpr(s string) (func(map[string]struct{}) bool, error) {
	return tagexpr.Parse(s)
}
