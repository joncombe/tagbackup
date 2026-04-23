package store

import "testing"

func TestNormalizeAPIEndpoint(t *testing.T) {
	const acc = "https://e8b9bfdc5feec62c7b11d79232ad0611.r2.cloudflarestorage.com"
	got := NormalizeAPIEndpoint(acc+"/test", "test")
	if got != acc {
		t.Fatalf("strip /bucket: got %q want %q", got, acc)
	}
	if NormalizeAPIEndpoint(acc, "test") != acc {
		t.Fatal("no path unchanged")
	}
	if NormalizeAPIEndpoint(acc+"/other", "test") != acc+"/other" {
		t.Fatal("non-matching path unchanged")
	}
}
