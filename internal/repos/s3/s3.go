package s3

import (
	"context"
	"time"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=s3_mock.go -source=s3.go -package=s3
type Repo interface {
	GetKey(context.Context, string, string, RoleConfig) ([]byte, error)
}

// RoleConfig is used to pass in information about the role and how it should be used
type RoleConfig struct {
	RoleARN     string
	SessionName string
	MaxDuration time.Duration
}

type repo struct{}

var _ Repo = (*repo)(nil)
