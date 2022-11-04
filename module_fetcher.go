package terraform

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mholt/archiver/v4"
)

const (
	baseTmpDir   string = "/tmp"
	tmpDirPrefix string = "nuon-module-"
)

type Module struct {
	BucketName       string `validate:"required"`
	BucketKey        string `validate:"required"`
	TerraformVersion string `validate:"required"`
}

// moduleFetcher is an api for grabbing a module into a local directory to terraform against
type moduleFetcher interface {
	createTmpDir(string) (string, error)
	cleanupTmpDir(string) error
	fetchModule(context.Context, Module, string) error
}

type s3ModuleFetcher struct{}

var _ moduleFetcher = (*s3ModuleFetcher)(nil)

// createTmpDir: create a temporary directory
func (s *s3ModuleFetcher) createTmpDir(installID string) (string, error) {
	dir, err := os.MkdirTemp(baseTmpDir, tmpDirPrefix+installID)
	if err != nil {
		return "", nil
	}

	return dir, nil
}

func (s s3ModuleFetcher) getS3Key(name, version string) string {
	return fmt.Sprintf("modulees/%s_%s.tar.gz", name, version)
}

// cleanupTmpDir: clean up the provided tmp directory
func (s *s3ModuleFetcher) cleanupTmpDir(path string) error {
	return os.RemoveAll(path)
}

// fetchModule: copy source files from s3 into the temporary directory
func (s *s3ModuleFetcher) fetchModule(ctx context.Context, module Module, tmpDir string) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg)
	downloader := manager.NewDownloader(client)
	byts, err := s.downloadModule(ctx, module, downloader)
	if err != nil {
		return err
	}

	if err := s.extractModule(ctx, tmpDir, byts); err != nil {
		return err
	}

	return nil
}

type s3Downloader interface {
	Download(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*manager.Downloader)) (int64, error)
}

func (s *s3ModuleFetcher) downloadModule(ctx context.Context, module Module, client s3Downloader) ([]byte, error) {
	writer := manager.NewWriteAtBuffer(nil)
	_, err := client.Download(ctx, writer, &s3.GetObjectInput{
		Bucket: &module.BucketName,
		Key:    &module.BucketKey,
	})
	if err != nil {
		return nil, err
	}

	return writer.Bytes(), nil
}

// extractModule: accepts a module and extracts it into the tmpdir
func (s *s3ModuleFetcher) extractModule(ctx context.Context, tmpDir string, byts []byte) error {
	bytReader := bytes.NewReader(byts)

	decom := archiver.Gz{}
	reader, err := decom.OpenReader(bytReader)
	if err != nil {
		return err
	}
	defer reader.Close()

	fmt := archiver.Tar{}
	if err := fmt.Extract(ctx, reader, nil, func(ctx context.Context, f archiver.File) error {
		outputFp := filepath.Join(tmpDir, f.NameInArchive)
		if outputFp == tmpDir {
			return nil
		}

		inputFile, err := f.Open()
		if err != nil {
			return err
		}
		defer inputFile.Close()

		outputFile, err := os.Create(outputFp)
		if err != nil {
			return err
		}
		defer outputFile.Close()

		_, err = io.Copy(outputFile, inputFile)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
