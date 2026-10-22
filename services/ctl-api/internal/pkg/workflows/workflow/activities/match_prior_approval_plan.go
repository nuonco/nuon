package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type MatchPriorApprovalPlanRequest struct {
	OrgID           string `json:"org_id" validate:"required"`
	InstallID       string `json:"install_id" validate:"required"`
	WorkflowID      string `json:"workflow_id" validate:"required"`
	StepID          string `json:"step_id" validate:"required"`
	ApprovalID      string `json:"approval_id" validate:"required"`
	PriorApprovalID string `json:"prior_approval_id" validate:"required"`
}

type MatchPriorApprovalPlanResult struct {
	Matched         bool   `json:"matched"`
	Reason          string `json:"reason"`
	PriorResponseID string `json:"prior_response_id,omitempty"`
}

type priorApprovalMismatch string

func (m priorApprovalMismatch) Error() string { return string(m) }

// @temporal-gen-v2 activity
// @max-retries 3
// @start-to-close-timeout 1m
// @schedule-to-close-timeout 5m
func (a *Activities) MatchPriorApprovalPlan(ctx context.Context, req *MatchPriorApprovalPlanRequest) (*MatchPriorApprovalPlanResult, error) {
	l := temporalzap.GetActivityLogger(ctx).With(
		zap.String("workflow_id", req.WorkflowID),
		zap.String("step_id", req.StepID),
		zap.String("approval_id", req.ApprovalID),
		zap.String("prior_approval_id", req.PriorApprovalID),
	)

	priorResponseID, err := a.matchPriorApprovalPlan(ctx, req)
	var mismatch priorApprovalMismatch
	if errors.As(err, &mismatch) {
		l.Info("prior approval plan did not match", zap.String("reason", string(mismatch)))
		return &MatchPriorApprovalPlanResult{Reason: string(mismatch)}, nil
	}
	if err != nil {
		l.Error("unable to match prior approval plan", zap.Error(err))
		return nil, err
	}

	l.Info("prior approval plan matched", zap.String("prior_response_id", priorResponseID))
	return &MatchPriorApprovalPlanResult{
		Matched:         true,
		Reason:          "plan_matched",
		PriorResponseID: priorResponseID,
	}, nil
}

