package orgcontext

import "context"

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=orgcontext_mocks.go -source=orgcontext.go -package=orgcontext

// This entire api is designed to work in a single tenant fashion, where every request is scoped to a single org. The
// org context is designed so we can easily inject all of the information needed for an org into the context and use it
// throughout the request lifecycle.

// Provider exposes an interface for setting a context which can be used by this package
type Provider interface {
	SetContext(context.Context, string) (context.Context, error)
}

type provider = Provider

// Bucket contains the information needed to access a specific bucket
type Bucket struct {
	Name          string `json:"name" validate:"required"`
	Prefix        string `json:"prefix" validate:"required"`
	AssumeRoleARN string `json:"assume_role_arn" validate:"required"`
}

// WaypointServer contains all of the information needed to access the waypoint server
type WaypointServer struct {
	Address         string `json:"address" validate:"required"`
	SecretNamespace string `json:"secret_namespace" validate:"required"`
	SecretName      string `json:"secret_name" validate:"required"`

	// TODO(jm): eventually update this to use kube.ClusterInfo and the orgs cluster
}

// Context is injected into each "request"
type Context struct {
	OrgsBucket          Bucket `json:"orgs_bucket" validate:"required"`
	InstallationsBucket Bucket `json:"installations_bucket" validate:"required"`
	DeploymentsBucket   Bucket `json:"deployments_bucket" validate:"required"`

	WaypointServer WaypointServer `json:"waypoint_server" validate:"required"`
}

type orgContextKey struct{}
