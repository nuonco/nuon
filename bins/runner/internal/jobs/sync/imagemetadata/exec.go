package imagemetadata

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/oci"
	"github.com/nuonco/nuon/pkg/oci/metadata"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	plan := h.state.plan

	l.Info("fetching image metadata",
		zap.String("repository", plan.Registry.Repository),
		zap.String("tag", plan.Tag),
	)

	// Build fetch options
	fetchOpts := &metadata.FetchOptions{
		Image:                       plan.Registry.Repository,
		Tag:                         plan.Tag,
		IncludeIndex:                plan.IncludeIndex,
		IncludeAttestationManifests: plan.IncludeAttestationManifests,
		IncludeAttestationLayers:    plan.IncludeAttestationLayers,
	}

	// Get registry auth if configured
	accessInfo, err := oci.FetchAccessInfo(ctx, plan.Registry)
	if err != nil {
		h.errRecorder.Record("get registry auth", err)
		return errors.Wrap(err, "unable to get registry auth")
	}
	if accessInfo != nil && accessInfo.Auth != nil {
		fetchOpts.Auth = &metadata.RegistryAuth{
			ServerAddress: accessInfo.Auth.ServerAddress,
			Username:      accessInfo.Auth.Username,
			Password:      accessInfo.Auth.Password,
		}
	}

	// Fetch the metadata
	imgMetadata, err := metadata.FetchImageMetadata(ctx, fetchOpts)
	if err != nil {
		h.errRecorder.Record("fetch image metadata", err)
		return errors.Wrap(err, "unable to fetch image metadata")
	}
	h.state.metadata = imgMetadata

	l.Info("image metadata fetched successfully",
		zap.String("digest", imgMetadata.Digest),
		zap.Bool("signed", imgMetadata.Signed),
		zap.Bool("has_sbom", imgMetadata.SBOM != nil && imgMetadata.SBOM.Present),
		zap.Int("attestations_count", len(imgMetadata.Attestations)),
	)

	// Serialize metadata to JSON and gzip it
	metadataJSON, err := json.Marshal(imgMetadata)
	if err != nil {
		h.errRecorder.Record("marshal metadata", err)
		return errors.Wrap(err, "unable to marshal image metadata")
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(metadataJSON); err != nil {
		h.errRecorder.Record("gzip metadata", err)
		return errors.Wrap(err, "unable to gzip image metadata")
	}
	if err := gw.Close(); err != nil {
		h.errRecorder.Record("close gzip writer", err)
		return errors.Wrap(err, "unable to close gzip writer")
	}

	// Base64 encode the gzipped content for the API
	contentsCompressed := base64.URLEncoding.EncodeToString(buf.Bytes())

	// Write job execution result with compressed metadata
	resultReq := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:            true,
		ContentsCompressed: contentsCompressed,
	}
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}
