package credentials

import (
	"context"
	"fmt"
)

func FetchEnv(ctx context.Context, cfg *Config) (map[string]string, error) {
	awsCfg, err := Fetch(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to get credentials: %w", err)
	}

	if awsCfg.Credentials == nil {
		return nil, fmt.Errorf("no credentials were set on aws config")
	}

	awsCreds, err := awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve credentials: %w", err)
	}

	envVars := make(map[string]string)
	envVars["AWS_ACCESS_KEY_ID"] = awsCreds.AccessKeyID
	envVars["AWS_SECRET_ACCESS_KEY"] = awsCreds.SecretAccessKey
	envVars["AWS_SESSION_TOKEN"] = awsCreds.SessionToken

	return envVars, nil
}
