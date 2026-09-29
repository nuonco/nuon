package credentials

import "context"

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
