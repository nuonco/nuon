package lookup

import (
	"context"

	"github.com/nuonco/nuon-go"
)

func AppID(ctx context.Context, apiClient nuon.Client, appIDOrName string) (string, error) {
	app, err := apiClient.GetApp(ctx, appIDOrName)
	if err != nil {
		return "", err
	}

	return app.ID, nil
}
