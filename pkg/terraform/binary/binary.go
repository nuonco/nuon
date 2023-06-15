package binary

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

// Binary exposes a way to initialize binary, for use by a workspace
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=binary_mock.go -source=binary.go -package=binary
type Binary interface {
	// Install should install the appropriate binary into a path that is within the provided dir, and return the
	// exec path to be used
	Install(context.Context, hclog.Logger, string) (string, error)

	Init(context.Context) error
}
