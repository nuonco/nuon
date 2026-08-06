package transport

import (
	"context"
	"errors"
	"io"
	"time"

	"go.uber.org/fx"
)

var ErrNotConfigured = errors.New("air-gap bundle storage is not configured")

type PublishRequest struct {
	Body   io.ReadSeeker
	Size   int64
	SHA256 string
}

type Replica struct {
	Provider          string
	Region            string
	StorageRef        string
	StorageVersion    string
	TransportChecksum string
	Size              int64
	VerifiedAt        time.Time
}

type DownloadGrant struct {
	URL           string
	ExpiresAt     time.Time
	SupportsRange bool
}

type Store interface {
	Configured() bool
	Publish(context.Context, PublishRequest) (Replica, error)
	Grant(context.Context, Replica, string, time.Time) (DownloadGrant, error)
}

func AsStore(constructor any) any {
	return fx.Annotate(constructor, fx.As(new(Store)))
}

type disabledStore struct{}

func NewDisabled() Store { return disabledStore{} }

func (disabledStore) Configured() bool { return false }

func (disabledStore) Publish(context.Context, PublishRequest) (Replica, error) {
	return Replica{}, ErrNotConfigured
}

func (disabledStore) Grant(context.Context, Replica, string, time.Time) (DownloadGrant, error) {
	return DownloadGrant{}, ErrNotConfigured
}
