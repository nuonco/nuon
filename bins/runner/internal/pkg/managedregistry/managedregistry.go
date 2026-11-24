// Package managedregistry fills OCIAuth on registries a plan marks ManagedCredential.
package managedregistry

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

// Resolve mints one credential and applies it to every managed registry in cfgs.
func Resolve(ctx context.Context, client nuonrunner.Client, jobID string, cfgs ...*configs.OCIRegistryRepository) error {
	var auth *configs.OCIRegistryAuth
	for _, cfg := range cfgs {
		if !needsCredential(cfg) {
			continue
		}
		if auth == nil {
			cred, err := client.CreateJobRegistryCredentials(ctx, jobID)
			if err != nil {
				return fmt.Errorf("unable to get managed registry credentials: %w", err)
			}
			if cred == nil || cred.Password == "" {
				return fmt.Errorf("managed registry credentials response was empty")
			}
			auth = &configs.OCIRegistryAuth{
				Username: cred.Username,
				Password: cred.Password,
			}
		}
		cfg.OCIAuth = auth
	}
	return nil
}

func needsCredential(cfg *configs.OCIRegistryRepository) bool {
	if cfg == nil || !cfg.ManagedCredential {
		return false
	}
	return cfg.OCIAuth == nil || cfg.OCIAuth.Password == ""
}
