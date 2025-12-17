package s3downloader

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nuonco/nuon/pkg/generics"
)

// S3Object represents an S3 object with its metadata
type S3Object struct {
	Key          string
	LastModified time.Time
	Size         int64
}

// ListPrefix assumes a role and returns a list of all the files in the s3 prefix
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=list_mock_test.go -source=list.go -package=s3downloader
func (s *s3Downloader) ListPrefix(ctx context.Context, key string) ([]string, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	return s.listPrefix(ctx, client, key)
}

func (s *s3Downloader) ListAll(ctx context.Context) ([]string, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	return s.listPrefix(ctx, client, "")
}

// ListPrefixWithMetadata returns a list of S3 objects with their metadata
func (s *s3Downloader) ListPrefixWithMetadata(ctx context.Context, key string) ([]S3Object, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	return s.listPrefixWithMetadata(ctx, client, key)
}

type s3Lister interface {
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

func (s *s3Downloader) listPrefix(ctx context.Context, client s3Lister, prefix string) ([]string, error) {
	req := &s3.ListObjectsV2Input{
		Bucket: generics.ToPtr(s.Bucket),
	}
	if prefix != "" {
		req.Prefix = generics.ToPtr(prefix)
	}

	keys := make([]string, 0)

	for {
		resp, err := client.ListObjectsV2(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("unable to list objects: %w", err)
		}

		for _, obj := range resp.Contents {
			keys = append(keys, *obj.Key)
		}

		if resp.NextContinuationToken == nil {
			break
		}
		req.ContinuationToken = resp.NextContinuationToken
	}

	return keys, nil
}

func (s *s3Downloader) listPrefixWithMetadata(ctx context.Context, client s3Lister, prefix string) ([]S3Object, error) {
	req := &s3.ListObjectsV2Input{
		Bucket: generics.ToPtr(s.Bucket),
	}
	if prefix != "" {
		req.Prefix = generics.ToPtr(prefix)
	}

	objects := make([]S3Object, 0)

	for {
		resp, err := client.ListObjectsV2(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("unable to list objects: %w", err)
		}

		for _, obj := range resp.Contents {
			s3Obj := S3Object{
				Key: *obj.Key,
			}
			if obj.Size != nil {
				s3Obj.Size = *obj.Size
			}
			if obj.LastModified != nil {
				s3Obj.LastModified = *obj.LastModified
			}
			objects = append(objects, s3Obj)
		}

		if resp.NextContinuationToken == nil {
			break
		}
		req.ContinuationToken = resp.NextContinuationToken
	}

	return objects, nil
}
