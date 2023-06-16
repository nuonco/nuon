package s3downloader

import (
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	bucketName := uuid.NewString()
	creds := generics.GetFakeObj[*credentials.Config]()

	tests := map[string]struct {
		optFns      func() []downloaderOption
		assertFn    func(*testing.T, *s3Downloader)
		errExpected error
	}{
		"happy path": {
			optFns: func() []downloaderOption {
				return []downloaderOption{
					WithCredentials(creds),
				}
			},
			assertFn: func(t *testing.T, d *s3Downloader) {
				assert.Equal(t, creds, d.Credentials)
				assert.Equal(t, bucketName, d.Bucket)
			},
		},
		"happy path creds": {
			optFns: func() []downloaderOption {
				return []downloaderOption{
					WithCredentials(creds),
				}
			},
			assertFn: func(t *testing.T, d *s3Downloader) {
				assert.Equal(t, creds, d.Credentials)
				assert.Equal(t, bucketName, d.Bucket)
			},
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
