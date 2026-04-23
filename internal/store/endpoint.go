package store

import (
	"net/url"
	"strings"
)

// NormalizeAPIEndpoint returns an endpoint suitable for the AWS S3 client BaseEndpoint.
// If the URL path is exactly "/{bucket}" matching the configured bucket name, the path
// is stripped. Some providers (e.g. Cloudflare R2) and console UIs expose an account
// URL that includes the bucket in the path; the S3 API still expects the account root
// with the bucket name passed on each request.
func NormalizeAPIEndpoint(endpoint, bucket string) string {
	if endpoint == "" || bucket == "" {
		return endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	path := strings.Trim(u.Path, "/")
	if path == bucket {
		u.Path = ""
		out := u.String()
		return strings.TrimSuffix(out, "/")
	}
	return endpoint
}
