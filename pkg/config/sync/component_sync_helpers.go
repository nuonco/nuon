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
	if err != nil && !nuon.IsNotFound(err) {
		return false, "", err
	}

	// Don't do checksum comparison if the latest build failed
	doChecksumCompare := true
	if cmpBuild != nil && cmpBuild.Status == "error" {
		doChecksumCompare = false
	}

	cmpLatestConfig, err := s.apiClient.GetComponentLatestConfig(ctx, compID)
	if err != nil {
		return false, "", err
	}

	if cmpLatestConfig == nil {
		return false, "", nil
	}

	if doChecksumCompare {
		if err != nil {
			return false, "", err
		}
		if cmpLatestConfig.Checksum == newChecksum {
			return true, cmpLatestConfig.ID, nil
		}
	}

	return false, "", nil
}
