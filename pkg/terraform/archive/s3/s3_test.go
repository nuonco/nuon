package s3

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	expected := generics.GetFakeObj[s3]()
	v := validator.New()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []s3Option
		assertFn    func(*testing.T, *s3)
	}{
		"happy path": {
			optsFn: func() []s3Option {
				return []s3Option{
					WithRoleSessionName(expected.RoleSessionName),
					WithRoleARN(expected.RoleARN),
					WithBucketName(expected.BucketName),
					WithBucketKey(expected.Key),
				}
			},
			assertFn: func(t *testing.T, s *s3) {
				assert.Equal(t, expected.BucketName, s.BucketName)
				assert.Equal(t, expected.Key, s.Key)
				assert.Equal(t, expected.RoleARN, s.RoleARN)
				assert.Equal(t, expected.RoleSessionName, s.RoleSessionName)
			},
		},
		"missing role arn": {
			optsFn: func() []s3Option {
				return []s3Option{
					WithRoleSessionName(expected.RoleSessionName),
					WithBucketName(expected.BucketName),
					WithBucketKey(expected.Key),
				}
			},
			errExpected: fmt.Errorf("RoleARN"),
		},
		"missing role session name": {
			optsFn: func() []s3Option {
				return []s3Option{
					WithRoleARN(expected.RoleARN),
					WithBucketName(expected.BucketName),
					WithBucketKey(expected.Key),
				}
			},
			errExpected: fmt.Errorf("RoleSessionName"),
		},
		"missing bucket name": {
			optsFn: func() []s3Option {
				return []s3Option{
					WithRoleSessionName(expected.RoleSessionName),
					WithRoleARN(expected.RoleARN),
					WithBucketKey(expected.Key),
				}
			},
			errExpected: fmt.Errorf("BucketName"),
		},
		"missing bucket key": {
			optsFn: func() []s3Option {
				return []s3Option{
					WithRoleSessionName(expected.RoleSessionName),
					WithRoleARN(expected.RoleARN),
					WithBucketName(expected.BucketName),
				}
			},
			errExpected: fmt.Errorf("Key"),
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
