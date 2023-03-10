package orgcontext

import (
	"context"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=orgcontext_mocks.go -source=orgcontext.go -package=orgcontext

// This entire api is designed to work in a single tenant fashion, where every request is scoped to a single org. The
// org context is designed so we can easily inject all of the information needed for an org into the context and use it
// throughout the request lifecycle.

// Provider exposes an interface for setting a context which can be used by this package
type Provider interface {
	Get(context.Context, string) (*Context, error)
}

type provider = Provider

type BucketType string

const (
	BucketTypeUnknown BucketType = ""

	BucketTypeOrgs        BucketType = "orgs"
	BucketTypeApps        BucketType = "apps"
	BucketTypeInstalls    BucketType = "installs"
	BucketTypeDeployments BucketType = "deployments"
	BucketTypeInstances   BucketType = "instances"
)

const (
	defaultAssumeRoleName string = "orgs-api"
)

type Bucket struct {
	Name               string `validate:"required"`
	IamRoleArn         string `validate:"required"`
	IamRoleSessionName string `validate:"required"`
}

// WaypointServer contains all of the information needed to access the waypoint server
type WaypointServer struct {
	Address         string `json:"address" validate:"required"`
	SecretNamespace string `json:"secret_namespace" validate:"required"`
	SecretName      string `json:"secret_name" validate:"required"`
}

// Context is injected into each "request"
type Context struct {
	OrgID string `json:"org_id" validate:"required"`

	Buckets        map[BucketType]Bucket `json:"bucket_access" validate:"required" faker:"-"`
	WaypointServer WaypointServer        `json:"waypoint_server" validate:"required"`
}
