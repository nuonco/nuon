package terraform

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-hclog"
	tfjson "github.com/hashicorp/terraform-json"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/pkg/policies"
	"github.com/nuonco/nuon/pkg/terraform/workspace"
)

// vendorPoliciesResourceType is the resource type used by the sandbox TF module
// for Kyverno vendor policies.
const vendorPoliciesResourceType = "kubectl_manifest"

// vendorPoliciesResourceName is the resource name used by the sandbox TF module.
const vendorPoliciesResourceName = "vendor_policies"

// migrateLegacyPolicyKeys renames legacy positional state keys (`N.yaml`) to
// content-derived keys without touching the underlying K8s objects. Intended
// to run between `terraform init` and `terraform plan`, after which plan is
// a no-op for the migrated entries.
//
// It is safe to run on every sandbox apply: after migration completes, state
// contains no legacy keys and subsequent runs are a quick state-list no-op.
func (h *handler) migrateLegacyPolicyKeys(ctx context.Context, log hclog.Logger, ws workspace.Workspace) error {
	zl, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	state, err := ws.Show(ctx, log)
	if err != nil {
		return fmt.Errorf("unable to read state for policy key migration: %w", err)
	}

	mvs, err := findLegacyPolicyKeyMigrations(state)
	if err != nil {
		return err
	}

	tags := []string{
		fmt.Sprintf("has_legacy_keys:%t", len(mvs) > 0),
	}
	if h.mw != nil {
		h.mw.Incr("nuon_sandbox_policy_runs_total", tags)
	}

	if len(mvs) == 0 {
		zl.Debug("no legacy policy keys to migrate")
		return nil
	}

	zl.Info("migrating legacy sandbox policy keys", zap.Int("count", len(mvs)))
	for _, m := range mvs {
		zl.Info("migrating policy state key",
			zap.String("old_key", m.oldKey),
			zap.String("new_key", m.newKey),
			zap.String("kind", m.kind),
			zap.String("name", m.name),
		)
		if err := ws.StateMv(ctx, log, m.sourceAddress, m.destinationAddress); err != nil {
			return fmt.Errorf("unable to migrate policy state key %q -> %q: %w", m.oldKey, m.newKey, err)
		}
	}

	return nil
}

type policyKeyMigration struct {
	sourceAddress      string
	destinationAddress string
	oldKey             string
	newKey             string
	kind               string
	name               string
}

// findLegacyPolicyKeyMigrations walks the state for vendor_policies resources
// with legacy positional keys, parses their yaml_body to derive the new
// content-derived key, and returns the list of state-mv operations to perform.
func findLegacyPolicyKeyMigrations(state *tfjson.State) ([]policyKeyMigration, error) {
	if state == nil || state.Values == nil || state.Values.RootModule == nil {
		return nil, nil
	}

	var mvs []policyKeyMigration
	if err := walkModuleForPolicyMigrations(state.Values.RootModule, &mvs); err != nil {
		return nil, err
	}
	return mvs, nil
}

func walkModuleForPolicyMigrations(mod *tfjson.StateModule, out *[]policyKeyMigration) error {
	if mod == nil {
		return nil
	}

	for _, r := range mod.Resources {
		if r.Type != vendorPoliciesResourceType || r.Name != vendorPoliciesResourceName {
			continue
		}
		key, ok := r.Index.(string)
		if !ok || !policies.IsLegacyKey(key) {
			continue
		}
		yamlBody, _ := r.AttributeValues["yaml_body"].(string)
		if yamlBody == "" {
			return fmt.Errorf("resource %s has legacy key %q but no yaml_body attribute", r.Address, key)
		}
		newKey, err := policies.ManifestKeyFromYAML(yamlBody)
		if err != nil {
			return fmt.Errorf("unable to derive new key for %s: %w", r.Address, err)
		}
		if newKey == key {
			continue
		}
		kind, name := manifestKindAndName(yamlBody)
		*out = append(*out, policyKeyMigration{
			sourceAddress:      r.Address,
			destinationAddress: replaceIndexInAddress(r.Address, key, newKey),
			oldKey:             key,
			newKey:             newKey,
			kind:               kind,
			name:               name,
		})
	}

	for _, child := range mod.ChildModules {
		if err := walkModuleForPolicyMigrations(child, out); err != nil {
			return err
		}
	}
	return nil
}

// replaceIndexInAddress swaps the `[...]` suffix of a TF state address with a
// new key. Example: `module.x.kubectl_manifest.vendor_policies["0.yaml"]` ->
// `module.x.kubectl_manifest.vendor_policies["clusterrole-foo.yaml"]`.
func replaceIndexInAddress(address, oldKey, newKey string) string {
	oldSuffix := fmt.Sprintf("[%q]", oldKey)
	newSuffix := fmt.Sprintf("[%q]", newKey)
	if strings.HasSuffix(address, oldSuffix) {
		return strings.TrimSuffix(address, oldSuffix) + newSuffix
	}
	// tfjson typically quotes with double-quotes; fall back to a rough replace
	// if the exact form isn't present (defensive, should not happen in practice).
	return address[:strings.LastIndex(address, "[")] + newSuffix
}

// manifestKindAndName is a best-effort extraction of kind/metadata.name for
// structured logging. Errors are swallowed: if we got this far, ManifestKey
// already succeeded, so values are present.
func manifestKindAndName(yamlBody string) (string, string) {
	var m map[string]any
	if err := yaml.Unmarshal([]byte(yamlBody), &m); err != nil {
		return "", ""
	}
	kind, _ := m["kind"].(string)
	md, _ := m["metadata"].(map[string]any)
	if md == nil {
		if mdAny, ok := m["metadata"].(map[any]any); ok {
			md = make(map[string]any, len(mdAny))
			for k, v := range mdAny {
				if ks, ok := k.(string); ok {
					md[ks] = v
				}
			}
		}
	}
	name, _ := md["name"].(string)
	return kind, name
}
