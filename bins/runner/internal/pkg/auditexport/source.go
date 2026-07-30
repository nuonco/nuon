package auditexport

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"
)

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
