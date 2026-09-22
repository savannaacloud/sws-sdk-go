package sws

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer wires a Client to an httptest.Server with the given handler.
// Both are torn down by t.Cleanup so individual tests stay short.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := NewClient("ctk_test", WithBaseURL(srv.URL), WithRegion("ng-lagos-1"))
	return srv, client
}

func mustJSON(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAuthHeaderAndRegionSent(t *testing.T) {
	var gotAuth, gotRegion, gotUA string
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotRegion = r.Header.Get("x-region")
		gotUA = r.Header.Get("User-Agent")
		mustJSON(t, w, 200, []Instance{})
	})
	if _, err := client.Compute.ListInstances(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer ctk_test" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer ctk_test")
	}
	if gotRegion != "ng-lagos-1" {
		t.Errorf("x-region = %q, want %q", gotRegion, "ng-lagos-1")
	}
	if !strings.HasPrefix(gotUA, "sws-sdk-go/") {
		t.Errorf("User-Agent = %q, want prefix %q", gotUA, "sws-sdk-go/")
	}
}

func TestListInstancesParsesFlavorAsPlan(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		mustJSON(t, w, 200, []map[string]any{{
			"id":     "i-1",
			"name":   "web-1",
			"status": "ACTIVE",
			"flavor": map[string]any{"id": "m1.small", "vcpus": 1, "ram": 2048},
		}})
	})
	instances, err := client.Compute.ListInstances(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(instances) != 1 {
		t.Fatalf("got %d instances, want 1", len(instances))
	}
	if instances[0].Plan["id"] != "m1.small" {
		t.Errorf("Plan[id] = %v, want m1.small", instances[0].Plan["id"])
	}
}

func TestCreateInstanceTranslatesPlanToFlavorID(t *testing.T) {
	var gotBody string
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		mustJSON(t, w, 201, Instance{ID: "i-2", Name: "web-2", Status: "BUILD"})
	})
	inst, err := client.Compute.CreateInstance(context.Background(), &CreateInstanceOpts{
		Name:      "web-2",
		Image:     "ubuntu-22.04",
		Plan:      "m1.medium",
		NetworkID: "net-1",
		KeyName:   "my-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	if inst.ID != "i-2" {
		t.Errorf("ID = %q", inst.ID)
	}
	// "Plan" struct field maps to flavor_id JSON tag — that's the contract
	// users count on. If this assertion ever fires, someone changed the
	// tag and the legacy field name is leaking.
	if !strings.Contains(gotBody, `"flavor_id":"m1.medium"`) {
		t.Errorf("body should contain flavor_id, got: %s", gotBody)
	}
	if strings.Contains(gotBody, `"plan"`) {
		t.Errorf("body should NOT contain raw 'plan' key, got: %s", gotBody)
	}
}

func TestNotFoundError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		mustJSON(t, w, 404, map[string]string{"detail": "Not found"})
	})
	_, err := client.Compute.GetInstance(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false, err = %v", err)
	}
	var nfe *NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("errors.As to *NotFoundError failed, got %T", err)
	}
}

func TestQuotaExceededError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		mustJSON(t, w, 403, map[string]string{
			"detail": "Quota exceeded for instances: 10/10",
		})
	})
	_, err := client.Compute.CreateInstance(context.Background(), &CreateInstanceOpts{
		Name: "x", Image: "i", Plan: "m1.tiny",
	})
	if !errors.Is(err, ErrQuota) {
		t.Errorf("errors.Is(err, ErrQuota) = false, err = %v", err)
	}
}

func TestAuthenticationError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		mustJSON(t, w, 401, map[string]string{"detail": "bad token"})
	})
	_, err := client.Compute.ListInstances(context.Background())
	if !errors.Is(err, ErrAuthentication) {
		t.Errorf("errors.Is(err, ErrAuthentication) = false, err = %v", err)
	}
}

func TestValidationError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		mustJSON(t, w, 422, map[string]string{"detail": "cidr required"})
	})
	_, err := client.Network.CreateSubnet(context.Background(), &CreateSubnetOpts{
		Name: "s", NetworkID: "n", CIDR: "",
	})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("errors.Is(err, ErrValidation) = false, err = %v", err)
	}
}

func TestSecurityGroupRuleDefaults(t *testing.T) {
	var gotBody string
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		mustJSON(t, w, 201, SecurityGroupRule{ID: "r-1"})
	})
	_, err := client.Network.AddSecurityGroupRule(context.Background(), &AddSecurityGroupRuleOpts{
		GroupID:      "sg-1",
		Protocol:     "tcp",
		PortRangeMin: 22,
		PortRangeMax: 22,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"security_group_id":"sg-1"`,
		`"direction":"ingress"`,
		`"remote_ip_prefix":"0.0.0.0/0"`,
		`"ethertype":"IPv4"`,
	} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("body missing %q, got: %s", want, gotBody)
		}
	}
}

func TestVolumeAttachUsesInstanceID(t *testing.T) {
	var gotBody string
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(202)
	})
	if err := client.Storage.AttachVolume(context.Background(), "v-1", "i-9"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotBody, `"instance_id":"i-9"`) {
		t.Errorf("body missing instance_id: %s", gotBody)
	}
}

func TestMissingAPIKeyReturnsAuthError(t *testing.T) {
	t.Setenv("SWS_API_KEY", "")
	client := NewClient("")
	_, err := client.Compute.ListInstances(context.Background())
	if !errors.Is(err, ErrAuthentication) {
		t.Errorf("expected ErrAuthentication, got %v", err)
	}
}

func TestEnvVarResolution(t *testing.T) {
	t.Setenv("SWS_API_KEY", "sws_from_env")
	t.Setenv("SWS_REGION", "ng-abuja-1")
	client := NewClient("")
	if client.Region() != "ng-abuja-1" {
		t.Errorf("region = %q, want ng-abuja-1", client.Region())
	}
	if client.apiKey != "sws_from_env" {
		t.Errorf("apiKey = %q, want sws_from_env", client.apiKey)
	}
}
