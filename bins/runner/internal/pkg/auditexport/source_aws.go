package auditexport

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
)

type awsSecretClient interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type awsClientFactory interface {
	New(context.Context) (awsSecretClient, error)
}

type awsFactory struct{}

func newAWSFactory() awsClientFactory { return awsFactory{} }

func (awsFactory) New(ctx context.Context) (awsSecretClient, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	if cfg.Region == "" {
		region, regionErr := imds.NewFromConfig(cfg).GetRegion(ctx, nil)
		if regionErr != nil {
			return nil, regionErr
		}
		cfg.Region = region.Region
	}
	return secretsmanager.NewFromConfig(cfg), nil
}

type sourceResolver struct {
	awsFactory awsClientFactory
}

func newConfigSourceResolver(factory awsClientFactory) configSourceResolver {
	return &sourceResolver{awsFactory: factory}
}

func (r *sourceResolver) Resolve(platform, installID string) configSource {
	if installID == "" || !strings.HasPrefix(strings.ToLower(platform), "aws") {
		return nil
	}
	return &awsConfigSource{
		factory: r.awsFactory,
		name:    "nuon/" + installID + "/runner-audit-export",
	}
}

type awsConfigSource struct {
	factory awsClientFactory
	name    string
}

func (s *awsConfigSource) Watch(ctx context.Context, interval time.Duration) <-chan configUpdate {
	updates := make(chan configUpdate)
	go func() {
		defer close(updates)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var client awsSecretClient
		var observation configObservation
		refresh := func() bool {
			if client == nil {
				var err error
				client, err = s.factory.New(ctx)
				if err != nil {
					client = nil
					return publishConfigUpdate(ctx, updates, &observation, configUpdate{
						state:         configSourceInitializationFailed,
						err:           err,
						errorIdentity: awsErrorIdentity(err),
					})
				}
			}

			result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &s.name})
			return publishConfigUpdate(ctx, updates, &observation, newAWSConfigUpdate(result, err))
		}

		if !refresh() {
			return
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !refresh() {
					return
				}
			}
		}
	}()
	return updates
}

func newAWSConfigUpdate(result *secretsmanager.GetSecretValueOutput, err error) configUpdate {
	if err != nil {
		var notFound *types.ResourceNotFoundException
		var invalidRequest *types.InvalidRequestException
		switch {
		case errors.As(err, &notFound):
			return configUpdate{state: configNotFound, err: err, errorIdentity: awsErrorIdentity(err)}
		case errors.As(err, &invalidRequest):
			return configUpdate{state: configUnavailable, err: err, errorIdentity: awsErrorIdentity(err)}
		default:
			return configUpdate{state: configLookupFailed, err: err, errorIdentity: awsErrorIdentity(err)}
		}
	}
	if result == nil || result.SecretString == nil {
		return configUpdate{state: configAvailable}
	}
	return configUpdate{state: configAvailable, value: *result.SecretString}
}

func awsErrorIdentity(err error) string {
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		return fmt.Sprintf("%T:%s:%s", apiError, apiError.ErrorCode(), apiError.ErrorMessage())
	}
	return fmt.Sprintf("%T:%s", err, err.Error())
}
