package s3

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()

	iam := generics.GetFakeObj[*IAMConfig]()
	creds := generics.GetFakeObj[*Credentials]()
	bucket := generics.GetFakeObj[*BucketConfig]()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []s3Option
		assertFn    func(*testing.T, *s3)
	}{
		"happy path - with creds": {
			optsFn: func() []s3Option {
				return []s3Option{
					WithBucketConfig(bucket),
					WithCredentials(creds),
				}
			},
			assertFn: func(t *testing.T, s *s3) {
				assert.Equal(t, s.Bucket, bucket)
				assert.Equal(t, s.Credentials, creds)
			},
		},
		"happy path - iam": {
			optsFn: func() []s3Option {
				return []s3Option{
					WithBucketConfig(bucket),
					WithIAMConfig(iam),
				}
			},
			assertFn: func(t *testing.T, s *s3) {
				assert.Equal(t, iam, s.IAM)
				assert.Equal(t, bucket, s.Bucket)
			},
		},
		"missing bucket": {
			optsFn: func() []s3Option {
				return []s3Option{
					WithIAMConfig(iam),
				}
			},
			errExpected: fmt.Errorf("Bucket"),
		},
		"missing creds/iam": {
			optsFn: func() []s3Option {
				return []s3Option{
					WithBucketConfig(bucket),
				}
			},
			errExpected: fmt.Errorf("invalid options, one of IAM or Credentials"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			e, err := New(v, test.optsFn()...)
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
