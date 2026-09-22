// Package sws is the official Go SDK for the SWS cloud platform.
//
// Quickstart:
//
//	client := sws.NewClient("ctk_...", sws.WithRegion("ng-lagos-1"))
//	instances, err := client.Compute.ListInstances(ctx)
package sws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Version is the SDK version reported in the User-Agent header.
const Version = "0.1.1"

const (
	defaultBaseURL = "https://savannaa.com"
	defaultRegion  = "ng-lagos-1"
	defaultTimeout = 30 * time.Second
)

// Client is the entry point to the SWS API. Construct one with NewClient
// and reuse it across goroutines — *Client is safe for concurrent use.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	region     string
	userAgent  string

	Compute    *ComputeService
	Network    *NetworkService
	Storage    *StorageService
	Database   *DatabaseService
	credSource string // where apiKey came from — see credentials.go
}

// Option configures a Client at construction time.
type Option func(*Client)

// WithRegion sets the x-region header sent with every request.
// Default: "ng-lagos-1" (or $SWS_REGION).
func WithRegion(region string) Option {
	return func(c *Client) { c.region = region }
}

// WithBaseURL overrides the API base URL.
// Default: "https://savannaa.com" (or $SWS_BASE_URL).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = baseURL }
}

// WithHTTPClient swaps in a caller-provided *http.Client (for custom
// transports, proxies, or test doubles).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// WithTimeout sets the request timeout on the underlying http.Client.
// Default: 30 seconds.
func WithTimeout(t time.Duration) Option {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{Timeout: t}
		} else {
			c.httpClient.Timeout = t
		}
	}
}

// NewClient returns a Client ready for use. Pass "" and the credentials are found the
// way every modern SDK finds them — $SWS_API_KEY, then the file `sws auth login` writes
// — so a quick start never has to contain a secret. See credentials.go.
func NewClient(apiKey string, opts ...Option) *Client {
	key, region, baseURL, source := resolveCredentials(apiKey)
	c := &Client{
		apiKey:     key,
		baseURL:    defaultBaseURL,
		region:     defaultRegion,
		httpClient: &http.Client{Timeout: defaultTimeout},
		userAgent:  "sws-sdk-go/" + Version,
		credSource: source,
	}
	if baseURL != "" {
		c.baseURL = baseURL
	}
	if region != "" {
		c.region = region
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Compute = &ComputeService{client: c}
	c.Network = &NetworkService{client: c}
	c.Storage = &StorageService{client: c}
	c.Database = &DatabaseService{client: c}
	return c
}

// Region returns the region the client is currently scoped to.
func (c *Client) Region() string { return c.region }

// CredentialSource says where the API key came from — "argument", "SWS_API_KEY", the
// credentials file, or "none". Useful when a script authenticates as somebody unexpected.
func (c *Client) CredentialSource() string { return c.credSource }

// do executes an HTTP request and decodes the JSON response into out.
// If out is nil, the body is discarded but errors are still surfaced.
// Non-2xx responses are translated into the *Error hierarchy in errors.go.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	if c.apiKey == "" {
		return &AuthenticationError{&APIError{
			StatusCode: 401,
			Message:    MissingCredentials,
		}}
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("sws: marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("sws: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("x-region", c.region)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sws: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("sws: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseError(resp.StatusCode, respBody)
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("sws: decode response: %w", err)
	}
	return nil
}
