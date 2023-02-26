package waypoint

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	waypointclient "github.com/powertoolsdev/go-waypoint/v2/pkg/client"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=waypoint_mocks.go -source=waypoint.go -package=waypoint
type Repo interface {
	GetVersionInfo(context.Context) (*waypointv1.GetVersionInfoResponse, error)
	GetRunner(context.Context, string) (*waypointv1.Runner, error)
	ListRunners(context.Context) (*waypointv1.ListRunnersResponse, error)
}

type repo struct {
	v *validator.Validate `validate:"required"`

	Address         string `validate:"required"`
	SecretNamespace string `validate:"required"`
	SecretName      string `validate:"required"`

	// set internally, or overridden for testing
	WaypointClientProvider waypointClientProvider `validate:"required"`
}

var _ Repo = (*repo)(nil)

func New(v *validator.Validate, opts ...repoOption) (*repo, error) {
	r := &repo{
		v: v,
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	// build a waypoint client provider
	wpClientProvider, err := waypointclient.NewOrgProvider(v, waypointclient.WithOrgConfig(waypointclient.Config{
		Address: r.Address,
		Token: waypointclient.Token{
			Namespace: r.SecretNamespace,
			Name:      r.SecretName,
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to fetch waypoint client provider: %w", err)
	}
	r.WaypointClientProvider = wpClientProvider

	if err := v.Struct(r); err != nil {
		return nil, err
	}

	return r, nil
}

type repoOption func(*repo) error

func WithAddress(addr string) repoOption {
	return func(r *repo) error {
		r.Address = addr
		return nil
	}
}

func WithSecretNamespace(ns string) repoOption {
	return func(r *repo) error {
		r.SecretNamespace = ns
		return nil
	}
}

func WithSecretName(name string) repoOption {
	return func(r *repo) error {
		r.SecretName = name
		return nil
	}
}
