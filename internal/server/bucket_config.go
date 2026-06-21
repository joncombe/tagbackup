package server

import (
	"os"

	"github.com/joncombe/tagbackup/internal/config"
)

// bucketConfigDTO is the sanitized JSON shape returned for bucket configuration.
type bucketConfigDTO struct {
	Alias              string  `json:"alias"`
	Bucket             string  `json:"bucket"`
	Endpoint           string  `json:"endpoint"`
	Region             string  `json:"region"`
	Prefix             string  `json:"prefix,omitempty"`
	ForcePathStyle     bool    `json:"force_path_style"`
	InsecureSkipVerify bool    `json:"insecure_skip_verify"`
	CredentialType     string  `json:"credential_type"`
	CredentialSource   string  `json:"credential_source"`
	AccessKeyID        *string `json:"access_key_id,omitempty"`
	SecretAccessKey    *string `json:"secret_access_key,omitempty"`
	CredentialsProfile string  `json:"credentials_profile,omitempty"`
}

func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 4 {
		return "****"
	}
	return "****" + s[len(s)-4:]
}

func credentialSource(alias string, b config.Bucket) string {
	frag := config.EnvKeyFragment(alias)
	envAccess := "TAGBACKUP_BUCKET_" + frag + "_ACCESS_KEY_ID"
	envSecret := "TAGBACKUP_BUCKET_" + frag + "_SECRET_ACCESS_KEY"
	if os.Getenv(envAccess) != "" && os.Getenv(envSecret) != "" {
		return "env"
	}
	switch b.CredentialType {
	case "static":
		return "static"
	case "profile":
		return "profile"
	case "iam":
		return "iam"
	default:
		return "iam"
	}
}

func bucketConfigFrom(alias string, b config.Bucket) bucketConfigDTO {
	src := credentialSource(alias, b)
	dto := bucketConfigDTO{
		Alias:              alias,
		Bucket:             b.Bucket,
		Endpoint:           b.Endpoint,
		Region:             b.Region,
		Prefix:             b.Prefix,
		ForcePathStyle:     b.ForcePathStyle,
		InsecureSkipVerify: b.InsecureSkipVerify,
		CredentialType:     b.CredentialType,
		CredentialSource:   src,
		CredentialsProfile: b.CredentialsProfile,
	}
	if src != "env" {
		if b.AccessKeyID != "" {
			masked := maskSecret(b.AccessKeyID)
			dto.AccessKeyID = &masked
		}
		if b.SecretAccessKey != "" {
			masked := maskSecret(b.SecretAccessKey)
			dto.SecretAccessKey = &masked
		}
	}
	return dto
}
