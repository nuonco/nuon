package activities

import (
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func ResolveInstallComponentID(ctx workflow.Context, installComponentID, installID, componentID string) (string, error) {
	if installComponentID != "" {
		return installComponentID, nil
	}
	if installID == "" || componentID == "" {
		return "", temporal.NewNonRetryableApplicationError("install component id, or install and component id, required", "InvalidArgument", nil)
	}

	installComp, err := AwaitGetInstallComponent(ctx, GetInstallComponentRequest{
		InstallID:   installID,
		ComponentID: componentID,
	})
	if err != nil {
		return "", err
	}
	if installComp == nil {
		return "", fmt.Errorf("install component not found for install %s and component %s", installID, componentID)
	}
	return installComp.ID, nil
}
