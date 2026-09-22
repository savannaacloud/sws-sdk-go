package sws

import "encoding/json"

// Instance is a virtual machine. The "Plan" field carries what the
// underlying platform calls a flavor — same shape, friendlier name.
type Instance struct {
	ID        string                     `json:"id"`
	Name      string                     `json:"name"`
	Status    string                     `json:"status"`
	Plan      map[string]any             `json:"flavor,omitempty"` // backend uses "flavor" key
	Image     map[string]any             `json:"image,omitempty"`
	Addresses map[string]any             `json:"addresses,omitempty"`
	KeyName   string                     `json:"key_name,omitempty"`
	Created   string                     `json:"created,omitempty"`
	Raw       map[string]json.RawMessage `json:"-"`
}

// Plan is a compute size/SKU.
type Plan struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	VCPUs int    `json:"vcpus"`
	RAM   int    `json:"ram"`
	Disk  int    `json:"disk"`
}

// Image is an OS image. The shape is left flexible because different
// backends expose different metadata sets.
type Image map[string]any

// Keypair is an SSH key pair stored in the platform.
// PrivateKey is only populated on Create, never on List/Get.
type Keypair struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint,omitempty"`
	PublicKey   string `json:"public_key,omitempty"`
	PrivateKey  string `json:"private_key,omitempty"`
}

// Network is a layer-2 isolated network.
type Network struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Status  string   `json:"status,omitempty"`
	Subnets []string `json:"subnets,omitempty"`
}

// Subnet is an IPv4 or IPv6 range inside a Network.
type Subnet struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	NetworkID  string `json:"network_id"`
	CIDR       string `json:"cidr"`
	IPVersion  int    `json:"ip_version"`
	EnableDHCP bool   `json:"enable_dhcp"`
}

// SecurityGroup is a layer-4 firewall rule set.
type SecurityGroup struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Rules       []SecurityGroupRule `json:"security_group_rules,omitempty"`
}

// SecurityGroupRule is one direction/protocol/port-range entry inside a SecurityGroup.
type SecurityGroupRule struct {
	ID             string `json:"id,omitempty"`
	Direction      string `json:"direction"`
	Protocol       string `json:"protocol"`
	PortRangeMin   int    `json:"port_range_min"`
	PortRangeMax   int    `json:"port_range_max"`
	RemoteIPPrefix string `json:"remote_ip_prefix"`
	Ethertype      string `json:"ethertype,omitempty"`
}

// PublicIP is a routable IPv4 address that can be assigned to an instance.
type PublicIP struct {
	ID         string `json:"id"`
	Address    string `json:"floating_ip_address,omitempty"`
	InstanceID string `json:"port_id,omitempty"`
	Status     string `json:"status,omitempty"`
}

// Volume is a block-storage volume.
type Volume struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Size        int                `json:"size"` // GB
	Status      string             `json:"status,omitempty"`
	Type        string             `json:"volume_type,omitempty"`
	Attachments []VolumeAttachment `json:"attachments,omitempty"`
}

// VolumeAttachment links a Volume to an Instance.
type VolumeAttachment struct {
	ServerID string `json:"server_id"`
	Device   string `json:"device,omitempty"`
}

// Database is a managed database instance (mysql, postgresql, ...).
type Database struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Datastore json.RawMessage `json:"datastore,omitempty"`
	Status    string          `json:"status,omitempty"`
	Plan      map[string]any  `json:"flavor,omitempty"`
}

// ── Request option structs ──────────────────────────────────────────────

// CreateInstanceOpts is the payload for ComputeService.CreateInstance.
// Plan is sent over the wire as flavor_id (struct tag handles the rename).
type CreateInstanceOpts struct {
	Name           string   `json:"name"`
	Image          string   `json:"image_id"`
	Plan           string   `json:"flavor_id"`
	NetworkID      string   `json:"network_id,omitempty"`
	KeyName        string   `json:"key_name,omitempty"`
	SecurityGroups []string `json:"security_groups,omitempty"`
	UserData       string   `json:"user_data,omitempty"`
}

// CreateSubnetOpts is the payload for NetworkService.CreateSubnet.
type CreateSubnetOpts struct {
	Name           string   `json:"name"`
	NetworkID      string   `json:"network_id"`
	CIDR           string   `json:"cidr"`
	IPVersion      int      `json:"ip_version,omitempty"`
	EnableDHCP     *bool    `json:"enable_dhcp,omitempty"`
	DNSNameservers []string `json:"dns_nameservers,omitempty"`
}

// AddSecurityGroupRuleOpts is the payload for NetworkService.AddSecurityGroupRule.
type AddSecurityGroupRuleOpts struct {
	GroupID        string `json:"security_group_id"`
	Direction      string `json:"direction"` // "ingress" or "egress"
	Protocol       string `json:"protocol"`
	PortRangeMin   int    `json:"port_range_min"`
	PortRangeMax   int    `json:"port_range_max"`
	RemoteIPPrefix string `json:"remote_ip_prefix"`
	Ethertype      string `json:"ethertype,omitempty"` // defaults to IPv4
}

// CreateVolumeOpts is the payload for StorageService.CreateVolume.
type CreateVolumeOpts struct {
	Name        string `json:"name"`
	Size        int    `json:"size"` // GB
	VolumeType  string `json:"volume_type,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateDatabaseOpts is the payload for DatabaseService.CreateInstance.
type CreateDatabaseOpts struct {
	Name             string `json:"name"`
	Datastore        string `json:"datastore_type"`
	DatastoreVersion string `json:"datastore_version"`
	Plan             string `json:"flavor"`
	Size             int    `json:"size"` // GB
	AdminUser        string `json:"admin_user,omitempty"`
	AdminPassword    string `json:"admin_password"`
	NetworkID        string `json:"network_id,omitempty"`
}
