package installdelegationdns

import (
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"
)

type DeprovisionDNSDelegationRequest struct {
	WorkflowID string `json:"workflow_id"`
	Domain     string `json:"domain"`
}

type DeprovisionDNSDelegationResponse struct{}

// @temporal-gen-v2 workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-template {{ .CallerID }}-deprovision-dns-delegation
func (w Wkflow) DeprovisionDNSDelegation(ctx workflow.Context, req *DeprovisionDNSDelegationRequest) (*DeprovisionDNSDelegationResponse, error) {
	domain := strings.TrimSuffix(req.Domain, ".")
	rootDomain := strings.TrimSuffix(w.cfg.DNSRootDomain, ".")
	if domain == "" || !strings.Contains(domain, rootDomain) {
		return &DeprovisionDNSDelegationResponse{}, nil
	}

	if _, err := AwaitDeleteDNS(ctx, DeleteDNSRequest{
		DNSAccessIAMRoleARN: w.cfg.DNSManagementIAMRoleARN,
		ZoneID:              w.cfg.DNSZoneID,
		Domain:              req.Domain,
	}); err != nil {
		return nil, fmt.Errorf("failed to delete dns delegation: %w", err)
	}

	return &DeprovisionDNSDelegationResponse{}, nil
}
