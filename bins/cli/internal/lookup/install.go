package lookup

import (
	"context"

	"github.com/nuonco/nuon-go"
)

func InstallID(ctx context.Context, apiClient nuon.Client, installIDOrName string) (string, error) {
	install, err := apiClient.GetInstall(ctx, installIDOrName)
	if err != nil {
		return "", err
	}

	return install.ID, nil
}
