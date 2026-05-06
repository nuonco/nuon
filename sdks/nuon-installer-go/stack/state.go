// Package stack provisions and tears down the AWS resources that make up
// a Nuon install stack: S3 artifact bucket, VPC + subnets, IAM roles, EKS
// control plane, and a managed node group.
package stack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// State records the AWS resource IDs created during provisioning so we can
// tear down idempotently or surface outputs to the caller.
type State struct {
	InstallID string `json:"install_id"`
	Region    string `json:"region"`

	S3BucketName string `json:"s3_bucket_name,omitempty"`

	VPCID            string   `json:"vpc_id,omitempty"`
	IGWID            string   `json:"igw_id,omitempty"`
	NATGatewayID     string   `json:"nat_gateway_id,omitempty"`
	NATElasticIP     string   `json:"nat_elastic_ip,omitempty"`
	PublicSubnetIDs  []string `json:"public_subnet_ids,omitempty"`
	PrivateSubnetIDs []string `json:"private_subnet_ids,omitempty"`
	PublicRouteTable string   `json:"public_route_table,omitempty"`
	PrivateRouteTbl  string   `json:"private_route_table,omitempty"`

	ClusterRoleName string `json:"cluster_role_name,omitempty"`
	NodeRoleName    string `json:"node_role_name,omitempty"`

	ClusterName     string `json:"cluster_name,omitempty"`
	NodeGroupName   string `json:"node_group_name,omitempty"`
	ClusterEndpoint string `json:"cluster_endpoint,omitempty"`
	ClusterCAData   string `json:"cluster_ca_data,omitempty"`
	OIDCIssuer      string `json:"oidc_issuer,omitempty"`
}

// StatePath returns the on-disk location of the state file for an install.
func StatePath(installID string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cache, "nuon", "installer-sdk")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, installID+".json"), nil
}

// LoadState reads state from disk; returns a fresh state if none exists.
func LoadState(installID, region string) (*State, error) {
	p, err := StatePath(installID)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &State{InstallID: installID, Region: region}, nil
	}
	if err != nil {
		return nil, err
	}
	s := &State{}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	if s.InstallID == "" {
		s.InstallID = installID
	}
	if s.Region == "" {
		s.Region = region
	}
	return s, nil
}

// Save flushes state to disk. Called after every successful resource create
// so that a mid-flight failure leaves the state consistent for deprovision.
func (s *State) Save() error {
	p, err := StatePath(s.InstallID)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// Delete removes the state file (after deprovision).
func (s *State) Delete() error {
	p, err := StatePath(s.InstallID)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
