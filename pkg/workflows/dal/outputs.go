package dal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/mono/pkg/aws/s3downloader"
	plugincomponentv1 "github.com/powertoolsdev/mono/pkg/types/plugins/component/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	"google.golang.org/protobuf/proto"
)

const (
	defaultTerraformOutputFilename string = "output-nuon.pb"
)

func (r *client) GetInstanceOutputs(ctx context.Context, orgID, appID, componentID, installID string) (*plugincomponentv1.Outputs, error) {
	creds := r.deploymentsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.InstallsBucket, s3downloader.WithCredentials(creds))
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

func (r *client) getOutputs(ctx context.Context, client s3downloader.Downloader, key string) (*plugincomponentv1.Outputs, error) {
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	outputs := &plugincomponentv1.Outputs{}
	if err := proto.Unmarshal(byts, outputs); err != nil {
		return nil, fmt.Errorf("invalid outputs: %w", err)
	}

	return outputs, nil
}
