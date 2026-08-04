package telemetryexport

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

const telemetryExportConfigSecretName = "telemetry-export-config"

type configUpdateState uint8

const (
	configAvailable configUpdateState = iota
	configNotFound
	configUnavailable
	configSourceInitializationFailed
	configLookupFailed
)

type configUpdate struct {
	state         configUpdateState
	value         string
	err           error
	errorIdentity string
}

type configSource interface {
	Watch(context.Context, time.Duration) <-chan configUpdate
}

type configSourceResolver interface {
	Resolve(string, string) configSource
}

type sourceResolver struct {
	awsFactory   awsClientFactory
	azureFactory azureClientFactory
	gcpFactory   gcpClientFactory
}

func newConfigSourceResolver(awsFactory awsClientFactory, azureFactory azureClientFactory, gcpFactory gcpClientFactory) configSourceResolver {
	return &sourceResolver{awsFactory: awsFactory, azureFactory: azureFactory, gcpFactory: gcpFactory}
}

func (r *sourceResolver) Resolve(platform, installID string) configSource {
	if installID == "" {
		return nil
	}
	platform = strings.ToLower(platform)
	switch {
	case strings.HasPrefix(platform, "aws"):
		return newAWSConfigSource(r.awsFactory, installID)
	case strings.HasPrefix(platform, "azure"):
		return newAzureConfigSource(r.azureFactory, installID)
	case platform == "gcp":
		return newGCPConfigSource(r.gcpFactory, installID)
	default:
		return nil
	}
}

func watchConfig(ctx context.Context, interval time.Duration, fetch func() configUpdate, cleanup func()) <-chan configUpdate {
	updates := make(chan configUpdate)
	go func() {
		defer close(updates)
		if cleanup != nil {
			defer cleanup()
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var observation configObservation
		refresh := func() bool {
			return publishConfigUpdate(ctx, updates, &observation, fetch())
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

func publishConfigUpdate(ctx context.Context, updates chan<- configUpdate, observation *configObservation, update configUpdate) bool {
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

type configObservation struct {
	initialized bool
	state       configUpdateState
	fingerprint [sha256.Size]byte
}

func (o *configObservation) changed(update configUpdate) bool {
	identity := update.value
	if update.err != nil {
		identity = update.errorIdentity
		if identity == "" {
			identity = fmt.Sprintf("%T:%s", update.err, update.err.Error())
		}
	}
	fingerprint := sha256.Sum256([]byte(identity))
	if o.initialized && o.state == update.state && o.fingerprint == fingerprint {
		return false
	}
	o.initialized = true
	o.state = update.state
	o.fingerprint = fingerprint
	return true
}
