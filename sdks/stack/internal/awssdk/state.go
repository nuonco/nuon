// Package stack provisions and tears down the AWS resources that make up a
// Nuon install stack: VPC + subnets (incl. dedicated runner subnet/SG), IAM
// roles (runner + ops + dynamic break-glass/custom), Secrets Manager entries,
// and the runner EC2 ASG with its CloudWatch log group. Mirrors the reference
// Terraform module at install-stacks/aws/.
package stack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// State caches the AWS resource IDs we've already discovered or created. AWS
// is the source of truth — every Provision step does its own existence check
// (by tag or deterministic name) and falls back to a Create when the resource
// is gone. State just lets us skip those Describe calls on the happy path and
// gives Deprovision a quick handle on what to tear down.
type State struct {
	InstallID   string `json:"install_id"`
	OrgID       string `json:"org_id,omitempty"`
	AppID       string `json:"app_id,omitempty"`
	ClusterName string `json:"cluster_name,omitempty"`
	Region      string `json:"region"`

	// Networking
	VPCID            string   `json:"vpc_id,omitempty"`
	IGWID            string   `json:"igw_id,omitempty"`
	NATGatewayID     string   `json:"nat_gateway_id,omitempty"`
	NATElasticIP     string   `json:"nat_elastic_ip,omitempty"`
	PublicSubnetIDs  []string `json:"public_subnet_ids,omitempty"`
	PrivateSubnetIDs []string `json:"private_subnet_ids,omitempty"`
	PublicRouteTable string   `json:"public_route_table,omitempty"`
	PrivateRouteTbl  string   `json:"private_route_table,omitempty"`

	RunnerSubnetID        string `json:"runner_subnet_id,omitempty"`
	RunnerSecurityGroupID string `json:"runner_security_group_id,omitempty"`

	// Runner compute
	RunnerLogGroupName        string `json:"runner_log_group_name,omitempty"`
	RunnerRoleName            string `json:"runner_role_name,omitempty"`
	RunnerInstanceProfileName string `json:"runner_instance_profile_name,omitempty"`
	RunnerLaunchTemplateID    string `json:"runner_launch_template_id,omitempty"`
	RunnerASGName             string `json:"runner_asg_name,omitempty"`

	// Operation roles. Empty when the ctl-api config didn't request them.
	ProvisionRoleName   string `json:"provision_role_name,omitempty"`
	MaintenanceRoleName string `json:"maintenance_role_name,omitempty"`
	DeprovisionRoleName string `json:"deprovision_role_name,omitempty"`

	// Dynamic role names (verbatim map keys from the config — no extra prefixing).
	BreakGlassRoleNames []string `json:"break_glass_role_names,omitempty"`
	CustomRoleNames     []string `json:"custom_role_names,omitempty"`

	// Secrets Manager ARNs keyed by `<name>_arn` to match the CloudFormation
	// phone-home payload contract. Why: app templates resolving
	// `nuon.install_stack.outputs.<name>_arn` must work identically across
	// CFN, TF, and SDK install paths.
	SecretARNs map[string]string `json:"secret_arns,omitempty"`
}

// statePath returns the on-disk location of the state file for an install.
func statePath(installID string) (string, error) {
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

// loadState reads state from disk; returns a fresh state if none exists.
func loadState(installID, region string) (*State, error) {
	p, err := statePath(installID)
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
	p, err := statePath(s.InstallID)
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
	p, err := statePath(s.InstallID)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
