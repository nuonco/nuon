package day2state

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

var ErrObjectExists = errors.New("state object already exists")

type State interface {
	Get(context.Context, string) ([]byte, bool, error)
	Put(context.Context, string, []byte) error
	PutIfAbsent(context.Context, string, []byte) error
	List(context.Context, string) ([]string, error)
}

type Local struct {
	dir string
}

func NewLocal(dir string) *Local {
	return &Local{dir: dir}
}

func (s *Local) Get(_ context.Context, key string) ([]byte, bool, error) {
	raw, err := os.ReadFile(filepath.Join(s.dir, filepath.FromSlash(key)))
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return raw, err == nil, err
}

func (s *Local) PutIfAbsent(_ context.Context, key string, raw []byte) error {
	path := filepath.Join(s.dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return ErrObjectExists
	}
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(raw)
	return err
}

func (s *Local) Put(_ context.Context, key string, raw []byte) error {
	path := filepath.Join(s.dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

func (s *Local) List(_ context.Context, prefix string) ([]string, error) {
	root := filepath.Join(s.dir, filepath.FromSlash(prefix))
	var keys []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return filepath.SkipDir
		}
		if err != nil || entry.IsDir() {
			return err
		}
		rel, err := filepath.Rel(s.dir, path)
		if err == nil {
			keys = append(keys, filepath.ToSlash(rel))
		}
		return err
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	sort.Strings(keys)
	return keys, err
}

type S3API interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

type S3 struct {
	client S3API
	bucket string
	prefix string
}

func NewS3(client S3API, bucket, prefix string) *S3 {
	prefix = strings.Trim(prefix, "/")
	if prefix != "" {
		prefix += "/"
	}
	return &S3{client: client, bucket: bucket, prefix: prefix}
}

func (s *S3) key(key string) string {
	return s.prefix + key
}

func (s *S3) Get(ctx context.Context, key string) ([]byte, bool, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(s.key(key))})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return nil, false, nil
		}
		var apiError smithy.APIError
		if errors.As(err, &apiError) && (apiError.ErrorCode() == "NoSuchKey" || apiError.ErrorCode() == "NotFound") {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get s3://%s/%s: %w", s.bucket, s.key(key), err)
	}
	defer out.Body.Close()
	raw, err := io.ReadAll(out.Body)
	return raw, err == nil, err
}

func (s *S3) PutIfAbsent(ctx context.Context, key string, raw []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s.key(key)),
		Body:        bytes.NewReader(raw),
		ContentType: aws.String("application/json"),
		IfNoneMatch: aws.String("*"),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) && (apiError.ErrorCode() == "PreconditionFailed" || apiError.ErrorCode() == "ConditionalRequestConflict") {
			return ErrObjectExists
		}
		return fmt.Errorf("put s3://%s/%s: %w", s.bucket, s.key(key), err)
	}
	return nil
}

func (s *S3) Put(ctx context.Context, key string, raw []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s.key(key)),
		Body:        bytes.NewReader(raw),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("put s3://%s/%s: %w", s.bucket, s.key(key), err)
	}
	return nil
}

func (s *S3) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	var token *string
	for {
		out, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.bucket),
			Prefix:            aws.String(s.key(prefix)),
			ContinuationToken: token,
		})
		if err != nil {
			return nil, err
		}
		for _, object := range out.Contents {
			keys = append(keys, strings.TrimPrefix(aws.ToString(object.Key), s.prefix))
		}
		if !aws.ToBool(out.IsTruncated) {
			break
		}
		token = out.NextContinuationToken
	}
	return keys, nil
}

func New(ctx context.Context, state, profile, region string) (State, error) {
	if !strings.HasPrefix(state, "s3://") {
		if _, err := os.Stat(state); err != nil {
			return nil, fmt.Errorf("unable to open state directory: %w", err)
		}
		return NewLocal(state), nil
	}
	bucket, prefix, _ := strings.Cut(strings.TrimPrefix(state, "s3://"), "/")
	if bucket == "" {
		return nil, fmt.Errorf("state S3 URI bucket is required")
	}
	options := []func(*config.LoadOptions) error{}
	if profile != "" {
		options = append(options, config.WithSharedConfigProfile(profile))
	}
	if region != "" {
		options = append(options, config.WithRegion(region))
	}
	cfg, err := config.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}
	return NewS3(s3.NewFromConfig(cfg), bucket, prefix), nil
}
