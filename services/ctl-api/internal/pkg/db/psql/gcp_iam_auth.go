package psql

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
)

func FetchGcpCloudSqlPassword(ctx context.Context, _ database) (string, error) {
	ts, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/sqlservice.login")
	if err != nil {
		return "", fmt.Errorf("failed to get GCP token source: %w", err)
	}

	token, err := ts.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get GCP access token: %w", err)
	}

	return token.AccessToken, nil
}
