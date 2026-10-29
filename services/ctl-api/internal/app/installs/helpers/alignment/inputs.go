package alignment

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

var alignmentInputWorkflowTypes = []app.WorkflowType{
	app.WorkflowTypeInputUpdate,
	app.WorkflowTypeComponentEnabled,
	app.WorkflowTypeComponentDisabled,
}

func (st *installAlignmentState) inputWorkflowTargets(wf *app.Workflow) map[string]bool {
	out := map[string]bool{}
	if wf.IsInputsOnly() {
		return out
	}
	if id := wf.Metadata["resolve_drift_component_id"]; id != nil && *id != "" {
		for _, ic := range st.components {
			if ic.ComponentID == *id {
				out[ic.ID] = true
			}
		}
		return out
	}
	raw := wf.Metadata["inputs"]
	if raw == nil || *raw == "" {
		return out
	}
	changed := strings.Split(*raw, ",")
	intersects := func(names []string) bool {
		return slices.ContainsFunc(names, func(name string) bool { return slices.Contains(changed, name) })
	}
	sandboxChanged := st.expectedSandbox != nil && intersects(installhelpers.InputNames(st.expectedSandbox.Refs))
	if sandboxChanged && st.sandbox != nil {
		out[st.sandbox.ID] = true
	}
	var componentIDs []string
	for id, target := range st.targets {
		if sandboxChanged || intersects(installhelpers.ComponentInputNames(target.Config, st.desiredInputs(""))) {
			componentIDs = append(componentIDs, id)
		}
	}
	if dependents := wf.Metadata["deploy_dependents"]; dependents != nil && *dependents == "true" {
		componentIDs = st.enablement.TransitiveDependentsClosure(componentIDs)
	}
	for _, ic := range st.components {
		if slices.Contains(componentIDs, ic.ComponentID) {
			out[ic.ID] = true
		}
	}
	return out
}

func (s *Helpers) addInputWorkflowAttempts(ctx context.Context, db *gorm.DB, st *installAlignmentState, ev *alignmentEvidence) error {
	latest := db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (changed.name, COALESCE(w.metadata->'resolve_drift_component_id', '')) w.id
		FROM install_workflows w
		CROSS JOIN LATERAL unnest(string_to_array(w.metadata->'inputs', ',')) AS changed(name)
		WHERE w.deleted_at = 0 AND w.owner_id = ? AND w.owner_type = 'installs'
		  AND w.org_id = ? AND w.plan_only = false AND w.type IN ?
		ORDER BY changed.name, COALESCE(w.metadata->'resolve_drift_component_id', ''), w.created_at DESC, w.id DESC`,
		st.install.ID, st.install.OrgID, alignmentInputWorkflowTypes)
	var workflows []app.Workflow
	if err := db.WithContext(ctx).Select("id", "type", "status", "metadata", "created_at").
		Where("id IN (?)", latest).Find(&workflows).Error; err != nil {
		return fmt.Errorf("unable to get input update attempts: %w", err)
	}
	for i := range workflows {
		wf := &workflows[i]
		for id := range st.inputWorkflowTargets(wf) {
			if last := ev.lastAttempts[id]; last != nil && !wf.CreatedAt.After(last.CreatedAt) {
				continue
			}
			ev.lastAttempts[id] = &InstallAlignmentAttempt{
				ID: wf.ID, WorkflowID: wf.ID, Type: wf.Type,
				Status: string(wf.Status.Status), CreatedAt: wf.CreatedAt,
			}
		}
	}
	return nil
}
