package provision

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getFakeUploadMetadataRequest() UploadMetadataRequest {
	return UploadMetadataRequest{
		BucketName:   "bucket-name",
		BucketPrefix: "bucket-prefix",
	}
}

func TestUploadMetadataRequest_validation(t *testing.T) {
	tests := map[string]struct {
		reqFn       func() UploadMetadataRequest
		errExpected error
	}{
		"happy path": {
			reqFn: getFakeUploadMetadataRequest,
		},
		"missing-org-id": {
			reqFn: func() UploadMetadataRequest {
				req := getFakeUploadMetadataRequest()
				req.BucketName = ""
				return req
			},
			errExpected: fmt.Errorf("UploadMetadataRequest.BucketName"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := test.reqFn()
			err := req.validate()
			if test.errExpected == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorContains(t, err, test.errExpected.Error())
			}
		})
	}
}
