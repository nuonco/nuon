package auditexport

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/compute/metadata"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const gcpAuditExportSecretSuffix = "runner-audit-export"

type gcpSecretClient interface {
	AccessSecretVersion(context.Context, *secretmanagerpb.AccessSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

type gcpClientFactory interface {
	New(context.Context) (gcpSecretClient, string, error)
}

type gcpFactory struct{}

func newGCPFactory() gcpClientFactory { return gcpFactory{} }

func (gcpFactory) New(ctx context.Context) (gcpSecretClient, string, error) {
	projectID, err := metadata.ProjectIDWithContext(ctx)
	if err != nil {
		return nil, "", err
	}
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, "", err
	}
	return client, projectID, nil
}

type gcpConfigSource struct {
	factory  gcpClientFactory
	secretID string
}

func newGCPConfigSource(factory gcpClientFactory, installID string) configSource {
	return &gcpConfigSource{
		factory:  factory,
		secretID: installID + "-" + gcpAuditExportSecretSuffix,
	}
}

func (s *gcpConfigSource) Watch(ctx context.Context, interval time.Duration) <-chan configUpdate {
	var client gcpSecretClient
	var name string
	return watchConfig(ctx, interval, func() configUpdate {
		if client == nil {
			var err error
			var projectID string
			client, projectID, err = s.factory.New(ctx)
			if err != nil {
				client = nil
				return configUpdate{
					state:         configSourceInitializationFailed,
					err:           err,
					errorIdentity: gcpErrorIdentity(err),
				}
			}
			name = fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, s.secretID)
		}

		result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{Name: name})
		return newGCPConfigUpdate(result, err)
	}, func() {
		if client != nil {
			_ = client.Close()
		}
	})
}

func newGCPConfigUpdate(result *secretmanagerpb.AccessSecretVersionResponse, err error) configUpdate {
	if err != nil {
		state := configLookupFailed
		switch status.Code(err) {
		case codes.NotFound:
			state = configNotFound
		case codes.FailedPrecondition:
			state = configUnavailable
		}
		return configUpdate{state: state, err: err, errorIdentity: gcpErrorIdentity(err)}
	}
	if result == nil || result.Payload == nil {
		return configUpdate{state: configAvailable}
	}
	return configUpdate{state: configAvailable, value: string(result.Payload.Data)}
}

func gcpErrorIdentity(err error) string {
	if grpcStatus, ok := status.FromError(err); ok {
		return fmt.Sprintf("%T:%s:%s", err, grpcStatus.Code(), grpcStatus.Message())
	}
	return fmt.Sprintf("%T:%s", err, err.Error())
}
