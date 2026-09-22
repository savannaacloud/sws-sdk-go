package sws

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withCleanEnv(t *testing.T) string {
	t.Helper()
	for _, v := range []string{"SWS_API_KEY", "SWS_REGION", "SWS_BASE_URL", "SWS_API_URL", "SWS_PROFILE"} {
		t.Setenv(v, "")
		os.Unsetenv(v)
	}
	dir := t.TempDir()
	t.Setenv("SWS_CREDENTIALS_FILE", filepath.Join(dir, "credentials"))
	return dir
}

func TestCredentialResolutionOrder(t *testing.T) {
	dir := withCleanEnv(t)
	if k, _, _, src := resolveCredentials("ctk_explicit"); k != "ctk_explicit" || src != "argument" {
		t.Fatalf("argument should win: %s %s", k, src)
	}
	t.Setenv("SWS_API_KEY", "ctk_env")
	if k, _, _, src := resolveCredentials(""); k != "ctk_env" || src != "SWS_API_KEY" {
		t.Fatalf("env should be used: %s %s", k, src)
	}
	os.Unsetenv("SWS_API_KEY")
	os.WriteFile(filepath.Join(dir, "credentials"), []byte("[default]\napi_key = ctk_file\nregion = ng-abuja-1\n\n[ci]\napi_key = ctk_ci\napi_url = https://savannaa.com\n"), 0o600)
	k, region, _, src := resolveCredentials("")
	if k != "ctk_file" || region != "ng-abuja-1" || !strings.Contains(src, "credentials") {
		t.Fatalf("file: %s %s %s", k, region, src)
	}
	t.Setenv("SWS_PROFILE", "ci")
	if k, _, base, _ := resolveCredentials(""); k != "ctk_ci" || base != "https://savannaa.com" {
		t.Fatalf("profile: %s %s", k, base)
	}
	os.Unsetenv("SWS_PROFILE")
	os.Remove(filepath.Join(dir, "credentials"))
	os.WriteFile(filepath.Join(dir, "token"), []byte("eyJhbGciOi.session.token\n"), 0o600)
	if k, _, _, src := resolveCredentials(""); k != "eyJhbGciOi.session.token" || !strings.HasSuffix(src, "token") {
		t.Fatalf("cli token: %s %s", k, src)
	}
}

func TestMissingCredentialsNamesTheEnvVar(t *testing.T) {
	withCleanEnv(t)
	c := NewClient("")
	err := c.do(nil, "GET", "/api/v1/compute/servers", nil, nil) //nolint:staticcheck // nil ctx is fine: it never reaches the transport
	if err == nil || !strings.Contains(err.Error(), "SWS_API_KEY") {
		t.Fatalf("error should name the env var, got %v", err)
	}
	if c.CredentialSource() != "none" {
		t.Fatalf("source should be none, got %s", c.CredentialSource())
	}
}
