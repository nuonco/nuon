package lookup

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func InstallID(ctx context.Context, apiClient nuon.Client, installIDOrName string) (string, error) {
	install, err := Install(ctx, apiClient, installIDOrName)
	if err != nil {
		return "", err
	}

	return install.ID, nil
}

func Install(ctx context.Context, apiClient nuon.Client, installIDOrName string) (*models.AppInstall, error) {
	if installIDOrName == "" {
		return nil, &ui.CLIUserError{
			Msg: "current install is not set, use installs select to set one or pass the --install-id flag",
		}
	}

	return apiClient.GetInstall(ctx, installIDOrName)
}
