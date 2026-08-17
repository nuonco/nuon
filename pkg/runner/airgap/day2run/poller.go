package day2run

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
)

const (
	defaultPollInterval = 15 * time.Second
	claimTTL            = 10 * time.Minute
	claimGrace          = 2 * time.Minute
)

type RefExecutor interface {
	Busy() bool
	Execute(context.Context, day2.Request, string) (*day2.RunStatus, error)
}

type dispatcher struct {
	mailbox      *Mailbox
	envelope     *airgap.Envelope
	digest       string
	deploymentID string
	owner        string
	executor     RefExecutor
	flushRun     func(context.Context, string) error
	logger       *zap.Logger
	now          func() time.Time
}

type Poller struct {
	dispatcher *dispatcher
	interval   time.Duration
}

func (p *Poller) Run(ctx context.Context) {
	p.poll(ctx)
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	requests, err := p.dispatcher.mailbox.ListRequests(ctx)
	if err != nil {
		p.dispatcher.logger.Warn("list day-2 requests", zap.Error(err))
		return
	}
	for _, listed := range requests {
		if err := p.dispatcher.handle(ctx, listed.DispatchID, listed.Request); err != nil {
			p.dispatcher.logger.Warn("handle day-2 request", zap.String("dispatch_id", listed.DispatchID), zap.Error(err))
		}
	}
}

func (d *dispatcher) handle(ctx context.Context, id string, req day2.Request) error {
	if _, found, err := d.mailbox.GetReceipt(ctx, id); err != nil || found {
		return err
	}
	if reason := d.rejectReason(id, req); reason != "" {
		return d.mailbox.PutReceipt(ctx, day2.Receipt{DispatchID: id, RefID: req.RefID, Status: day2.ReceiptStatusRejected, Reason: reason, FinishedAt: d.now().UTC()})
	}
	claim, won, err := d.claim(ctx, id, req)
	if err != nil || !won {
		return err
	}
	run, executeErr := d.executor.Execute(ctx, req, claim.RunID)
	if d.flushRun != nil {
		if err := d.flushRun(ctx, claim.RunID); err != nil {
			d.logger.Warn("flush day-2 run", zap.String("run_id", claim.RunID), zap.Error(err))
		}
	}
	status := day2.ReceiptStatusFinished
	reason := ""
	if executeErr != nil || run == nil || run.Status == day2.RunStatusFailed {
		status = day2.ReceiptStatusFailed
		if executeErr != nil {
			reason = executeErr.Error()
		} else if run != nil {
			reason = run.Error
		}
	}
	return d.mailbox.PutReceipt(ctx, day2.Receipt{DispatchID: id, RefID: req.RefID, RunID: claim.RunID, Status: status, Reason: reason, FinishedAt: d.now().UTC()})
}

func (d *dispatcher) rejectReason(id string, req day2.Request) string {
	if err := day2.ValidateDispatchID(id); err != nil {
		return err.Error()
	}
	if req.DispatchID != id {
		return fmt.Sprintf("request dispatch ID %q does not match key dispatch ID %q", req.DispatchID, id)
	}
	if req.DeploymentID != d.deploymentID {
		return fmt.Sprintf("deployment ID mismatch: got %q", req.DeploymentID)
	}
	if req.RefKind == day2.RefKindBundlePlan {
		if err := req.ValidateBundlePlan(); err != nil {
			return err.Error()
		}
		return ""
	}
	if req.BundleDigest != d.digest {
		return fmt.Sprintf("bundle digest mismatch: got %q", req.BundleDigest)
	}
	if d.envelope.FindAction(req.RefID) == nil && d.envelope.FindDrift(req.RefID) == nil && d.envelope.FindRunbook(req.RefID) == nil {
		return fmt.Sprintf("unknown day-2 ref %q", req.RefID)
	}
	return ""
}

func (d *dispatcher) claim(ctx context.Context, id string, req day2.Request) (*day2.Claim, bool, error) {
	now := d.now().UTC()
	existing, etag, found, err := d.mailbox.GetClaim(ctx, id)
	if err != nil {
		return nil, false, err
	}
	if !found {
		runID := ulid.Make().String()
		if req.RefKind == day2.RefKindBundlePlan {
			runID = req.RunID
		}
		claim := &day2.Claim{DispatchID: id, Owner: d.owner, RunID: runID, Attempt: 1, CreatedAt: now, ExpiresAt: now.Add(claimTTL)}
		if err := d.mailbox.ClaimNew(ctx, *claim); err != nil {
			if errors.Is(err, ErrAlreadyClaimed) {
				return nil, false, nil
			}
			return nil, false, err
		}
		return claim, true, nil
	}
	if now.Before(existing.ExpiresAt.Add(claimGrace)) {
		return nil, false, nil
	}
	runID := ulid.Make().String()
	if req.RefKind == day2.RefKindBundlePlan {
		runID = req.RunID
	}
	claim := &day2.Claim{DispatchID: id, Owner: d.owner, RunID: runID, Attempt: existing.Attempt + 1, CreatedAt: now, ExpiresAt: now.Add(claimTTL)}
	if err := d.mailbox.TakeOverClaim(ctx, *claim, etag); err != nil {
		if isConditionFailed(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return claim, true, nil
}
