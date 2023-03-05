package waypoint

import (
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

func DefaultLabels(meta *planv1.Metadata, componentID, phaseName string) map[string]string {
	labels := map[string]string{
		"deployment-id": meta.DeploymentShortId,
		"app-id":        meta.AppShortId,
		"org-id":        meta.OrgShortId,
		"component-id":  componentID,
		"phase":         phaseName,
	}
	if meta.InstallShortId != "" {
		labels["install-id"] = meta.InstallShortId
	}

	return labels
}