func (a *Activities) matchPriorApprovalPlan(ctx context.Context, req *MatchPriorApprovalPlanRequest) (string, error) {
	if req.ApprovalID == req.PriorApprovalID {
		return "", priorApprovalMismatch("prior_approval_is_current")
	}

	flw, err := a.loadPriorMatchWorkflow(ctx, req.WorkflowID, req.OrgID, req.InstallID)
	if err != nil {
		return "", err
	}

	var step app.WorkflowStep
	if err := a.findPriorMatchRecord(ctx, &step, req.StepID, app.WorkflowStep{ID: req.StepID}, "Approval", "Approval.RunnerJob"); err != nil {
		return "", err
	}
	if step.OrgID != req.OrgID || step.InstallWorkflowID != flw.ID {
		return "", priorApprovalMismatch("step_scope_mismatch")
	}
	current := step.Approval
	if current == nil || current.ID != req.ApprovalID {
		return "", priorApprovalMismatch("current_approval_mismatch")
	}
	if current.OrgID != req.OrgID || current.OwnerID != step.StepTargetID || current.OwnerType != step.StepTargetType {
		return "", priorApprovalMismatch("current_approval_scope_mismatch")
	}

	var prior app.WorkflowStepApproval
	if err := a.findPriorMatchRecord(ctx, &prior, req.PriorApprovalID, app.WorkflowStepApproval{ID: req.PriorApprovalID}, "Response", "RunnerJob", "InstallWorkflowStep"); err != nil {
		return "", err
	}
	if prior.OrgID != req.OrgID || prior.InstallWorkflowStep.OrgID != req.OrgID {
		return "", priorApprovalMismatch("prior_org_mismatch")
	}
	if _, err := a.loadPriorMatchWorkflow(ctx, prior.InstallWorkflowStep.InstallWorkflowID, req.OrgID, req.InstallID); err != nil {
		return "", err
	}

	resp := prior.Response
	if resp == nil || resp.Type != app.WorkflowStepApprovalResponseTypeApprove || resp.OrgID != req.OrgID {
		return "", priorApprovalMismatch("prior_not_explicitly_approved")
	}
	if strings.HasPrefix(resp.Note, "auto-approved") {
		return "", priorApprovalMismatch("prior_auto_approved")
	}

	switch current.Type {
	case app.TerraformPlanApprovalType, app.HelmApprovalApprovalType, app.KubernetesManifestApprovalType, app.PulumiApprovalType:
	default:
		return "", priorApprovalMismatch("unsupported_approval_type")
	}
	if prior.Type != current.Type || prior.OwnerType != current.OwnerType {
		return "", priorApprovalMismatch("approval_type_mismatch")
	}
	if prior.RunnerJob == nil || current.RunnerJob == nil || prior.RunnerJob.Type != current.RunnerJob.Type {
		return "", priorApprovalMismatch("runner_job_type_mismatch")
	}

	if err := a.matchApprovalTarget(ctx, req, current.OwnerType, prior.OwnerID, current.OwnerID); err != nil {
		return "", err
	}

	priorContents, _ := prior.GetContents(ctx, a.cfg.BlobReadEnabled)
	currentContents, _ := current.GetContents(ctx, a.cfg.BlobReadEnabled)
	priorPlan, err := canonicalApprovalPlan(prior.Type, priorContents)
	if err != nil {
		return "", priorApprovalMismatch("prior_plan_invalid")
	}
	currentPlan, err := canonicalApprovalPlan(current.Type, currentContents)
	if err != nil {
		return "", priorApprovalMismatch("current_plan_invalid")
	}
	if !bytes.Equal(priorPlan, currentPlan) {
		return "", priorApprovalMismatch("plan_changed")
	}

	return resp.ID, nil
}

func (a *Activities) findPriorMatchRecord(ctx context.Context, dest any, id string, where any, preloads ...string) error {
	if id == "" {
		return priorApprovalMismatch("record_id_missing")
	}
	q := a.db.WithContext(ctx).Where(where)
	for _, p := range preloads {
		q = q.Preload(p)
	}
	if err := q.First(dest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return priorApprovalMismatch("record_not_found")
		}
		return errors.Wrapf(err, "unable to load %T", dest)
	}
	return nil
}

func (a *Activities) loadPriorMatchWorkflow(ctx context.Context, workflowID, orgID, installID string) (*app.Workflow, error) {
	var flw app.Workflow
	if err := a.findPriorMatchRecord(ctx, &flw, workflowID, app.Workflow{ID: workflowID}); err != nil {
		return nil, err
	}
	if flw.OrgID != orgID || flw.OwnerType != "installs" || flw.OwnerID != installID {
		return nil, priorApprovalMismatch("workflow_scope_mismatch")
	}
	return &flw, nil
}

