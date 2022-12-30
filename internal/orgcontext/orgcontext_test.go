package orgcontext

import (
	"github.com/powertoolsdev/go-generics"
)

func getFakeContext() *Context {
	obj := generics.GetFakeObj[*Context]()
	obj.Buckets = map[BucketType]Bucket{
		BucketTypeDeployments:   generics.GetFakeObj[Bucket](),
		BucketTypeInstallations: generics.GetFakeObj[Bucket](),
		BucketTypeOrgs:          generics.GetFakeObj[Bucket](),
	}
	return obj
}
