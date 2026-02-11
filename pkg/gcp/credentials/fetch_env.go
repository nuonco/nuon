package credentials

import (
	"context"
)

func FetchEnv(_ context.Context, cfg *Config) (map[string]string, error) {
	env := map[string]string{}
	if cfg.ProjectID != "" {
		env["GOOGLE_PROJECT"] = cfg.ProjectID
		env["CLOUDSDK_CORE_PROJECT"] = cfg.ProjectID
		env["GCLOUD_PROJECT"] = cfg.ProjectID
	}
	if cfg.Region != "" {
		env["GOOGLE_REGION"] = cfg.Region
		env["CLOUDSDK_COMPUTE_REGION"] = cfg.Region
	}
	return env, nil
}
