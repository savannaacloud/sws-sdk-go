package sws

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

// Live checks against a real console. Skipped unless SWS_LIVE_BASE is set:
//
//	SWS_LIVE_BASE=https://savannaa.com go test -run Live -v
//
// Read-only; creates nothing. The httptest-based tests can only prove the SDK builds
// the URL it intends to — only a real gateway proves that URL exists. A wrong path
// (a doubled prefix, or the alias after it is withdrawn) returns 404, which surfaces
// as *NotFoundError rather than *AuthenticationError.

const liveBogusKey = "ctk_live_probe_never_issued_000000000000"

func TestLiveVersionedPathExistsAndKeyIsRefused(t *testing.T) {
	base := os.Getenv("SWS_LIVE_BASE")
	if base == "" {
		t.Skip("set SWS_LIVE_BASE to run live checks")
	}
	base = strings.TrimRight(base, "/")

	for _, suffix := range []string{"", "/", "/api", "/api/v1", "/api/v1/"} {
		t.Run("base"+strings.ReplaceAll(suffix, "/", "_"), func(t *testing.T) {
			client := NewClient(liveBogusKey, WithBaseURL(base+suffix))
			_, err := client.Compute.ListInstances(context.Background())
			if err == nil {
				t.Fatal("a key that was never issued was accepted")
			}
			var notFound *NotFoundError
			if errors.As(err, &notFound) {
				t.Fatalf("base %q did not reach /api/v1/compute/servers: %v", base+suffix, err)
			}
			var auth *AuthenticationError
			if !errors.As(err, &auth) {
				t.Fatalf("want *AuthenticationError, got %T: %v", err, err)
			}
		})
	}
}
