package credentials

import "context"

// FetchEnv returns environment variables needed for Terraform to operate in a GCP project.
// When ImpersonateServiceAccount is set, GOOGLE_IMPERSONATE_SERVICE_ACCOUNT is exported so the
// Terraform google provider calls iamcredentials.generateAccessToken on demand and handles token
// refresh automatically — avoiding the 1-hour static-token expiry problem.
func FetchEnv(_ context.Context, cfg *Config) (map[string]string, error) {
	if cfg == nil {
		return map[string]string{}, nil
	}
	env := map[string]string{
		"GOOGLE_PROJECT": cfg.ProjectID,
		"GOOGLE_REGION":  cfg.Region,
	}
	if cfg.ImpersonateServiceAccount != "" {
		env["GOOGLE_IMPERSONATE_SERVICE_ACCOUNT"] = cfg.ImpersonateServiceAccount
	}
	return env, nil
}
