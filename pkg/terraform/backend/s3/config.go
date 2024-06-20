package s3

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

// backendConfig represents the full backend configuration
type backendConfig struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Region string `json:"region"`

	// assume role creds
	RoleArn     string `json:"role_arn,omitempty"`
	SessionName string `json:"session_name,omitempty"`

	// static creds
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Token     string `json:"token,omitempty"`
}

func (s *s3) ConfigFile(ctx context.Context) ([]byte, error) {
	cfg := backendConfig{
		Bucket: s.Bucket.Name,
		Key:    s.Bucket.Key,
		Region: s.Bucket.Region,
	}

	if s.Credentials.Static != nil {
		cfg.AccessKey = s.Credentials.Static.AccessKeyID
		cfg.SecretKey = s.Credentials.Static.SecretAccessKey
		cfg.Token = s.Credentials.Static.SessionToken
	}

	// NOTE: we assume the credentials here, and write creds into the config file, so we can use the run-auth to set
	// the environment variable creds.
	if s.Credentials.AssumeRole != nil {
		awsCfg, err := credentials.Fetch(ctx, s.Credentials)
		if err != nil {
			return nil, fmt.Errorf("unable to get config: %w", err)
		}

		creds, err := awsCfg.Credentials.Retrieve(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve credentials from config: %w", err)
		}

		cfg.AccessKey = creds.AccessKeyID
		cfg.SecretKey = creds.SecretAccessKey
		cfg.Token = creds.SessionToken
	}

	byts, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create config file: %w", err)
	}

	return byts, nil
}
