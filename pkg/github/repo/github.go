package github

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/kube"
)

//nolint:gosec
const appKeySecretKeyKey string = "github_app_key"

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=github_mock.go -source=github.go -package=github
type CloneTokenGetter interface {
	InstallationToken(context.Context) (string, error)
	ClonePath(context.Context) (string, error)
}

type gh struct {
	v *validator.Validate `validate:"required"`

	RepoName  string `validate:"required"`
	RepoOwner string `validate:"required"`
	AppKeyID  string `validate:"required"`
	InstallID int64  `validate:"required"`

	AppKeySecretName      string            `validate:"required"`
	AppKeySecretNamespace string            `validate:"required"`
	AppKeyClusterInfo     *kube.ClusterInfo `validate:"required"`
}

type Option func(*gh) error

func New(v *validator.Validate, opts ...Option) (*gh, error) {
	g := &gh{
		v: v,
	}

	for _, opt := range opts {
		if err := opt(g); err != nil {
			return nil, err
		}
	}

	if err := g.v.Struct(g); err != nil {
		return nil, err
	}

	return g, nil
}

func WithInstallID(installID string) Option {
	return func(g *gh) error {
		ghInstallID, err := strconv.ParseInt(installID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid github install id: %w", err)
		}
		g.InstallID = ghInstallID
		return nil
	}
}

func WithAppKeyID(appKeyID string) Option {
	return func(g *gh) error {
		g.AppKeyID = appKeyID
		return nil
	}
}

func WithAppKeySecretName(secretName string) Option {
	return func(g *gh) error {
		g.AppKeySecretName = secretName
		return nil
	}
}

func WithAppKeySecretNamespace(ns string) Option {
	return func(g *gh) error {
		g.AppKeySecretNamespace = ns
		return nil
	}
}

func WithAppKeyClusterInfo(info *kube.ClusterInfo) Option {
	return func(g *gh) error {
		g.AppKeyClusterInfo = info
		return nil
	}
}

func WithRepo(repo string) Option {
	return func(g *gh) error {
		owner, name, err := ParseRepo(repo)
		if err != nil {
			return err
		}

		g.RepoName = name
		g.RepoOwner = owner
		return nil
	}
}
