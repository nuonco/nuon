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

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-template {{ .CallerID }}-provision-dns-delegation
func (w Wkflow) ProvisionDNSDelegation(ctx workflow.Context, req *ProvisionDNSDelegationRequest) (*ProvisionDNSDelegationResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if !strings.Contains(req.Domain, defaultNuonRunDomain) {
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
