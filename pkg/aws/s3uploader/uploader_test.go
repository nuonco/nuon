package s3uploader

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type recordingUploader struct {
	calls      int
	bodies     [][]byte
	err        error
	started    chan time.Time
	release    chan struct{}
	finishedAt time.Time
}

func (u *recordingUploader) Upload(ctx context.Context, input *s3.PutObjectInput, _ ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
	u.calls++
	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	u.bodies = append(u.bodies, body)
	if u.started != nil {
		u.started <- time.Now()
		select {
		case <-u.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	u.finishedAt = time.Now()
	return &manager.UploadOutput{}, u.err
}

func TestUploadStreamReusesConfiguredUploader(t *testing.T) {
	client := &recordingUploader{}
	uploader, err := NewS3Uploader(
		validator.New(),
		WithBucketName("test-bucket"),
		WithUploader(client),
	)
	require.NoError(t, err)

	checksumOne, err := uploader.UploadStream(context.Background(), bytes.NewBufferString("one"), "one.txt")
	require.NoError(t, err)
	checksumTwo, err := uploader.UploadStream(context.Background(), bytes.NewBufferString("two"), "two.txt")
	require.NoError(t, err)

	require.Equal(t, 2, client.calls)
	require.Equal(t, [][]byte{[]byte("one"), []byte("two")}, client.bodies)
	require.Equal(t, "sha256:7692c3ad3540bb803c020b3aee66cd8887123234ea0c6e7143c0add73ff431ed", checksumOne)
	require.Equal(t, "sha256:3fc4ccfe745870e2c0d99f71f30ff0656c8dedd41cc1d7d3d376b0dbe685e2f3", checksumTwo)
}

func TestUploadSpanParentsAndBoundsClientUpload(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		require.NoError(t, provider.Shutdown(context.Background()))
	})

	client := &recordingUploader{
		started: make(chan time.Time, 1),
		release: make(chan struct{}),
	}
	uploader, err := NewS3Uploader(
		validator.New(),
		WithBucketName("test-bucket"),
		WithUploader(client),
	)
	require.NoError(t, err)

	ctx, parent := provider.Tracer("test").Start(context.Background(), "gorm.create")
	done := make(chan error, 1)
	go func() {
		_, err := uploader.UploadStream(ctx, bytes.NewBufferString("payload"), "object-key")
		done <- err
	}()

	clientStartedAt := <-client.started
	require.Empty(t, recorder.Ended())
	close(client.release)
	require.NoError(t, <-done)
	parent.End()

	var uploadSpan sdktrace.ReadOnlySpan
	for _, span := range recorder.Ended() {
		if span.Name() == "s3.upload" {
			uploadSpan = span
			break
		}
	}
	require.NotNil(t, uploadSpan)
	require.Equal(t, parent.SpanContext().SpanID(), uploadSpan.Parent().SpanID())
	require.False(t, uploadSpan.StartTime().After(clientStartedAt))
	require.False(t, uploadSpan.EndTime().Before(client.finishedAt))
	require.Empty(t, uploadSpan.Attributes())
}

func TestUploadSpanRecordsErrors(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		require.NoError(t, provider.Shutdown(context.Background()))
	})

	uploadErr := errors.New("upload failed")
	uploader, err := NewS3Uploader(
		validator.New(),
		WithBucketName("test-bucket"),
		WithUploader(&recordingUploader{err: uploadErr}),
	)
	require.NoError(t, err)

	_, err = uploader.UploadStream(context.Background(), bytes.NewBufferString("payload"), "object-key")
	require.ErrorIs(t, err, uploadErr)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, codes.Error, spans[0].Status().Code)
	require.Len(t, spans[0].Events(), 1)
	require.Equal(t, "exception", spans[0].Events()[0].Name)
	require.Empty(t, spans[0].Attributes())
}
