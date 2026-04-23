package store

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joncombe/tagbackup/internal/objectkey"
)

// Upload streams an object using the S3 upload manager (multipart when needed).
// LeavePartsOnError is false so a failure or context cancellation aborts the
// multipart upload and does not leave billable part garbage on the bucket.
// size is currently unused; the SDK picks multipart parameters itself. It is
// kept in the ObjectStore signature so callers can pass known content lengths
// if/when we tune PartSize.
func (s *S3) Upload(ctx context.Context, key string, r io.Reader, size int64) error {
	_ = size
	uploader := manager.NewUploader(s.Client, func(u *manager.Uploader) {
		u.LeavePartsOnError = false
	})
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	return err
}

// GetObjectReader returns the object body and content length if present.
func (s *S3) GetObjectReader(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	out, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, 0, err
	}
	var n int64
	if out.ContentLength != nil {
		n = *out.ContentLength
	}
	return out.Body, n, nil
}

// DeleteObject removes one key.
func (s *S3) DeleteObject(ctx context.Context, key string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	return err
}

// BuildPushKey assembles a full S3 key for a new upload.
func (s *S3) BuildPushKey(original string, tagList []string) (string, error) {
	ts := time.Now().UTC().UnixMilli()
	np := normalizedPrefix(s.Prefix)
	rel, err := objectkey.BuildKey("", original, tagList, ts)
	if err != nil {
		return "", err
	}
	if np != "" {
		return np + rel, nil
	}
	return rel, nil
}
