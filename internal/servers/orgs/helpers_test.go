package orgs

import (
	"github.com/powertoolsdev/go-generics"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
)

func getFakeOrgContext() *orgcontext.Context {
	obj := generics.GetFakeObj[*orgcontext.Context]()
	obj.Buckets = map[orgcontext.BucketType]orgcontext.Bucket{
		orgcontext.BucketTypeDeployments:   generics.GetFakeObj[orgcontext.Bucket](),
		orgcontext.BucketTypeInstallations: generics.GetFakeObj[orgcontext.Bucket](),
		orgcontext.BucketTypeOrgs:          generics.GetFakeObj[orgcontext.Bucket](),
	}
	return obj
}
