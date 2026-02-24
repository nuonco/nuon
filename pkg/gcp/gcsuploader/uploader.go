package gcsuploader

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type Uploader struct {
	bucketName string
}

func New(bucketName string) *Uploader {
	return &Uploader{bucketName: bucketName}
}

func (u *Uploader) Upload(ctx context.Context, data []byte, objectName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("unable to create GCS client: %w", err)
	}
	defer client.Close()

	wc := client.Bucket(u.bucketName).Object(objectName).NewWriter(ctx)
	wc.ContentType = "application/json"

	if _, err := io.Writer(wc).Write(data); err != nil {
		wc.Close()
		return fmt.Errorf("unable to write to GCS: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("unable to close GCS writer: %w", err)
	}

	return nil
}
