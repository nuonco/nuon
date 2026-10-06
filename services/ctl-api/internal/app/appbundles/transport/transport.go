package transport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"go.uber.org/fx"

	ctlconfig "github.com/nuonco/nuon/services/ctl-api/internal"
)

var ErrNotConfigured = errors.New("portable bundle storage is not configured")

type StoreParams struct {
	fx.In
	Config    *ctlconfig.Config
	Lifecycle fx.Lifecycle
}

func NewStore(params StoreParams) (Store, error) {
	if params.Config == nil {
		return nil, errors.New("app bundle storage config is required")
	}
	cfg := *params.Config
	provider := cfg.AppBundleStorageProvider
	if provider == "" {
		provider = cfg.BlobStorageProvider
	}
	if strings.TrimSpace(cfg.AppBundleStorageBucket) == "" {
		if cfg.AppBundleStorageProvider != "" && cfg.AppBundleStorageProvider != cfg.BlobStorageProvider {
			return nil, errors.New("a dedicated app bundle bucket is required when overriding the blob storage provider")
		}
		cfg.AppBundleStorageBucket = cfg.BlobStorageBucket
		cfg.AppBundleStorageRegion = cfg.BlobStorageRegion
	}
	if strings.TrimSpace(cfg.AppBundleStorageBucket) == "" {
		return NewDisabled(), nil
	}
	switch provider {
	case "s3":
		return NewS3(S3Params{Config: &cfg})
	case "gcs":
		return NewGCS(StoreParams{Config: &cfg, Lifecycle: params.Lifecycle})
	default:
		return nil, fmt.Errorf("unsupported app bundle storage provider %q", provider)
	}
}

type PublishRequest struct {
	Body   io.ReadSeeker
	Size   int64
	SHA256 string
}

type Replica struct {
	Provider          string
	Bucket            string
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
	Delete(context.Context, Replica) error
	Grant(context.Context, Replica, string, time.Time) (DownloadGrant, error)
}

func AsStore(constructor any) any {
	return fx.Annotate(constructor, fx.As(new(Store)))
}

func canonicalSHA256(value string) (string, []byte, error) {
	value = strings.TrimPrefix(strings.ToLower(value), "sha256:")
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size {
		return "", nil, errors.New("SHA-256 must be 64 hexadecimal characters")
	}
	return value, decoded, nil
}

func safeFilename(value string) string {
	value = path.Base(strings.ReplaceAll(value, "\\", "/"))
	value = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || r == '"' {
			return '_'
		}
		return r
	}, value)
	if value == "" || value == "." {
		return "portable-bundle.tar.zst"
	}
	return value
}

type disabledStore struct{}

func NewDisabled() Store { return disabledStore{} }

func (disabledStore) Configured() bool { return false }

func (disabledStore) Publish(context.Context, PublishRequest) (Replica, error) {
	return Replica{}, ErrNotConfigured
}

func (disabledStore) Delete(context.Context, Replica) error {
	return ErrNotConfigured
}

func (disabledStore) Grant(context.Context, Replica, string, time.Time) (DownloadGrant, error) {
	return DownloadGrant{}, ErrNotConfigured
}
