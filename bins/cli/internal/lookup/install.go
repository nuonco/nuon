package lookup

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

func InstallID(ctx context.Context, apiClient nuon.Client, installIDOrName string) (string, error) {
	if installIDOrName == "" {
		return "", &ui.CLIUserError{
			Msg: "current install is not set, use installs select to set one or pass the --install-id flag",
		}
	}

	install, err := apiClient.GetInstall(ctx, installIDOrName)
	if nuon.IsNotFound(err) {
		return "", &ui.CLIUserError{
			Msg: fmt.Sprintf("install \"%s\" not found", installIDOrName),
		}
	}
	if err != nil {
		return "", err
	}

	return install.ID, nil
}
