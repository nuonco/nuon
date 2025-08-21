package sync

import (
	"context"

	"github.com/nuonco/nuon-go"
)

// shouldSkipBuildDueToChecksum checks if a component build should be skipped
// based on checksum comparison, considering the latest build status
func (s *sync) shouldSkipBuildDueToChecksum(ctx context.Context, compID, newChecksum string) (bool, string, error) {
	// Get the latest build to check its status
	cmpBuild, err := s.apiClient.GetComponentLatestBuild(ctx, compID)
	if err != nil {
		// if no build was found, attempt to build
		if nuon.IsNotFound(err) {
			return false, "", nil
		}

		return false, "", err
	}

	// if previous build failed, attempt to build again
	if cmpBuild.Status == "error" {
		return false, "", nil
	}

	// grab the latest config
	cmpLatestConfig, err := s.apiClient.GetComponentLatestConfig(ctx, compID)
	if err != nil {
		if nuon.IsNotFound(err) {
			return false, "", nil
		}

		return false, "", err
	}

	// if the new checksum equals the old one, skip
	if cmpLatestConfig.Checksum == newChecksum {
		return true, cmpLatestConfig.ID, nil
	}

	return false, "", nil
}
