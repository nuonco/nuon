package waypoint

import (
	"github.com/powertoolsdev/go-generics"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
)

// NOTE: we can't use faker for this, because we need the map to have the proper keys, like the orgcontext package
// verifies. This should probably be exposed by the orgcontext package, instead (fwiw).
func getFakeOrgContext() *orgcontext.Context {
	obj := generics.GetFakeObj[*orgcontext.Context]()
	obj.Buckets = map[orgcontext.BucketType]orgcontext.Bucket{
		orgcontext.BucketTypeDeployments:   generics.GetFakeObj[orgcontext.Bucket](),
		orgcontext.BucketTypeInstallations: generics.GetFakeObj[orgcontext.Bucket](),
		orgcontext.BucketTypeOrgs:          generics.GetFakeObj[orgcontext.Bucket](),
	}
	return obj
}
