package s3downloader

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// these are not table tests. There are only 2 cases and I
// cannot follow the code with another layer of anonymous types
// and complex code flow
func TestDeleteBlobs(t *testing.T) {
	ctx := context.Background()
	input := &s3.DeleteObjectsInput{}
	expectedReturn := &s3.DeleteObjectsOutput{}
	mockCtl := gomock.NewController(t)
	mockDeleter := NewMocks3Deleter(mockCtl)
	t.Run("success case", func(t *testing.T) {
		mockDeleter.EXPECT().DeleteObjects(ctx, input).Return(expectedReturn, nil)
		actualReturn, err := deleteBlobs(ctx, input, mockDeleter)
		assert.NoError(t, err)
		assert.Equal(t, expectedReturn, actualReturn)
	})

	t.Run("error case", func(t *testing.T) {
		fakeError := fmt.Errorf("unit test error")
		mockDeleter.EXPECT().DeleteObjects(ctx, input).Return(nil, fakeError)
		actualReturn, err := deleteBlobs(ctx, input, mockDeleter)
		assert.Error(t, err)
		assert.Nil(t, actualReturn)
		assert.ErrorContains(t, err, fakeError.Error())
	})
}
