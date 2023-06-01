package callbacks

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()

	bucketKeySettings := generics.GetFakeObj[BucketKeySettings]()
	creds := generics.GetFakeObj[*credentials.Config]()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []s3CallbackOption
		assertFn    func(*testing.T, *s3Callback)
	}{
		"happy path": {
			optsFn: func() []s3CallbackOption {
				return []s3CallbackOption{
					WithCredentials(creds),
					WithBucketKeySettings(bucketKeySettings),
				}
			},
			assertFn: func(t *testing.T, e *s3Callback) {
				assert.Equal(t, bucketKeySettings.Bucket, e.Bucket)
				assert.Equal(t, bucketKeySettings.BucketPrefix, e.BucketPrefix)
				assert.Equal(t, bucketKeySettings.Filename, e.Filename)
				assert.Equal(t, creds, e.Credentials)
			},
		},
		"invalid settings": {
			optsFn: func() []s3CallbackOption {
				return []s3CallbackOption{
					WithCredentials(creds),
					WithBucketKeySettings(BucketKeySettings{}),
				}
			},
			errExpected: fmt.Errorf("unable to validate bucket"),
		},
		"invalid creds": {
			optsFn: func() []s3CallbackOption {
				return []s3CallbackOption{
					WithBucketKeySettings(bucketKeySettings),
					WithCredentials(&credentials.Config{}),
				}
			},
			errExpected: fmt.Errorf("unable to validate cred"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			e, err := newS3Callback(v, test.optsFn()...)
			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, e)
		})
	}
}
