package s3downloader

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/powertoolsdev/go-generics"
)

// GetBlob assumes a role and returns the actual blob from s3
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=get_blob_mock_test.go -source=get_blob.go -package=downloader
func (s *s3Downloader) GetBlob(ctx context.Context, key string) ([]byte, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	downloader := manager.NewDownloader(client)
	return s.getBlob(ctx, downloader, key)
}

type s3BlobGetter interface {
	Download(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*manager.Downloader)) (int64, error)
}

func (s *s3Downloader) getBlob(ctx context.Context, client s3BlobGetter, key string) ([]byte, error) {
	buf := aws.NewWriteAtBuffer([]byte{})
	_, err := client.Download(ctx, buf, &s3.GetObjectInput{
		Bucket: generics.ToPtr(s.Bucket),
		Key:    generics.ToPtr(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to download bytes: key=%s: %w", key, err)
	}

	return buf.Bytes(), err
}
