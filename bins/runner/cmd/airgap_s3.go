package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type airgapS3Client interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type airgapS3Sync struct {
	client airgapS3Client
	bucket string
	prefix string
}

func parseS3URI(uri string) (string, string, error) {
	if !strings.HasPrefix(uri, "s3://") {
		return "", "", fmt.Errorf("S3 URI must start with s3://")
	}
	value := strings.TrimPrefix(uri, "s3://")
	bucket, key, found := strings.Cut(value, "/")
	if bucket == "" {
		return "", "", fmt.Errorf("S3 URI bucket is required")
	}
	if !found {
		key = ""
	}
	return bucket, strings.Trim(key, "/"), nil
}

func newAirgapS3Sync(ctx context.Context, uri string) (*airgapS3Sync, error) {
	bucket, prefix, err := parseS3URI(uri)
	if err != nil {
		return nil, err
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}
	return &airgapS3Sync{client: s3.NewFromConfig(cfg), bucket: bucket, prefix: prefix}, nil
}

func downloadAirgapBundle(ctx context.Context, uri, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return fmt.Errorf("create bundle download directory: %w", err)
	}
	if !strings.HasPrefix(uri, "s3://") {
		src, err := os.Open(uri)
		if err != nil {
			return fmt.Errorf("open local bundle: %w", err)
		}
		defer src.Close()
		dstFile, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("create bundle archive: %w", err)
		}
		_, copyErr := io.Copy(dstFile, src)
		closeErr := dstFile.Close()
		if copyErr != nil {
			return fmt.Errorf("copy local bundle: %w", copyErr)
		}
		return closeErr
	}
	bucket, key, err := parseS3URI(uri)
	if err != nil {
		return err
	}
	if key == "" {
		return fmt.Errorf("bundle S3 URI key is required")
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("load AWS configuration: %w", err)
	}
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create bundle archive: %w", err)
	}
	_, downloadErr := manager.NewDownloader(s3.NewFromConfig(cfg)).Download(ctx, f, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
	closeErr := f.Close()
	if downloadErr != nil {
		return fmt.Errorf("download bundle: %w", downloadErr)
	}
	return closeErr
}

const (
	stackOutputsPollInterval = 15 * time.Second
	stackOutputsPollTimeout  = 30 * time.Minute
)

// fetchInstallStackOutputs retries all fetch errors because ASG startup races both object creation and IAM propagation.
func fetchInstallStackOutputs(ctx context.Context, uri string, logger *zap.Logger) ([]byte, error) {
	bucket, key, err := parseS3URI(uri)
	if err != nil {
		return nil, err
	}
	if key == "" {
		return nil, fmt.Errorf("install stack outputs S3 URI key is required")
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}
	client := s3.NewFromConfig(cfg)

	deadline := time.Now().Add(stackOutputsPollTimeout)
	var lastErr error
	for {
		out, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
		if err == nil {
			data, readErr := io.ReadAll(out.Body)
			closeErr := out.Body.Close()
			if readErr != nil {
				return nil, fmt.Errorf("read install stack outputs object: %w", readErr)
			}
			if closeErr != nil {
				return nil, closeErr
			}
			return data, nil
		}
		lastErr = err
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("install stack outputs did not appear at %s within %s: %w", uri, stackOutputsPollTimeout, lastErr)
		}
		logger.Info("waiting for install stack outputs object",
			zap.String("uri", uri),
			zap.Duration("retry_in", stackOutputsPollInterval),
			zap.Error(err),
		)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(stackOutputsPollInterval):
		}
	}
}

func (s *airgapS3Sync) downloadPrefix(ctx context.Context, dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if err := removeTerraformLocks(dir); err != nil {
		return err
	}
	var token *string
	for {
		out, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: &s.bucket, Prefix: &s.prefix, ContinuationToken: token})
		if err != nil {
			return err
		}
		for _, object := range out.Contents {
			if object.Key == nil || strings.HasSuffix(*object.Key, "/") || *object.Key == s.objectKey("DONE") || *object.Key == s.objectKey(airgapLeaseObject) {
				continue
			}
			rel := strings.TrimPrefix(*object.Key, s.prefix)
			rel = strings.TrimPrefix(rel, "/")
			if strings.HasPrefix(rel, "dispatch/") || isTerraformLock(rel) {
				continue
			}
			if rel == "" || filepath.IsAbs(rel) || strings.HasPrefix(filepath.Clean(rel), "..") {
				return fmt.Errorf("unsafe state object key %q", *object.Key)
			}
			out, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: &s.bucket, Key: object.Key})
			if err != nil {
				return err
			}
			path := filepath.Join(dir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				out.Body.Close()
				return err
			}
			f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
			if err != nil {
				out.Body.Close()
				return err
			}
			_, copyErr := io.Copy(f, out.Body)
			closeBodyErr := out.Body.Close()
			closeFileErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeBodyErr != nil {
				return closeBodyErr
			}
			if closeFileErr != nil {
				return closeFileErr
			}
		}
		if !aws.ToBool(out.IsTruncated) || out.NextContinuationToken == nil {
			return nil
		}
		token = out.NextContinuationToken
	}
}

type uploadFile struct {
	path string
	key  string
}

func collectUploadFiles(dir, prefix string) ([]uploadFile, error) {
	var files []uploadFile
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if name == "DONE" || name == airgapLeaseObject || strings.HasPrefix(name, "dispatch/") || isTerraformLock(name) {
			return nil
		}
		files = append(files, uploadFile{path: path, key: joinS3Key(prefix, name)})
		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	return files, err
}

func isTerraformLock(name string) bool {
	return strings.HasPrefix(name, "tfstate/") && strings.HasSuffix(name, ".json.lock")
}

func removeTerraformLocks(dir string) error {
	locks, err := filepath.Glob(filepath.Join(dir, "tfstate", "*.json.lock"))
	if err != nil {
		return err
	}
	for _, lock := range locks {
		if err := os.Remove(lock); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (s *airgapS3Sync) uploadDir(ctx context.Context, dir string) error {
	files, err := collectUploadFiles(dir, s.prefix)
	if err != nil {
		return err
	}
	for _, file := range files {
		f, err := os.Open(file.path)
		if err != nil {
			return err
		}
		_, putErr := s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: &s.bucket, Key: &file.key, Body: f})
		closeErr := f.Close()
		if putErr != nil {
			return putErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func (s *airgapS3Sync) uploadSubdir(ctx context.Context, dir, rel string) error {
	files, err := collectUploadFiles(dir, joinS3Key(s.prefix, rel))
	if err != nil {
		return err
	}
	for _, file := range files {
		f, err := os.Open(file.path)
		if err != nil {
			return err
		}
		_, putErr := s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: &s.bucket, Key: &file.key, Body: f})
		closeErr := f.Close()
		if putErr != nil {
			return putErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func (s *airgapS3Sync) writeDone(ctx context.Context, result string) error {
	key := s.objectKey("DONE")
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: &s.bucket, Key: &key, Body: strings.NewReader(result)})
	return err
}

func (s *airgapS3Sync) syncLoop(ctx context.Context, dir string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.uploadDir(ctx, dir)
		}
	}
}

func (s *airgapS3Sync) objectKey(name string) string { return joinS3Key(s.prefix, name) }

func joinS3Key(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(name, "/")
}
