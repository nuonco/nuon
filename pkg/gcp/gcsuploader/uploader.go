package gcsuploader

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
)

type Uploader struct {
	bucketName string
}

func New(bucketName string) *Uploader {
	return &Uploader{bucketName: bucketName}
}

// UploadAsZip packages the data as a zip file containing main.tf.json and uploads to GCS.
// Infrastructure Manager requires a zip or directory, not a single file.
func (u *Uploader) Upload(ctx context.Context, data []byte, objectName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("unable to create GCS client: %w", err)
	}
	defer client.Close()

	// create zip containing main.tf.json
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create("main.tf.json")
	if err != nil {
		return fmt.Errorf("unable to create zip entry: %w", err)
	}
	if _, err := fw.Write(data); err != nil {
		return fmt.Errorf("unable to write zip entry: %w", err)
	}
	if err := zw.Close(); err != nil {
		return fmt.Errorf("unable to close zip: %w", err)
	}

	// upload as .zip
	zipObjectName := strings.TrimSuffix(objectName, ".json") + ".zip"
	wc := client.Bucket(u.bucketName).Object(zipObjectName).NewWriter(ctx)
	wc.ContentType = "application/zip"

	if _, err := io.Writer(wc).Write(buf.Bytes()); err != nil {
		wc.Close()
		return fmt.Errorf("unable to write to GCS: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("unable to close GCS writer: %w", err)
	}

	return nil
}
