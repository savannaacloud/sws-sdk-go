package sws

import "context"

// NetworkService manages networks, subnets, security groups, and public IPs.
type NetworkService struct {
	client *Client
}

// ─── Networks ───────────────────────────────────────────────────────────

func (s *NetworkService) ListNetworks(ctx context.Context) ([]Network, error) {
	var out []Network
	if err := s.client.do(ctx, "GET", "/api/network/networks", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *NetworkService) CreateNetwork(ctx context.Context, name, description string) (*Network, error) {
	body := map[string]string{"name": name}
	if description != "" {
		body["description"] = description
	}
	var out Network
	if err := s.client.do(ctx, "POST", "/api/network/networks", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *NetworkService) DeleteNetwork(ctx context.Context, id string) error {
	return s.client.do(ctx, "DELETE", "/api/network/networks/"+id, nil, nil)
}

// ─── Subnets ────────────────────────────────────────────────────────────

func (s *NetworkService) ListSubnets(ctx context.Context) ([]Subnet, error) {
	var out []Subnet
	if err := s.client.do(ctx, "GET", "/api/network/subnets", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *NetworkService) CreateSubnet(ctx context.Context, opts *CreateSubnetOpts) (*Subnet, error) {
	if opts.IPVersion == 0 {
		opts.IPVersion = 4
	}
	var out Subnet
	if err := s.client.do(ctx, "POST", "/api/network/subnets", opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *NetworkService) DeleteSubnet(ctx context.Context, id string) error {
	return s.client.do(ctx, "DELETE", "/api/network/subnets/"+id, nil, nil)
}

// ─── Security groups ────────────────────────────────────────────────────

func (s *NetworkService) ListSecurityGroups(ctx context.Context) ([]SecurityGroup, error) {
	var out []SecurityGroup
	if err := s.client.do(ctx, "GET", "/api/network/security-groups", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *NetworkService) CreateSecurityGroup(ctx context.Context, name, description string) (*SecurityGroup, error) {
	body := map[string]string{"name": name, "description": description}
	var out SecurityGroup
	if err := s.client.do(ctx, "POST", "/api/network/security-groups", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *NetworkService) DeleteSecurityGroup(ctx context.Context, id string) error {
	return s.client.do(ctx, "DELETE", "/api/network/security-groups/"+id, nil, nil)
}

// AddSecurityGroupRule appends a single ingress/egress rule. Defaults
// applied: Direction=ingress, Ethertype=IPv4 (when blank), RemoteIPPrefix=0.0.0.0/0
// (when blank). Returns the new rule's ID and other server-assigned fields.
func (s *NetworkService) AddSecurityGroupRule(ctx context.Context, opts *AddSecurityGroupRuleOpts) (*SecurityGroupRule, error) {
	if opts.Direction == "" {
		opts.Direction = "ingress"
	}
	if opts.Ethertype == "" {
		opts.Ethertype = "IPv4"
	}
	if opts.RemoteIPPrefix == "" {
		opts.RemoteIPPrefix = "0.0.0.0/0"
	}
	var out SecurityGroupRule
	if err := s.client.do(ctx, "POST", "/api/network/security-group-rules", opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *NetworkService) RemoveSecurityGroupRule(ctx context.Context, ruleID string) error {
	return s.client.do(ctx, "DELETE", "/api/network/security-group-rules/"+ruleID, nil, nil)
}

// ─── Public IPs ─────────────────────────────────────────────────────────

func (s *NetworkService) ListPublicIPs(ctx context.Context) ([]PublicIP, error) {
	var out []PublicIP
	if err := s.client.do(ctx, "GET", "/api/network/public-ips", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AllocatePublicIP reserves a new public IP for the tenant. Pass
// floatingNetworkID="" to use the platform's default external network.
func (s *NetworkService) AllocatePublicIP(ctx context.Context, floatingNetworkID string) (*PublicIP, error) {
	body := map[string]string{}
	if floatingNetworkID != "" {
		body["floating_network_id"] = floatingNetworkID
	}
	var out PublicIP
	if err := s.client.do(ctx, "POST", "/api/network/public-ips", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AssignPublicIP attaches an allocated IP to an instance.
func (s *NetworkService) AssignPublicIP(ctx context.Context, ipID, instanceID string) error {
	return s.client.do(ctx, "POST", "/api/network/public-ips/"+ipID+"/associate",
		map[string]string{"instance_id": instanceID}, nil)
}

// UnassignPublicIP detaches an IP from its instance without releasing it.
func (s *NetworkService) UnassignPublicIP(ctx context.Context, ipID string) error {
	return s.client.do(ctx, "POST", "/api/network/public-ips/"+ipID+"/disassociate", nil, nil)
}

// ReleasePublicIP returns an IP to the pool.
func (s *NetworkService) ReleasePublicIP(ctx context.Context, ipID string) error {
	return s.client.do(ctx, "DELETE", "/api/network/public-ips/"+ipID, nil, nil)
}
