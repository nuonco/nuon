package backend

import "context"

const (
	DefaultConfigFileName string = "backend.json"
)

// Backend exposes a way to configure a backend using s3, or other providers
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=backend_mock.go -source=backend.go -package=backend
type Backend interface {
	Init(ctx context.Context) error
	ConfigFile(ctx context.Context) ([]byte, error)
}
