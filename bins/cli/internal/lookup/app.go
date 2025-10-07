package lookup

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func AppID(ctx context.Context, apiClient nuon.Client, appIDOrName string) (string, error) {
	if appIDOrName == "" {
		return "", &ui.CLIUserError{
			Msg: "current app is not set, use apps select to set one or pass the --app-id flag",
		}
	}

	app, err := apiClient.GetApp(ctx, appIDOrName)
	if nuon.IsNotFound(err) {
		return "", &ui.CLIUserError{
			Msg: fmt.Sprintf("app \"%s\" not found", appIDOrName),
		}
	}

	if err != nil {
		return "", err
	}

	return app.ID, nil
}
