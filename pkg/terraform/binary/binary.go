package binary

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=binary_mock.go -source=binary.go -package=binary
type Binary interface {
	Install(context.Context, hclog.Logger, string) (string, error)

	Uninstall(context.Context) error

	Init(context.Context) error
}
