package s3

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

// backendConfig represents the full backend configuration
type backendConfig struct {
	*BucketConfig
	*credentials.Config
}

func (s *s3) ConfigFile(ctx context.Context) ([]byte, error) {
	byts, err := json.Marshal(backendConfig{
		s.Bucket,
		s.Credentials,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create config file: %w", err)
	}

	return byts, nil
}
