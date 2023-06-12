package s3

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

// backendConfig represents the full backend configuration
type backendConfig struct {
	*BucketConfig       `mapstructure:",squash"`
	*credentials.Config `mapstructure:",squash"`
}

func (b backendConfig) MarshalJSON() ([]byte, error) {
	var output map[string]interface{}

	if err := mapstructure.Decode(b.BucketConfig, &output); err != nil {
		return nil, fmt.Errorf("unable to decode bucket config to mapstructure: %w", err)
	}

	if err := mapstructure.Decode(b.Config, &output); err != nil {
		return nil, fmt.Errorf("unable to decode credentials config to mapstructure: %w", err)
	}

	byts, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("unable to convert mapstructure to json: %w", err)
	}

	return byts, nil
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
