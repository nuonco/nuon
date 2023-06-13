package s3

import (
	"context"
	"encoding/json"
	"fmt"
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
	if s.Credentials.AssumeRole != nil {
		cfg.RoleArn = s.Credentials.AssumeRole.RoleARN
		cfg.SessionName = s.Credentials.AssumeRole.SessionName
	}

	byts, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create config file: %w", err)
	}

	return byts, nil
}
