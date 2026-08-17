package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2run"
	"github.com/nuonco/nuon/pkg/runner/oci/bundle"
)

type s3CandidateBundleLoader struct {
	syncer       *airgapS3Sync
	deploymentID string
}

func (l *s3CandidateBundleLoader) Load(ctx context.Context, request day2.Request) (*day2run.CandidateBundle, error) {
	workdir, err := os.MkdirTemp("", "nuon-candidate-plan-*")
	if err != nil {
		return nil, err
	}
	cleanup := func() error { return os.RemoveAll(workdir) }
	fail := func(err error) (*day2run.CandidateBundle, error) {
		_ = cleanup()
		return nil, err
	}

	var candidate day2.BundleCandidate
	if err := l.getJSON(ctx, request.CandidateRecordKey, &candidate); err != nil {
		return fail(fmt.Errorf("read candidate record: %w", err))
	}
	if candidate.Bundle.BundleDigest != request.BundleDigest {
		return fail(fmt.Errorf("candidate record digest %q does not match request %q", candidate.Bundle.BundleDigest, request.BundleDigest))
	}
	if candidate.Bundle.DeploymentID != request.DeploymentID {
		return fail(fmt.Errorf("candidate record deployment %q does not match request %q", candidate.Bundle.DeploymentID, request.DeploymentID))
	}
	if candidate.Deployment == nil || candidate.Deployment.CandidateBundleKey != request.CandidateArchiveKey {
		return fail(fmt.Errorf("candidate archive key does not match immutable candidate record"))
	}

	archivePath := filepath.Join(workdir, "candidate.oci.tar.zst")
	if err := l.download(ctx, request.CandidateArchiveKey, archivePath); err != nil {
		return fail(err)
	}
	archive, err := os.Open(archivePath)
	if err != nil {
		return fail(err)
	}
	extractDir := filepath.Join(workdir, "bundle")
	checksum, extractErr := bundle.Extract(extractDir, archive)
	closeErr := archive.Close()
	if extractErr != nil {
		return fail(fmt.Errorf("extract candidate bundle: %w", extractErr))
	}
	if closeErr != nil {
		return fail(closeErr)
	}
	if expected := candidate.Bundle.ArchiveDigest; expected != "" && expected != "sha256:"+checksum {
		return fail(fmt.Errorf("candidate archive digest mismatch: got %q", "sha256:"+checksum))
	}
	if err := bundle.VerifyBlobs(extractDir); err != nil {
		return fail(fmt.Errorf("verify candidate bundle blobs: %w", err))
	}
	b, err := bundle.Open(ctx, extractDir)
	if err != nil {
		return fail(fmt.Errorf("open candidate bundle: %w", err))
	}
	envelope, err := airgap.Parse(b.PlanEnvelope)
	if err != nil {
		return fail(fmt.Errorf("parse candidate plan envelope: %w", err))
	}
	if day2.EnvelopeDigest(b.PlanEnvelope) != request.BundleDigest {
		return fail(fmt.Errorf("candidate envelope digest does not match request"))
	}
	if l.deploymentID != "" {
		if _, err := envelope.ApplyDeploymentID(l.deploymentID); err != nil {
			return fail(fmt.Errorf("apply candidate deployment ID: %w", err))
		}
	}
	if envelope.InstallID != request.DeploymentID {
		return fail(fmt.Errorf("candidate envelope deployment ID does not match request"))
	}
	source, err := airgap.NewBundleSource(b.Store(), b.Members(), b.Provenance)
	if err != nil {
		return fail(err)
	}
	if missing := source.MissingPlanSources(envelope); len(missing) > 0 {
		return fail(fmt.Errorf("candidate bundle has unpackaged plan sources: %v", missing))
	}
	if err := source.AddPlanAliases(envelope); err != nil {
		return fail(fmt.Errorf("resolve candidate plan sources: %w", err))
	}
	return &day2run.CandidateBundle{Envelope: envelope, Source: source, Close: cleanup}, nil
}

func (l *s3CandidateBundleLoader) getJSON(ctx context.Context, key string, dst any) error {
	raw, found, err := l.syncer.readControlObject(ctx, key)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("candidate record %q not found", key)
	}
	return json.Unmarshal(raw, dst)
}

func (l *s3CandidateBundleLoader) download(ctx context.Context, key, path string) error {
	out, err := l.syncer.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(l.syncer.bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("download candidate bundle: %w", err)
	}
	defer out.Body.Close()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := f.ReadFrom(out.Body)
	return firstError(copyErr, f.Close())
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
