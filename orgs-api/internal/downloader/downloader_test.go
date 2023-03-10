package downloader

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	bucketName := uuid.NewString()

	tests := map[string]struct {
		optFns      func() []downloaderOption
		assertFn    func(*testing.T, *s3Downloader)
		errExpected error
	}{
		"happy path": {
			optFns: func() []downloaderOption {
				return []downloaderOption{
					WithAssumeRoleARN("test-role-arn"),
					WithAssumeRoleSessionName("test-role-session-name"),
				}
			},
			assertFn: func(t *testing.T, d *s3Downloader) {
				assert.Equal(t, "test-role-arn", d.AssumeRoleARN)
				assert.Equal(t, "test-role-session-name", d.AssumeRoleSessionName)
				assert.Equal(t, bucketName, d.Bucket)
			},
		},
		"missing assume role arn": {
			optFns: func() []downloaderOption {
				return []downloaderOption{
					WithAssumeRoleSessionName("test-role-session-name"),
				}
			},
			errExpected: fmt.Errorf("s3Downloader.AssumeRoleARN"),
		},
		"missing session name": {
			optFns: func() []downloaderOption {
				return []downloaderOption{
					WithAssumeRoleARN("test-role-arn"),
				}
			},
			errExpected: fmt.Errorf("s3Downloader.AssumeRoleSessionName"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			opts := test.optFns()
			srv, err := New(bucketName, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, srv)
		})
	}
}
