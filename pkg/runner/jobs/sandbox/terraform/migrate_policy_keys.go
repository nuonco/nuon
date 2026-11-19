package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/nuonco/nuon/pkg/policies"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
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
// contains no legacy keys and subsequent runs are a state-read no-op.
func (h *handler) migrateLegacyPolicyKeys(ctx context.Context, log hclog.Logger, ws workspace.Workspace) error {
	zl, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// Show requires current provider schemas before plan has upgraded older state.
	state, err := ws.StatePull(ctx, log)
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
func findLegacyPolicyKeyMigrations(rawState string) ([]policyKeyMigration, error) {
	if strings.TrimSpace(rawState) == "" {
		return nil, nil
	}

	var state struct {
		Version   int `json:"version"`
		Resources []struct {
			Module    string `json:"module"`
			Mode      string `json:"mode"`
			Type      string `json:"type"`
			Name      string `json:"name"`
			Instances []struct {
				IndexKey   any             `json:"index_key"`
				Deposed    string          `json:"deposed"`
				Attributes json.RawMessage `json:"attributes"`
			} `json:"instances"`
		} `json:"resources"`
	}
	if err := json.Unmarshal([]byte(rawState), &state); err != nil {
		return nil, fmt.Errorf("unable to decode state for policy key migration: %w", err)
	}
	if state.Version != 4 {
		return nil, fmt.Errorf("unsupported state version %d for policy key migration", state.Version)
	}

	var mvs []policyKeyMigration
	for _, r := range state.Resources {
		if r.Mode != "managed" || r.Type != vendorPoliciesResourceType || r.Name != vendorPoliciesResourceName {
			continue
		}
		address := r.Type + "." + r.Name
		if r.Module != "" {
			address = r.Module + "." + address
		}
		for _, instance := range r.Instances {
			key, ok := instance.IndexKey.(string)
			if !ok || !policies.IsLegacyKey(key) || instance.Deposed != "" {
				continue
			}
			source := fmt.Sprintf("%s[%q]", address, key)
			var attributes struct {
				YAMLBody string `json:"yaml_body"`
			}
			if len(instance.Attributes) > 0 {
				if err := json.Unmarshal(instance.Attributes, &attributes); err != nil {
					return nil, fmt.Errorf("unable to decode policy attributes for %s: %w", source, err)
				}
			}
			if attributes.YAMLBody == "" {
				return nil, fmt.Errorf("resource %s has legacy key %q but no yaml_body attribute", source, key)
			}
			newKey, err := policies.ManifestKeyFromYAML(attributes.YAMLBody)
			if err != nil {
				return nil, fmt.Errorf("unable to derive new key for %s: %w", source, err)
			}
			kind, name := manifestKindAndName(attributes.YAMLBody)
			mvs = append(mvs, policyKeyMigration{
				sourceAddress:      source,
				destinationAddress: fmt.Sprintf("%s[%q]", address, newKey),
				oldKey:             key,
				newKey:             newKey,
				kind:               kind,
				name:               name,
			})
		}
	}
	return mvs, nil
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
