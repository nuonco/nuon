package github

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
)

// k8s secrets are actually a mapping, and this is the default key we use to store the actual value
//
//nolint:gosec
const appKeySecretKeyKey string = "github_app_key"

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=github_mock.go -source=github.go -package=github
type CloneTokenGetter interface {
	InstallationToken(context.Context) (string, error)
	ClonePath(context.Context) (string, error)
}

// gh is a type that lets you get a clone token for a repo
type gh struct {
	v *validator.Validate `validate:"required"`

	RepoName  string `validate:"required"`
	RepoOwner string `validate:"required"`
	AppKeyID  string `validate:"required"`
	InstallID int64  `validate:"required"`

	AppKeySecretName      string `validate:"required"`
	AppKeySecretNamespace string `validate:"required"`
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

// WithInstallID is used to set the registry id
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

// WithAppKeyID is used to set the iam role arn to auth with
func WithAppKeyID(appKeyID string) Option {
	return func(g *gh) error {
		g.AppKeyID = appKeyID
		return nil
	}
}

// WithAppKeySecretName is used to set the secret name
func WithAppKeySecretName(secretName string) Option {
	return func(g *gh) error {
		g.AppKeySecretName = secretName
		return nil
	}
}

// WithAppKeySecretNamesapce is used to set the k8s secret namespace
func WithAppKeySecretNamespace(ns string) Option {
	return func(g *gh) error {
		g.AppKeySecretNamespace = ns
		return nil
	}
}

// WithRepo is used to set the repo with the org/repo format
func WithRepo(repo string) Option {
	return func(g *gh) error {
		owner, name, err := parseRepo(repo)
		if err != nil {
			return err
		}

		g.RepoName = name
		g.RepoOwner = owner
		return nil
	}
}
