package sws

import (
	"context"
	"net/http"
	"testing"
)

// The unversioned /api alias is deprecated (Sunset 2027-09-22) and the platform's own
// version policy at GET /api/v1/version says first-party SDKs pin /api/v1. The older
// tests drive an httptest handler that ignores the path, so nothing caught this SDK
// calling the deprecated alias — these assert the path itself.

func recordPath(t *testing.T, baseURL func(string) string) string {
	t.Helper()
	var got string
	handler := func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Path
		mustJSON(t, w, 200, []Instance{})
	}
	srv, _ := newTestServer(t, handler)
	client := NewClient("ctk_test", WithBaseURL(baseURL(srv.URL)))
	if _, err := client.Compute.ListInstances(context.Background()); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestRequestsArePinnedToTheVersionedAPI(t *testing.T) {
	got := recordPath(t, func(base string) string { return base })
	if want := "/api/v1/compute/servers"; got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
}

// GET /api/v1/version publishes base_url as "https://savannaa.com/api/v1", so people
// paste exactly that into $SWS_API_URL. It must not double up.
func TestBaseURLThatAlreadyHasAPrefixIsNotDoubled(t *testing.T) {
	cases := map[string]func(string) string{
		"bare":             func(b string) string { return b },
		"trailing slash":   func(b string) string { return b + "/" },
		"with /api":        func(b string) string { return b + "/api" },
		"with /api/v1":     func(b string) string { return b + "/api/v1" },
		"with /api/v1/":    func(b string) string { return b + "/api/v1/" },
		"trailing slashes": func(b string) string { return b + "///" },
	}
	for name, mangle := range cases {
		t.Run(name, func(t *testing.T) {
			got := recordPath(t, mangle)
			if want := "/api/v1/compute/servers"; got != want {
				t.Errorf("path = %q, want %q", got, want)
			}
		})
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	cases := map[string]string{
		"https://savannaa.com":            "https://savannaa.com",
		"https://savannaa.com/":           "https://savannaa.com",
		"https://savannaa.com/api":        "https://savannaa.com",
		"https://savannaa.com/api/v1":     "https://savannaa.com",
		"https://savannaa.com/api/v1/":    "https://savannaa.com",
		"https://dev.example:3080/api/v1": "https://dev.example:3080",
	}
	for in, want := range cases {
		if got := normalizeBaseURL(in); got != want {
			t.Errorf("normalizeBaseURL(%q) = %q, want %q", in, got, want)
		}
	}
}
