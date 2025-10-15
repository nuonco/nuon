package lookup

import (
	"context"

	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func InstallID(ctx context.Context, apiClient nuon.Client, installIDOrName string) (string, error) {
	if installIDOrName == "" {
		return "", &ui.CLIUserError{
			Msg: "current install is not set, use installs select to set one or pass the --install-id flag",
		}
	}

	install, err := apiClient.GetInstall(ctx, installIDOrName)
	if err != nil {
		return "", err
	}

	return install.ID, nil
}
