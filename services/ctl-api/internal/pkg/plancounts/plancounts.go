package plancounts

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var helmPlanPattern = regexp.MustCompile(`Plan:\s*(\d+)\s+to add,\s*(\d+)\s+to change,\s*(\d+)\s+to destroy`)

type terraformPlan struct {
	ResourceChanges []struct {
		Change struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}

type pulumiPlan struct {
	ChangeSummary   map[string]int `json:"change_summary"`
	ResourceChanges []struct {
		Action string `json:"action"`
	} `json:"resource_changes"`
}

type helmPlan struct {
	Plan        string          `json:"plan"`
	ContentDiff json.RawMessage `json:"helm_content_diff"`
}

type helmContentDiffEntry struct {
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
}

type kubernetesPlan struct {
	ContentDiff []struct {
		Op    string `json:"op"`
		Type  int    `json:"type"`
		Error string `json:"error"`
	} `json:"k8s_content_diff"`
}

func Counts(approvalType app.WorkflowStepApprovalType, plan string) (app.StepChangeCounts, app.StepChangeState, error) {
	if !Supported(approvalType) {
		return app.StepChangeCounts{}, app.StepChangeStateUnsupported, nil
	}

	if strings.TrimSpace(plan) == "" {
		return app.StepChangeCounts{}, app.StepChangeStateError, errors.New("plan is empty")
	}

	counts, err := parse(approvalType, strings.NewReader(plan))
	if err != nil {
		return app.StepChangeCounts{}, app.StepChangeStateError, err
	}

	return counts, app.StepChangeStateOK, nil
}

func Supported(approvalType app.WorkflowStepApprovalType) bool {
	switch approvalType {
	case app.TerraformPlanApprovalType,
		app.PulumiApprovalType,
		app.HelmApprovalApprovalType,
		app.KubernetesManifestApprovalType:
		return true
	default:
		return false
	}
}

func parse(approvalType app.WorkflowStepApprovalType, r io.Reader) (app.StepChangeCounts, error) {
	switch approvalType {
	case app.TerraformPlanApprovalType:
		return terraformCounts(r)
	case app.PulumiApprovalType:
		return pulumiCounts(r)
	case app.HelmApprovalApprovalType:
		return helmCounts(r)
	case app.KubernetesManifestApprovalType:
		return kubernetesCounts(r)
	default:
		return app.StepChangeCounts{}, nil
	}
}

func terraformCounts(r io.Reader) (app.StepChangeCounts, error) {
	var plan terraformPlan
	if err := json.NewDecoder(r).Decode(&plan); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse terraform plan")
	}

	var counts app.StepChangeCounts
	for _, resource := range plan.ResourceChanges {
		actions := resource.Change.Actions
		if containsAction(actions, "create") && containsAction(actions, "delete") {
			counts.Replace++
			continue
		}
		for _, action := range actions {
			switch action {
			case "create":
				counts.Create++
			case "update":
				counts.Update++
			case "delete":
				counts.Delete++
			case "no-op":
				counts.Noop++
			}
		}
	}
	return counts, nil
}

func pulumiCounts(r io.Reader) (app.StepChangeCounts, error) {
	var plan pulumiPlan
	if err := json.NewDecoder(r).Decode(&plan); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse pulumi plan")
	}

	if len(plan.ChangeSummary) > 0 {
		return app.StepChangeCounts{
			Create:  plan.ChangeSummary["create"],
			Update:  plan.ChangeSummary["update"],
			Delete:  plan.ChangeSummary["delete"],
			Replace: plan.ChangeSummary["replace"],
			Noop:    plan.ChangeSummary["same"] + plan.ChangeSummary["no-op"],
		}, nil
	}

	var counts app.StepChangeCounts
	for _, resource := range plan.ResourceChanges {
		switch resource.Action {
		case "create":
			counts.Create++
		case "update":
			counts.Update++
		case "delete":
			counts.Delete++
		case "replace":
			counts.Replace++
		case "same", "no-op":
			counts.Noop++
		}
	}
	return counts, nil
}

func helmCounts(r io.Reader) (app.StepChangeCounts, error) {
	var plan helmPlan
	if err := json.NewDecoder(r).Decode(&plan); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse helm plan")
	}

	if matches := helmPlanPattern.FindStringSubmatch(plan.Plan); len(matches) == 4 {
		create, _ := strconv.Atoi(matches[1])
		update, _ := strconv.Atoi(matches[2])
		deleted, _ := strconv.Atoi(matches[3])
		return app.StepChangeCounts{Create: create, Update: update, Delete: deleted}, nil
	}

	if len(plan.ContentDiff) == 0 {
		return app.StepChangeCounts{}, nil
	}

	var entries []helmContentDiffEntry
	if err := json.Unmarshal(plan.ContentDiff, &entries); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse helm content diff")
	}

	var counts app.StepChangeCounts
	for _, entry := range entries {
		hasBefore := hasJSONValue(entry.Before)
		hasAfter := hasJSONValue(entry.After)
		switch {
		case !hasBefore && hasAfter:
			counts.Create++
		case hasBefore && !hasAfter:
			counts.Delete++
		default:
			counts.Update++
		}
	}
	return counts, nil
}

func kubernetesCounts(r io.Reader) (app.StepChangeCounts, error) {
	var plan kubernetesPlan
	if err := json.NewDecoder(r).Decode(&plan); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse kubernetes plan")
	}

	var counts app.StepChangeCounts
	for _, resource := range plan.ContentDiff {
		if resource.Error != "" {
			continue
		}
		switch {
		case resource.Op == "delete" || resource.Type == 1:
			counts.Delete++
		case resource.Type == 2:
			counts.Create++
		case resource.Type == 3:
			counts.Update++
		case resource.Type == 0:
			counts.Noop++
		}
	}
	return counts, nil
}

func containsAction(actions []string, target string) bool {
	for _, action := range actions {
		if action == target {
			return true
		}
	}
	return false
}

func hasJSONValue(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	return len(trimmed) > 0 && !bytes.Equal(trimmed, []byte("null"))
}
