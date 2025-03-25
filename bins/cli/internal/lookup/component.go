package lookup

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func ComponentID(ctx context.Context, apiClient nuon.Client, appID string, compIDOrName string) (string, error) {
	if appID == "" {
		return "", &ui.CLIUserError{
			Msg: "app must be set using nuon apps select first",
		}
	}

	app, err := apiClient.GetApp(ctx, appID)
	if err != nil {
		return "", &ui.CLIUserError{
			Msg: "unable to lookup app id",
		}
	}

	appID = app.ID
	appComp, err := apiClient.GetAppComponent(ctx, appID, compIDOrName)

	if err != nil {
		return "", &ui.CLIUserError{
			Msg: fmt.Sprintf("component id or name is not valid: %s", compIDOrName),
		}
	}

	return appComp.ID, nil
}
