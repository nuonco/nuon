package s3downloader

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=deleter_mock.go -source=deleter.go -package=s3downloader
type Deleter interface {
	DeleteBlobs(context.Context, s3.DeleteObjectsInput) (s3.DeleteObjectOutput, error)
}

// This interface exists to allow us to mock s3 calls in test code
// It exactly matches the function signature for S3 client.DeleteObjects
type s3Deleter interface {
	DeleteObjects(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

func (s *s3Downloader) DeleteBlobs(ctx context.Context, input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	return deleteBlobs(ctx, input, client)
}

func deleteBlobs(ctx context.Context, input *s3.DeleteObjectsInput, deleter s3Deleter) (*s3.DeleteObjectsOutput, error) {
	return deleter.DeleteObjects(ctx, input)
}
