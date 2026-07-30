package auditexport

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
)

type secretClient interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type clientFactory interface {
	New(context.Context) (secretClient, error)
}

type awsFactory struct{}

func newAWSFactory() clientFactory { return awsFactory{} }

func (awsFactory) New(ctx context.Context) (secretClient, error) {
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

type secretUpdateState uint8

const (
	secretAvailable secretUpdateState = iota
	secretNotFound
	secretUnavailable
	secretInitializationFailed
	secretLookupFailed
)

type secretUpdate struct {
	state secretUpdateState
	value string
	err   error
}

type secretWatcher interface {
	Watch(context.Context, string, time.Duration) <-chan secretUpdate
}

type awsSecretWatcher struct {
	factory clientFactory
}

func newAWSSecretWatcher(factory clientFactory) secretWatcher {
	return &awsSecretWatcher{factory: factory}
}

func (w *awsSecretWatcher) Watch(ctx context.Context, name string, interval time.Duration) <-chan secretUpdate {
	updates := make(chan secretUpdate)
	go func() {
		defer close(updates)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var client secretClient
		var observation secretObservation
		refresh := func() bool {
			if client == nil {
				var err error
				client, err = w.factory.New(ctx)
				if err != nil {
					client = nil
					return publishSecretUpdate(ctx, updates, &observation, secretUpdate{state: secretInitializationFailed, err: err})
				}
			}

			result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name})
			return publishSecretUpdate(ctx, updates, &observation, newSecretUpdate(result, err))
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

func newSecretUpdate(result *secretsmanager.GetSecretValueOutput, err error) secretUpdate {
	if err != nil {
		var notFound *types.ResourceNotFoundException
		var invalidRequest *types.InvalidRequestException
		switch {
		case errors.As(err, &notFound):
			return secretUpdate{state: secretNotFound, err: err}
		case errors.As(err, &invalidRequest):
			return secretUpdate{state: secretUnavailable, err: err}
		default:
			return secretUpdate{state: secretLookupFailed, err: err}
		}
	}
	if result == nil || result.SecretString == nil {
		return secretUpdate{state: secretAvailable}
	}
	return secretUpdate{state: secretAvailable, value: *result.SecretString}
}

func publishSecretUpdate(ctx context.Context, updates chan<- secretUpdate, observation *secretObservation, update secretUpdate) bool {
	if !observation.changed(update) {
		return true
	}
	select {
	case updates <- update:
		return true
	case <-ctx.Done():
		return false
	}
}

type secretObservation struct {
	initialized bool
	state       secretUpdateState
	fingerprint [sha256.Size]byte
}

func (o *secretObservation) changed(update secretUpdate) bool {
	fingerprint := sha256.Sum256([]byte(update.value))
	if update.err != nil {
		fingerprint = sha256.Sum256([]byte(secretErrorIdentity(update.err)))
	}
	if o.initialized && o.state == update.state && o.fingerprint == fingerprint {
		return false
	}
	o.initialized = true
	o.state = update.state
	o.fingerprint = fingerprint
	return true
}

func secretErrorIdentity(err error) string {
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		return fmt.Sprintf("%T:%s:%s", apiError, apiError.ErrorCode(), apiError.ErrorMessage())
	}
	return fmt.Sprintf("%T:%s", err, err.Error())
}
