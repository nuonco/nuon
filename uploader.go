package uploader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// TODO(jdt): add interfaces and test

// uploader is the interface for uploading data into output runs directory
type Uploader interface {
	SetUploadPrefix(string)

	// uploadFile writes the data in the file into the output s3 blob
	UploadFile(context.Context, string, string, string) error

	// uploadBlob writes the data in the byte slice into the output s3 blob
	UploadBlob(context.Context, []byte, string) error
}

func NewS3Uploader(bucket, prefix string) *s3Uploader {
	return &s3Uploader{
		installBucket: bucket,
		prefix:        prefix,
	}
}

type s3Uploader struct {
	prefix        string
	installBucket string
}

func (s *s3Uploader) SetUploadPrefix(prefix string) {
	ts := time.Now()
	s.prefix = filepath.Join(prefix, fmt.Sprintf("runs/ts=%d", ts.Unix()))
}

func (s *s3Uploader) UploadFile(ctx context.Context, tmpDir, inputName, outputName string) error {
	if s.prefix == "" {
		return fmt.Errorf("unable to upload, missing prefix")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	inputFp := filepath.Join(tmpDir, inputName)
	f, err := os.Open(inputFp)
	if err != nil {
		return err
	}
	defer f.Close()

	return s.upload(ctx, uploader, f, outputName)
}

func (s *s3Uploader) UploadBlob(ctx context.Context, byts []byte, outputName string) error {
	if s.prefix == "" {
		return fmt.Errorf("unable to upload, missing prefix")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)
	f := bytes.NewReader(byts)

	return s.upload(ctx, uploader, f, outputName)
}

type s3UploaderClient interface {
	Upload(context.Context, *s3.PutObjectInput, ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

func (s *s3Uploader) upload(ctx context.Context, client s3UploaderClient, f io.Reader, name string) error {
	key := filepath.Join(s.prefix, name)
	bucket := s.installBucket
	_, err := client.Upload(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	})
	return err
}
