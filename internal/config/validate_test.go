package config

import "testing"

func valid() *Cfg {
	return &Cfg{
		Version: SupportedVersion,
		Buckets: map[string]Bucket{
			"ok": {
				Bucket:          "my-bucket",
				Endpoint:        "https://s3.example.com",
				Region:          "us-east-1",
				CredentialType:  "static",
				AccessKeyID:     "AKIA",
				SecretAccessKey: "secret",
			},
		},
	}
}

func TestValidateConfig_OK(t *testing.T) {
	if err := ValidateConfig(valid()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidateConfig_RejectsFutureVersion(t *testing.T) {
	c := valid()
	c.Version = SupportedVersion + 1
	if err := ValidateConfig(c); err == nil {
		t.Fatal("expected error for newer schema version")
	}
}

func TestValidateConfig_RejectsBadAlias(t *testing.T) {
	c := valid()
	c.Buckets["bad-alias"] = c.Buckets["ok"]
	if err := ValidateConfig(c); err == nil {
		t.Fatal("expected error for alias containing a hyphen")
	}
}

func TestValidateConfig_RequiresStaticKeys(t *testing.T) {
	c := valid()
	b := c.Buckets["ok"]
	b.AccessKeyID = ""
	c.Buckets["ok"] = b
	if err := ValidateConfig(c); err == nil {
		t.Fatal("expected error for static creds missing access_key_id")
	}
}

func TestValidateConfig_RequiresProfileName(t *testing.T) {
	c := valid()
	b := c.Buckets["ok"]
	b.CredentialType = "profile"
	b.AccessKeyID = ""
	b.SecretAccessKey = ""
	c.Buckets["ok"] = b
	if err := ValidateConfig(c); err == nil {
		t.Fatal("expected error for profile creds missing credentials_profile")
	}
}

func TestValidateConfig_RejectsUnknownCredType(t *testing.T) {
	c := valid()
	b := c.Buckets["ok"]
	b.CredentialType = "other"
	c.Buckets["ok"] = b
	if err := ValidateConfig(c); err == nil {
		t.Fatal("expected error for unknown credential type")
	}
}

func TestClearUnusedCredFields(t *testing.T) {
	in := Bucket{
		CredentialType:     "static",
		AccessKeyID:        "A",
		SecretAccessKey:    "S",
		CredentialsProfile: "stale",
	}
	out := ClearUnusedCredFields(in)
	if out.CredentialsProfile != "" {
		t.Errorf("expected profile cleared for static, got %q", out.CredentialsProfile)
	}

	in = Bucket{
		CredentialType:     "profile",
		AccessKeyID:        "stale-key",
		SecretAccessKey:    "stale-secret",
		CredentialsProfile: "default",
	}
	out = ClearUnusedCredFields(in)
	if out.AccessKeyID != "" || out.SecretAccessKey != "" {
		t.Errorf("expected static fields cleared for profile, got %q / %q", out.AccessKeyID, out.SecretAccessKey)
	}

	in = Bucket{
		CredentialType:     "iam",
		AccessKeyID:        "stale",
		SecretAccessKey:    "stale",
		CredentialsProfile: "stale",
	}
	out = ClearUnusedCredFields(in)
	if out.AccessKeyID != "" || out.SecretAccessKey != "" || out.CredentialsProfile != "" {
		t.Error("expected all creds cleared for iam")
	}
}

func TestValidateAlias(t *testing.T) {
	good := []string{"a", "A", "0", "my_bucket", "A_B_C_123"}
	for _, s := range good {
		if err := ValidateAlias(s); err != nil {
			t.Errorf("%q: unexpected err %v", s, err)
		}
	}
	bad := []string{"", "a-b", "a b", "a.b", "a/b"}
	for _, s := range bad {
		if err := ValidateAlias(s); err == nil {
			t.Errorf("%q: expected err", s)
		}
	}
}
