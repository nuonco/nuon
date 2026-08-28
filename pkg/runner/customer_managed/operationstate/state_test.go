package operationstate

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"
)

func TestLocalPutIfAbsent(t *testing.T) {
	store := NewLocal(t.TempDir())
	require.NoError(t, store.PutIfAbsent(context.Background(), "x/y.json", []byte("one")))
	require.ErrorIs(t, store.PutIfAbsent(context.Background(), "x/y.json", []byte("two")), ErrObjectExists)
	raw, ok, err := store.Get(context.Background(), "x/y.json")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []byte("one"), raw)
}

type preconditionError struct{}

func (preconditionError) Error() string                 { return "precondition failed" }
func (preconditionError) ErrorCode() string             { return "PreconditionFailed" }
func (preconditionError) ErrorMessage() string          { return "precondition failed" }
func (preconditionError) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

type fakeS3 struct {
	put *s3.PutObjectInput
}

func (f *fakeS3) GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return nil, errors.New("unused")
}

func (f *fakeS3) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.put = input
	return nil, preconditionError{}
}

func (f *fakeS3) ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return nil, errors.New("unused")
}

func TestS3PutIfAbsentUsesConditionalRequest(t *testing.T) {
	client := &fakeS3{}
	store := NewS3(client, "bucket", "prefix")
	err := store.PutIfAbsent(context.Background(), "dispatch/requests/id.json", []byte(`{"id":"one"}`))
	require.ErrorIs(t, err, ErrObjectExists)
	require.Equal(t, "*", *client.put.IfNoneMatch)
	require.Equal(t, "prefix/dispatch/requests/id.json", *client.put.Key)
	raw, readErr := io.ReadAll(client.put.Body)
	require.NoError(t, readErr)
	require.JSONEq(t, `{"id":"one"}`, string(raw))
}
