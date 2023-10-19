package hooks

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

func ValidHooks() []string {
	return []string{
		"pre-apply.sh",
		"post-apply.sh",
		"error-apply.sh",

		"pre-destroy.sh",
		"post-destroy.sh",
		"error-destroy.sh",
	}
}

// Hooks expose an interface that enables post/pre commands to be executed in a terraform workspace
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=hooks_mock.go -source=hooks.go -package=hooks
type Hooks interface {
	Init(context.Context, string) error

	PreApply(context.Context, hclog.Logger) error
	PostApply(context.Context, hclog.Logger) error
	ErrorApply(context.Context, hclog.Logger) error

	PreDestroy(context.Context, hclog.Logger) error
	PostDestroy(context.Context, hclog.Logger) error
	ErrorDestroy(context.Context, hclog.Logger) error
}
