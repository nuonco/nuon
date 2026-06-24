package terraform

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-exec/tfexec"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// tfOutString / tfOutStrMap decode a single `terraform output` entry into dst.
// A missing key is a no-op (dst keeps its zero value); a present null decodes
// cleanly (the module emits null for service accounts it didn't create).
func tfOutString(meta map[string]tfexec.OutputMeta, key string, dst *string) error {
	m, ok := meta[key]
	if !ok {
		return nil
	}
	if err := json.Unmarshal(m.Value, dst); err != nil {
		return fmt.Errorf("output %q: %w", key, err)
	}
	return nil
}

func tfOutStrMap(meta map[string]tfexec.OutputMeta, key string, dst *map[string]string) error {
	m, ok := meta[key]
	if !ok {
		return nil
	}
	if err := json.Unmarshal(m.Value, dst); err != nil {
		return fmt.Errorf("output %q: %w", key, err)
	}
	return nil
}

// gcpOutputsToCore maps the module's `terraform output` result into core.Outputs.
// Output names mirror install-stacks/gcp/outputs.tf (which mirror the gcp
// phone-home payload), so this is a direct key-by-key translation.
func gcpOutputsToCore(meta map[string]tfexec.OutputMeta) (*core.Outputs, error) {
	g := &core.GCPOutputs{
		BreakGlassSAEmails:    map[string]string{},
		BreakGlassSAUniqueIDs: map[string]string{},
		CustomSAEmails:        map[string]string{},
		CustomSAUniqueIDs:     map[string]string{},
		SecretNames:           map[string]string{},
	}
	out := &core.Outputs{
		Cloud:         core.CloudGCP,
		InstallInputs: map[string]string{},
		GCP:           g,
	}

	strs := map[string]*string{
		"project_id":                       &g.ProjectID,
		"region":                           &g.Region,
		"network_name":                     &g.NetworkName,
		"network_id":                       &g.NetworkID,
		"public_subnet_name":               &g.PublicSubnetName,
		"private_subnet_name":              &g.PrivateSubnetName,
		"runner_subnet_name":               &g.RunnerSubnetName,
		"runner_service_account_email":     &g.RunnerSAEmail,
		"runner_service_account_unique_id": &g.RunnerSAUniqueID,
		"gke_node_pool_sa_email":           &g.GKENodePoolSAEmail,
		"gke_node_pool_sa_unique_id":       &g.GKENodePoolSAUniqueID,
		"provision_sa_email":               &g.ProvisionSAEmail,
		"provision_sa_unique_id":           &g.ProvisionSAUniqueID,
		"maintenance_sa_email":             &g.MaintenanceSAEmail,
		"maintenance_sa_unique_id":         &g.MaintenanceSAUniqueID,
		"deprovision_sa_email":             &g.DeprovisionSAEmail,
		"deprovision_sa_unique_id":         &g.DeprovisionSAUniqueID,
	}
	for key, dst := range strs {
		if err := tfOutString(meta, key, dst); err != nil {
			return nil, fmt.Errorf("decode terraform outputs (%s): %w", key, err)
		}
	}

	maps := map[string]*map[string]string{
		"break_glass_sa_emails":     &g.BreakGlassSAEmails,
		"break_glass_sa_unique_ids": &g.BreakGlassSAUniqueIDs,
		"custom_sa_emails":          &g.CustomSAEmails,
		"custom_sa_unique_ids":      &g.CustomSAUniqueIDs,
		"secret_names":              &g.SecretNames,
		"install_inputs":            &out.InstallInputs,
	}
	for key, dst := range maps {
		if err := tfOutStrMap(meta, key, dst); err != nil {
			return nil, fmt.Errorf("decode terraform outputs (%s): %w", key, err)
		}
	}

	return out, nil
}
