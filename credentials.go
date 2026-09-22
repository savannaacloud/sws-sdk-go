package sws

// Where credentials come from when the caller does not pass them.
//
// The first code anybody copies should not contain a secret, so NewClient("") is a
// complete quick start. Resolution order, the same one every modern SDK uses:
//
//  1. the apiKey argument
//  2. $SWS_API_KEY
//  3. the credentials file `sws auth login` writes — ~/.config/sws/credentials (INI,
//     profile-aware via $SWS_PROFILE) or the plain ~/.config/sws/token the CLI has
//     always written
//  4. nothing: the first request fails with an error that names the env var instead of
//     a bare 401.
//
// $SWS_CREDENTIALS_FILE overrides the path (used by the tests).

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// MissingCredentials is what a client without credentials reports.
const MissingCredentials = "Missing credentials: set SWS_API_KEY, run `sws auth login`, or pass an api key to NewClient"

func credentialsPath() string {
	if p := os.Getenv("SWS_CREDENTIALS_FILE"); p != "" {
		return p
	}
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "sws", "credentials")
}

// profileFromFile reads one INI profile (falling back to [default]).
func profileFromFile(profile string) map[string]string {
	path := credentialsPath()
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	sections := map[string]map[string]string{}
	current := "default"
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if sections[current] == nil {
			sections[current] = map[string]string{}
		}
		sections[current][strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	if s, ok := sections[profile]; ok {
		return s
	}
	return sections["default"]
}

// tokenFromCLI is the session token `sws auth login` saves. Handy on a developer
// machine; it expires, which is why an API key still wins.
func tokenFromCLI() string {
	path := credentialsPath()
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(filepath.Dir(path), "token"))
	if err != nil {
		return ""
	}
	t := strings.TrimSpace(string(b))
	if t == "" || strings.Contains(t, "\n") {
		return ""
	}
	return t
}

// resolveCredentials returns the key, region, base URL and where the key came from.
func resolveCredentials(apiKey string) (key, region, baseURL, source string) {
	region = os.Getenv("SWS_REGION")
	baseURL = firstNonEmpty(os.Getenv("SWS_API_URL"), os.Getenv("SWS_BASE_URL"))
	if apiKey != "" {
		return apiKey, region, baseURL, "argument"
	}
	if v := os.Getenv("SWS_API_KEY"); v != "" {
		return v, region, baseURL, "SWS_API_KEY"
	}
	profile := os.Getenv("SWS_PROFILE")
	if profile == "" {
		profile = "default"
	}
	if p := profileFromFile(profile); p != nil {
		if k := firstNonEmpty(p["api_key"], p["token"]); k != "" {
			return k, firstNonEmpty(region, p["region"]), firstNonEmpty(baseURL, p["api_url"], p["base_url"]), credentialsPath() + " [" + profile + "]"
		}
	}
	if t := tokenFromCLI(); t != "" {
		return t, region, baseURL, filepath.Join(filepath.Dir(credentialsPath()), "token")
	}
	return "", region, baseURL, "none"
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
