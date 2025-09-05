package lookup

import (
	"context"

	"github.com/nuonco/nuon-go"
)

func VCSConnectionID(ctx context.Context, apiClient nuon.Client, connID string) (string, error) {
	conn, err := apiClient.GetVCSConnection(ctx, connID)
	if err != nil {
		return "", err
	}

	return conn.ID, nil
}
