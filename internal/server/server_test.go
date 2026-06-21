package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joncombe/tagbackup/internal/config"
)

func newTestServer(configPath string) http.Handler {
	return New(Options{ConfigPath: configPath}).routes()
}

func writeConfig(t *testing.T, aliases ...string) string {
	t.Helper()
	cfg := &config.Cfg{Version: config.SupportedVersion, Buckets: map[string]config.Bucket{}}
	for _, a := range aliases {
		cfg.Buckets[a] = config.Bucket{
			Bucket:          "b-" + a,
			Endpoint:        "https://example.com",
			Region:          "us-east-1",
			CredentialType:  "static",
			AccessKeyID:     "key",
			SecretAccessKey: "secret",
		}
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	return path
}

func TestVersionReturnsConfiguredVersion(t *testing.T) {
	srv := New(Options{Version: "0.0.6"}).routes()
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/version", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["version"] != "0.0.6" {
		t.Fatalf("version = %q, want 0.0.6", got["version"])
	}
}

func TestVersionDefaultsToDev(t *testing.T) {
	srv := newTestServer(filepath.Join(t.TempDir(), "none.yaml"))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/version", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["version"] != "dev" {
		t.Fatalf("version = %q, want dev", got["version"])
	}
}

func TestBucketsEmptyWhenConfigMissing(t *testing.T) {
	srv := newTestServer(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/buckets", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var got []string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v (body=%q)", err, rr.Body.String())
	}
	if len(got) != 0 {
		t.Fatalf("buckets = %v, want empty", got)
	}
}

func TestBucketsSortedAlphabetically(t *testing.T) {
	srv := newTestServer(writeConfig(t, "zebra", "alpha", "mango"))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/buckets", nil))

	var got []string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := []string{"alpha", "mango", "zebra"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("buckets = %v, want %v", got, want)
	}
}

func TestObjectsUnknownBucketReturns404(t *testing.T) {
	srv := newTestServer(writeConfig(t, "alpha"))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/buckets/nope/objects", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestBucketConfigUnknownBucketReturns404(t *testing.T) {
	srv := newTestServer(writeConfig(t, "alpha"))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/buckets/nope", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestBucketConfigInvalidAliasReturns400(t *testing.T) {
	srv := newTestServer(writeConfig(t, "alpha"))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/buckets/bad%20alias", nil))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestBucketConfigMasksSecrets(t *testing.T) {
	cfg := &config.Cfg{Version: config.SupportedVersion, Buckets: map[string]config.Bucket{}}
	cfg.Buckets["alpha"] = config.Bucket{
		Bucket:          "b-alpha",
		Endpoint:        "https://example.com",
		Region:          "us-east-1",
		Prefix:          "nightly/",
		CredentialType:  "static",
		AccessKeyID:     "AKIAEXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	srv := newTestServer(path)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/buckets/alpha", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var got bucketConfigDTO
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Alias != "alpha" || got.Bucket != "b-alpha" || got.Prefix != "nightly/" {
		t.Fatalf("unexpected fields: %+v", got)
	}
	if got.AccessKeyID == nil || *got.AccessKeyID != "****MPLE" {
		t.Fatalf("access_key_id = %v, want ****MPLE", got.AccessKeyID)
	}
	if got.SecretAccessKey == nil || *got.SecretAccessKey != "****EKEY" {
		t.Fatalf("secret_access_key = %v, want ****EKEY", got.SecretAccessKey)
	}
	if got.CredentialSource != "static" {
		t.Fatalf("credential_source = %q, want static", got.CredentialSource)
	}
}

func TestBucketConfigEnvCredentialSource(t *testing.T) {
	path := writeConfig(t, "alpha")
	t.Setenv("TAGBACKUP_BUCKET_ALPHA_ACCESS_KEY_ID", "envkey")
	t.Setenv("TAGBACKUP_BUCKET_ALPHA_SECRET_ACCESS_KEY", "envsecret")
	srv := newTestServer(path)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/buckets/alpha", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var got bucketConfigDTO
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.CredentialSource != "env" {
		t.Fatalf("credential_source = %q, want env", got.CredentialSource)
	}
	if got.AccessKeyID != nil || got.SecretAccessKey != nil {
		t.Fatalf("expected no inline credentials, got access_key_id=%v secret_access_key=%v", got.AccessKeyID, got.SecretAccessKey)
	}
	body := rr.Body.String()
	if strings.Contains(body, "envkey") || strings.Contains(body, "envsecret") || strings.Contains(body, "secret") {
		// "secret_access_key" key name is fine; raw values must not appear
		if strings.Contains(body, "envkey") || strings.Contains(body, "envsecret") {
			t.Fatalf("response must not contain env credential values: %s", body)
		}
	}
}

func TestObjectsInvalidAliasReturns400(t *testing.T) {
	srv := newTestServer(writeConfig(t, "alpha"))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/buckets/bad%20alias/objects", nil))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestStaticServesIndex(t *testing.T) {
	srv := newTestServer(filepath.Join(t.TempDir(), "none.yaml"))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "<html") {
		t.Fatalf("index body does not look like HTML: %q", rr.Body.String())
	}
}

func TestStaticSPAFallback(t *testing.T) {
	srv := newTestServer(filepath.Join(t.TempDir(), "none.yaml"))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/some/client/route", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (SPA fallback)", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "<html") {
		t.Fatalf("fallback body is not index.html: %q", rr.Body.String())
	}
}
