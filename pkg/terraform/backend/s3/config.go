package s3

import (
	"context"
	"encoding/json"
	"fmt"
)

// backendConfig represents the full backend configuration
type backendConfig struct {
	*BucketConfig
	*Credentials
	*IAMConfig
}

func (s *s3) GetConfigFile(ctx context.Context) ([]byte, error) {
	byts, err := json.Marshal(backendConfig{
		s.Bucket,
		s.Credentials,
		s.IAM,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create config file: %w", err)
	}

	return byts, nil
}
