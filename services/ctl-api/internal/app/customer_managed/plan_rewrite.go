package customermanaged

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// BundlePins identifies packaged builds used to replace historical references in exported plans.
type BundlePins struct {
	SandboxBuildID string
	// SandboxRegistry is recorded on the rewritten sandbox oci_source. An
	// customer-managed runner serves the source from the bundle by tag and never
	// contacts the registry; keeping the management registry here means any
	// non-bundle fallback fails loudly instead of pulling the wrong content.
	SandboxRegistry   *configs.OCIRegistryRepository
	ComponentBuildIDs map[string]string
}

// RewriteEnvelopeForBundle replaces historical sources with packaged builds while preserving destination tags required by deploy plans.
func RewriteEnvelopeForBundle(ctx context.Context, db *gorm.DB, envelope *customermanaged.Envelope, pins BundlePins) error {
	componentByBuild := map[string]string{}
	for i := range envelope.Steps {
		step := &envelope.Steps[i]
		var plan map[string]json.RawMessage
		if err := json.Unmarshal(step.CompositePlan, &plan); err != nil {
			return fmt.Errorf("step %s: decode composite plan: %w", step.ID, err)
		}

		changed := false
		if raw, ok := plan["sandbox_run_plan"]; ok && !isJSONNull(raw) {
			rewritten, err := rewriteSandboxRunPlan(step.ID, raw, pins)
			if err != nil {
				return err
			}
			plan["sandbox_run_plan"] = rewritten
			changed = true
		}
		if raw, ok := plan["sync_oci_plan"]; ok && !isJSONNull(raw) {
			rewritten, syncChanged, err := rewriteSyncOCIPlan(ctx, db, envelope.OrgID, step.ID, raw, pins, componentByBuild)
			if err != nil {
				return err
			}
			if syncChanged {
				plan["sync_oci_plan"] = rewritten
				changed = true
			}
		}
		if !changed {
			continue
		}

		updated, err := json.Marshal(plan)
		if err != nil {
			return fmt.Errorf("step %s: encode rewritten composite plan: %w", step.ID, err)
		}
		step.CompositePlan = updated
	}
	if err := envelope.Validate(); err != nil {
		return fmt.Errorf("rewritten plan envelope is invalid: %w", err)
	}
	return nil
}

func rewriteSandboxRunPlan(stepID string, raw json.RawMessage, pins BundlePins) (json.RawMessage, error) {
	if pins.SandboxBuildID == "" {
		return nil, fmt.Errorf("step %s: bundle has no pinned sandbox build to rewrite the sandbox plan source to", stepID)
	}
	var sandboxPlan map[string]json.RawMessage
	if err := json.Unmarshal(raw, &sandboxPlan); err != nil {
		return nil, fmt.Errorf("step %s: decode sandbox run plan: %w", stepID, err)
	}
	ociSource, err := json.Marshal(plantypes.OCISource{Registry: pins.SandboxRegistry, Tag: pins.SandboxBuildID})
	if err != nil {
		return nil, fmt.Errorf("step %s: encode sandbox oci source: %w", stepID, err)
	}
	sandboxPlan["oci_source"] = ociSource
	sandboxPlan["git_source"] = json.RawMessage("null")
	rewritten, err := json.Marshal(sandboxPlan)
	if err != nil {
		return nil, fmt.Errorf("step %s: encode sandbox run plan: %w", stepID, err)
	}
	return rewritten, nil
}

func rewriteSyncOCIPlan(ctx context.Context, db *gorm.DB, orgID, stepID string, raw json.RawMessage, pins BundlePins, componentByBuild map[string]string) (json.RawMessage, bool, error) {
	var syncPlan map[string]json.RawMessage
	if err := json.Unmarshal(raw, &syncPlan); err != nil {
		return nil, false, fmt.Errorf("step %s: decode sync-oci plan: %w", stepID, err)
	}
	var srcTag string
	if rawTag, ok := syncPlan["src_tag"]; ok {
		if err := json.Unmarshal(rawTag, &srcTag); err != nil {
			return nil, false, fmt.Errorf("step %s: decode sync-oci src_tag: %w", stepID, err)
		}
	}
	if srcTag == "" {
		return nil, false, fmt.Errorf("step %s: sync-oci plan has no src_tag", stepID)
	}
	for _, pinned := range pins.ComponentBuildIDs {
		if srcTag == pinned {
			return raw, false, nil
		}
	}

	componentID, err := componentIDForBuild(ctx, db, orgID, srcTag, componentByBuild)
	if err != nil {
		return nil, false, fmt.Errorf("step %s: %w", stepID, err)
	}
	pinned := pins.ComponentBuildIDs[componentID]
	if pinned == "" {
		return nil, false, fmt.Errorf("step %s: sync-oci source %s belongs to component %s which has no pinned build in the bundle", stepID, srcTag, componentID)
	}
	newTag, err := json.Marshal(pinned)
	if err != nil {
		return nil, false, fmt.Errorf("step %s: encode sync-oci src_tag: %w", stepID, err)
	}
	syncPlan["src_tag"] = newTag
	rewritten, err := json.Marshal(syncPlan)
	if err != nil {
		return nil, false, fmt.Errorf("step %s: encode sync-oci plan: %w", stepID, err)
	}
	return rewritten, true, nil
}

func componentIDForBuild(ctx context.Context, db *gorm.DB, orgID, buildID string, cache map[string]string) (string, error) {
	if componentID, ok := cache[buildID]; ok {
		return componentID, nil
	}
	var build app.ComponentBuild
	if err := db.WithContext(ctx).Where(app.ComponentBuild{ID: buildID, OrgID: orgID}).First(&build).Error; err != nil {
		return "", fmt.Errorf("sync-oci source tag %s is not a component build in this org; portable bundles can only serve component-build sources: %w", buildID, err)
	}
	var connection app.ComponentConfigConnection
	if err := db.WithContext(ctx).Where(app.ComponentConfigConnection{ID: build.ComponentConfigConnectionID}).First(&connection).Error; err != nil {
		return "", fmt.Errorf("load component config connection %s for build %s: %w", build.ComponentConfigConnectionID, buildID, err)
	}
	cache[buildID] = connection.ComponentID
	return connection.ComponentID, nil
}

func isJSONNull(raw json.RawMessage) bool {
	trimmed := string(raw)
	return trimmed == "" || trimmed == "null"
}
