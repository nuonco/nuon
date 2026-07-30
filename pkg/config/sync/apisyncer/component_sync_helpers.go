package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// healthProbeRequests maps declared health probes onto the create-config request.
func healthProbeRequests(health *config.ComponentHealthConfig) []*models.ServiceHealthProbeRequest {
	if health == nil || len(health.Probes) == 0 {
		return nil
	}

	out := make([]*models.ServiceHealthProbeRequest, 0, len(health.Probes))
	for _, probe := range health.Probes {
		out = append(out, &models.ServiceHealthProbeRequest{
			Type:    probe.Type,
			Name:    probe.Name,
			URL:     probe.URL,
			Command: probe.Command,
		})
	}
	return out
}

// shouldSkipBuildDueToChecksum checks if a component build should be skipped
// based on checksum comparison, considering the latest build status
func (s *syncer) shouldSkipBuildDueToChecksum(ctx context.Context, compID string, cmpChecksum componentChecksum) (bool, string, error) {
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
	if cmpChecksum.Equals(cmpLatestConfig.Checksum) {
		return true, cmpLatestConfig.ID, nil
	}

	return false, "", nil
}
