package sws

import "context"

// ComputeService manages virtual machines, plans, images, and keypairs.
type ComputeService struct {
	client *Client
}

// ─── Instances ──────────────────────────────────────────────────────────

// ListInstances returns every VM in the configured region.
func (s *ComputeService) ListInstances(ctx context.Context) ([]Instance, error) {
	var out []Instance
	if err := s.client.do(ctx, "GET", "/api/compute/servers", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetInstance returns a single VM by ID.
func (s *ComputeService) GetInstance(ctx context.Context, id string) (*Instance, error) {
	var out Instance
	if err := s.client.do(ctx, "GET", "/api/compute/servers/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateInstance launches a new VM. The Plan field on opts is sent as
// flavor_id over the wire; callers never need to know that.
func (s *ComputeService) CreateInstance(ctx context.Context, opts *CreateInstanceOpts) (*Instance, error) {
	var out Instance
	if err := s.client.do(ctx, "POST", "/api/compute/servers", opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteInstance terminates a VM. Returns nil on success even if the VM
// is already gone — callers can treat this as idempotent.
func (s *ComputeService) DeleteInstance(ctx context.Context, id string) error {
	return s.client.do(ctx, "DELETE", "/api/compute/servers/"+id, nil, nil)
}

// StartInstance boots a stopped VM.
func (s *ComputeService) StartInstance(ctx context.Context, id string) error {
	return s.client.do(ctx, "POST", "/api/compute/servers/"+id+"/start", nil, nil)
}

// StopInstance gracefully halts a running VM.
func (s *ComputeService) StopInstance(ctx context.Context, id string) error {
	return s.client.do(ctx, "POST", "/api/compute/servers/"+id+"/stop", nil, nil)
}

// RebootInstance restarts a VM. Pass hard=true to power-cycle.
func (s *ComputeService) RebootInstance(ctx context.Context, id string, hard bool) error {
	body := map[string]string{"type": "SOFT"}
	if hard {
		body["type"] = "HARD"
	}
	return s.client.do(ctx, "POST", "/api/compute/servers/"+id+"/reboot", body, nil)
}

// ResizeInstance changes the plan (flavor) of a VM in place.
func (s *ComputeService) ResizeInstance(ctx context.Context, id, plan string) error {
	return s.client.do(ctx, "POST", "/api/compute/servers/"+id+"/resize",
		map[string]string{"flavor_id": plan}, nil)
}

// ─── Plans, images, keypairs ────────────────────────────────────────────

// ListPlans returns every compute plan available in the region.
func (s *ComputeService) ListPlans(ctx context.Context) ([]Plan, error) {
	var out []Plan
	if err := s.client.do(ctx, "GET", "/api/compute/plans", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListImages returns every OS image visible to the tenant.
func (s *ComputeService) ListImages(ctx context.Context) ([]Image, error) {
	var out []Image
	if err := s.client.do(ctx, "GET", "/api/images", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListKeypairs returns every SSH keypair the tenant has uploaded.
func (s *ComputeService) ListKeypairs(ctx context.Context) ([]Keypair, error) {
	var out []Keypair
	if err := s.client.do(ctx, "GET", "/api/compute/keypairs", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateKeypair uploads an existing public key, or — if publicKey is
// empty — asks the platform to generate a new pair. In the latter case
// the returned Keypair has a non-empty PrivateKey; save it immediately
// because the platform does not store it.
func (s *ComputeService) CreateKeypair(ctx context.Context, name, publicKey string) (*Keypair, error) {
	body := map[string]string{"name": name}
	if publicKey != "" {
		body["public_key"] = publicKey
	}
	var out Keypair
	if err := s.client.do(ctx, "POST", "/api/compute/keypairs", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteKeypair removes a stored keypair.
func (s *ComputeService) DeleteKeypair(ctx context.Context, name string) error {
	return s.client.do(ctx, "DELETE", "/api/compute/keypairs/"+name, nil, nil)
}
