package dns

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/dns/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg workers.Config
}

// ProvisionDNS is used to provision DNS for the nuon.run domain delegation
//
//nolint:funlen
func (w wkflow) ProvisionDNS(ctx workflow.Context, req *runnerv1.ProvisionDNSRequest) (*runnerv1.ProvisionDNSResponse, error) {
	resp := &runnerv1.ProvisionDNSResponse{}
	l := workflow.GetLogger(ctx)

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	act := NewActivities(nil)

	delegateReq := DelegateDNSRequest{
		DNSAccessIAMRoleARN: w.cfg.PublicDNSAccessRoleARN,
		ZoneID:              w.cfg.PublicDomainZoneID,
		Domain:              req.Domain,
		NameServers:         req.Nameservers,
	}
	_, err := execDelegateDNS(ctx, act, delegateReq)
	if err != nil {
		err = fmt.Errorf("failed to delegate dns: %w", err)
		return resp, err
	}

	l.Debug("finished provisioning dns", "response", resp)
	return resp, nil
}

// createWaypointProject executes an activity to create the waypoint project on the org's server
func execDelegateDNS(ctx workflow.Context, act *Activities, req DelegateDNSRequest) (DelegateDNSResponse, error) {
	var resp DelegateDNSResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing delegate dns activity")
	fut := workflow.ExecuteActivity(ctx, act.DelegateDNS, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
