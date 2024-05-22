package lookup

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
)

func ComponentID(ctx context.Context, apiClient nuon.Client, appID string, compIDOrName string) (string, error) {
	if appID == "" {
		return "", fmt.Errorf("app must be set using nuon apps select first")
	}

	app, err := apiClient.GetApp(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("unable to lookup app: %w", err)
	}

	appID = app.ID
	appComps, err := apiClient.GetAppComponents(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("unable to fetch app: %w", err)
	}

	for _, comp := range appComps {
		if comp.ID == compIDOrName || comp.Name == compIDOrName {
			return comp.ID, nil
		}
	}

	return "", fmt.Errorf("Make sure app is set correctly")
}
