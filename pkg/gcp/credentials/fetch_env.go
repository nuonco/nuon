package credentials

import "context"

// FetchEnv returns environment variables needed for Terraform to operate in a GCP project.
// GCP runners use an attached service account, so no credentials are needed — only project context.
func FetchEnv(ctx context.Context, cfg *Config) (map[string]string, error) {
	if cfg == nil {
		return map[string]string{}, nil
	}
	return map[string]string{
		"GOOGLE_PROJECT": cfg.ProjectID,
		"GOOGLE_REGION":  cfg.Region,
	}, nil
}
