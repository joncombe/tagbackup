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