func (a *Activities) matchApprovalTarget(ctx context.Context, req *MatchPriorApprovalPlanRequest, ownerType, priorID, currentID string) error {
	switch ownerType {
	case "install_deploys":
		var prior, current app.InstallDeploy
		if err := a.findPriorMatchRecord(ctx, &prior, priorID, app.InstallDeploy{ID: priorID}, "InstallComponent"); err != nil {
			return err
		}
		if err := a.findPriorMatchRecord(ctx, &current, currentID, app.InstallDeploy{ID: currentID}, "InstallComponent"); err != nil {
			return err
		}
		for _, d := range []app.InstallDeploy{prior, current} {
			if d.OrgID != req.OrgID || d.InstallComponent.OrgID != req.OrgID || d.InstallComponent.InstallID != req.InstallID {
				return priorApprovalMismatch("deploy_scope_mismatch")
			}
		}
		if prior.InstallComponentID != current.InstallComponentID {
			return priorApprovalMismatch("install_component_mismatch")
		}
		if prior.ComponentBuildID != current.ComponentBuildID {
			return priorApprovalMismatch("component_build_mismatch")
		}
		if prior.Type != current.Type {
			return priorApprovalMismatch("deploy_type_mismatch")
		}
		return nil
	case "install_sandbox_runs":
		var prior, current app.InstallSandboxRun
		if err := a.findPriorMatchRecord(ctx, &prior, priorID, app.InstallSandboxRun{ID: priorID}); err != nil {
			return err
		}
		if err := a.findPriorMatchRecord(ctx, &current, currentID, app.InstallSandboxRun{ID: currentID}); err != nil {
			return err
		}
		for _, r := range []app.InstallSandboxRun{prior, current} {
			if r.OrgID != req.OrgID || r.InstallID != req.InstallID || r.InstallSandboxID == nil || r.AppSandboxConfigID == "" {
				return priorApprovalMismatch("sandbox_run_scope_mismatch")
			}
		}
		if *prior.InstallSandboxID != *current.InstallSandboxID {
			return priorApprovalMismatch("install_sandbox_mismatch")
		}
		if prior.AppSandboxConfigID != current.AppSandboxConfigID {
			return priorApprovalMismatch("sandbox_config_mismatch")
		}
		if prior.RunType != current.RunType {
			return priorApprovalMismatch("sandbox_run_type_mismatch")
		}
		return nil
	default:
		return priorApprovalMismatch("unsupported_approval_target")
	}
}

func canonicalApprovalPlan(approvalType app.WorkflowStepApprovalType, contents string) ([]byte, error) {
	if contents == "" {
		return nil, errors.New("empty plan")
	}
	if !utf8.ValidString(contents) {
		return nil, errors.New("plan is not valid utf-8")
	}

	dec := json.NewDecoder(strings.NewReader(contents))
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if tok != json.Delim('{') {
		return nil, errors.New("plan is not a json object")
	}

	var buf bytes.Buffer
	if err := writeCanonicalJSONObject(dec, &buf, approvalType == app.TerraformPlanApprovalType); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, errors.New("trailing data after plan")
	}
	return buf.Bytes(), nil
}

func writeCanonicalJSONValue(dec *json.Decoder, buf *bytes.Buffer) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}

	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return writeCanonicalJSONObject(dec, buf, false)
		case '[':
			buf.WriteByte('[')
			for i := 0; dec.More(); i++ {
				if i > 0 {
					buf.WriteByte(',')
				}
				if err := writeCanonicalJSONValue(dec, buf); err != nil {
					return err
				}
			}
			if end, err := dec.Token(); err != nil || end != json.Delim(']') {
				return errors.New("unterminated json array")
			}
			buf.WriteByte(']')
			return nil
		default:
			return errors.New("unexpected json delimiter")
		}
	case string:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}
		buf.Write(b)
	case json.Number:
		buf.WriteString(t.String())
	case bool:
		if t {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case nil:
		buf.WriteString("null")
	default:
		return errors.New("unexpected json token")
	}
	return nil
}

func writeCanonicalJSONObject(dec *json.Decoder, buf *bytes.Buffer, stripRunTimestamp bool) error {
	fields := map[string][]byte{}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := tok.(string)
		if !ok {
			return errors.New("json object key is not a string")
		}
		if _, dup := fields[key]; dup {
			return errors.New("duplicate json object key")
		}

		var val bytes.Buffer
		if err := writeCanonicalJSONValue(dec, &val); err != nil {
			return err
		}
		if stripRunTimestamp && key == "timestamp" && val.Len() > 0 && val.Bytes()[0] == '"' {
			fields[key] = nil
			continue
		}
		fields[key] = val.Bytes()
	}
	if end, err := dec.Token(); err != nil || end != json.Delim('}') {
		return errors.New("unterminated json object")
	}

	keys := make([]string, 0, len(fields))
	for k, v := range fields {
		if v != nil {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, err := json.Marshal(k)
		if err != nil {
			return err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(fields[k])
	}
	buf.WriteByte('}')
	return nil
}
