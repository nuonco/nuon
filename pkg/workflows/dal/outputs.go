package dal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/mono/pkg/aws/s3downloader"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	"google.golang.org/protobuf/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

const (
	defaultTerraformOutputFilename string = "output-struct-v1.pb"
)

func (r *client) GetInstanceOutputs(ctx context.Context, orgID, appID, componentID, installID string) (*structpb.Struct, error) {
	creds := r.deploymentsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.DeploymentsBucket, s3downloader.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	bucketKey := filepath.Join(
		prefix.InstanceStatePath(orgID, appID, componentID, installID),
		defaultTerraformOutputFilename,
	)
	outputs, err := r.getOutputs(ctx, client, bucketKey)
	if err != nil {
		return nil, fmt.Errorf("unable to get outputs: %w", err)
	}

	return outputs, nil
}

func (r *client) getOutputs(ctx context.Context, client s3downloader.Downloader, key string) (*structpb.Struct, error) {
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	outputs := &structpb.Struct{}
	if err := proto.Unmarshal(byts, outputs); err != nil {
		return nil, fmt.Errorf("invalid outputs: %w", err)
	}

	return outputs, nil
}
