package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/mholt/archiver/v4"
	"github.com/powertoolsdev/mono/pkg/aws/s3downloader"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
)

func (s *s3) Unpack(ctx context.Context, cb archive.Callback) error {
	downloader, err := s3downloader.New(s.BucketName,
		s3downloader.WithCredentials(s.Credentials),
	)
	if err != nil {
		return fmt.Errorf("unable to get s3downloader: %w", err)
	}

	byts, err := downloader.GetBlob(ctx, s.Key)
	if err != nil {
		return fmt.Errorf("unable to get bucket key: %w", err)
	}

	reader := bytes.NewReader(byts)
	if err := s.unpack(ctx, reader, cb); err != nil {
		return fmt.Errorf("unable to get bucket: %w", err)
	}

	return nil
}

// unpack: accepts a gzipped, tarballed file and calls the callback for each file found
func (s *s3) unpack(ctx context.Context, r io.Reader, fn archive.Callback) error {
	gz := archiver.Gz{}
	reader, err := gz.OpenReader(r)
	if err != nil {
		return fmt.Errorf("unable to decompress gz: %w", err)
	}
	defer reader.Close()

	tar := archiver.Tar{}
	if err := tar.Extract(ctx, reader, nil, func(ctx context.Context, f archiver.File) error {
		if f.IsDir() {
			return nil
		}

		inputFile, err := f.Open()
		if err != nil {
			return fmt.Errorf("unable to open file from tar: %w", err)
		}
		defer inputFile.Close()

		if err := fn(ctx, f.NameInArchive, inputFile); err != nil {
			return fmt.Errorf("unable to write %s to callback: %w", f.NameInArchive, err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
