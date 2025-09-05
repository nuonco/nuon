package lookup

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func AppID(ctx context.Context, apiClient nuon.Client, appIDOrName string) (string, error) {
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
