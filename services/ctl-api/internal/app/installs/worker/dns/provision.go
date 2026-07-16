package installdelegationdns

import (
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"
)

type ProvisionDNSDelegationRequest struct {
	WorkflowID  string   `json:"workflow_id"`
	Domain      string   `json:"domain"`
	Nameservers []string `json:"nameservers"`
}

func (d ProvisionDNSDelegationRequest) Validate() error {
	return nil
}

type ProvisionDNSDelegationResponse struct{}

// @temporal-gen-v2 workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-template {{ .CallerID }}-provision-dns-delegation
func (w Wkflow) ProvisionDNSDelegation(ctx workflow.Context, req *ProvisionDNSDelegationRequest) (*ProvisionDNSDelegationResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// DNS names from cloud providers (e.g. GCP Cloud DNS) may carry a trailing
	// dot; normalize both sides so the delegation isn't silently skipped on a
	// dot-only mismatch.
	domain := strings.TrimSuffix(req.Domain, ".")
	rootDomain := strings.TrimSuffix(w.cfg.DNSRootDomain, ".")
	if !strings.Contains(domain, rootDomain) {
		return nil, nil
	}

	delegateReq := DelegateDNSRequest{
		DNSAccessIAMRoleARN: w.cfg.DNSManagementIAMRoleARN,
		ZoneID:              w.cfg.DNSZoneID,
		Domain:              req.Domain,
		NameServers:         req.Nameservers,
	}
	_, err := AwaitDelegateDNS(ctx, delegateReq)
	if err != nil {
		err = fmt.Errorf("failed to delegate dns: %w", err)
		return nil, err
	}

	return &ProvisionDNSDelegationResponse{}, nil
}

func ensureTrailingDot(s string) string {
	if s == "" || strings.HasSuffix(s, ".") {
		return s
	}
	return s + "."
}
