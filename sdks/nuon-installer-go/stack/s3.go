package stack

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// CreateS3Bucket creates the artifact bucket if it doesn't exist.
func CreateS3Bucket(ctx context.Context, log *slog.Logger, c *s3.Client, st *State) error {
	if st.S3BucketName == "" {
		st.S3BucketName = fmt.Sprintf("nuon-install-%s", st.InstallID)
	}
	log = log.With("bucket", st.S3BucketName)

	in := &s3.CreateBucketInput{Bucket: aws.String(st.S3BucketName)}
	if st.Region != "us-east-1" {
		in.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(st.Region),
		}
	}
	if _, err := c.CreateBucket(ctx, in); err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			switch ae.ErrorCode() {
			case "BucketAlreadyOwnedByYou", "BucketAlreadyExists":
				log.Info("bucket already exists, reusing")
				return st.Save()
			}
		}
		return fmt.Errorf("create bucket: %w", err)
	}
	log.Info("created bucket")
	return st.Save()
}

// DeleteS3Bucket empties and deletes the bucket. Idempotent.
func DeleteS3Bucket(ctx context.Context, log *slog.Logger, c *s3.Client, st *State) error {
	if st.S3BucketName == "" {
		return nil
	}
	log = log.With("bucket", st.S3BucketName)

	// Empty bucket. Best effort — small artifact buckets only.
	for {
		out, err := c.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: &st.S3BucketName})
		if err != nil {
			var ae smithy.APIError
			if errors.As(err, &ae) && ae.ErrorCode() == "NoSuchBucket" {
				st.S3BucketName = ""
				return st.Save()
			}
			return fmt.Errorf("list objects: %w", err)
		}
		if len(out.Contents) == 0 {
			break
		}
		ids := make([]s3types.ObjectIdentifier, 0, len(out.Contents))
		for _, o := range out.Contents {
			ids = append(ids, s3types.ObjectIdentifier{Key: o.Key})
		}
		if _, err := c.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &st.S3BucketName,
			Delete: &s3types.Delete{Objects: ids},
		}); err != nil {
			return fmt.Errorf("delete objects: %w", err)
		}
		if !aws.ToBool(out.IsTruncated) {
			break
		}
	}

	if _, err := c.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: &st.S3BucketName}); err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "NoSuchBucket" {
			log.Info("bucket already gone")
		} else {
			return fmt.Errorf("delete bucket: %w", err)
		}
	} else {
		log.Info("deleted bucket")
	}
	st.S3BucketName = ""
	return st.Save()
}
