package s3

import (
	"context"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=s3_mock.go -source=s3.go -package=s3
type Repo interface {
	GetOrgsKey(context.Context, string) ([]byte, error)
	GetInstallationsKey(context.Context, string) ([]byte, error)
	GetDeploymentsKey(context.Context, string) ([]byte, error)
}

type repo struct{}

var _ Repo = (*repo)(nil)
