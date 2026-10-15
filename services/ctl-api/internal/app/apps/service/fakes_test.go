package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	tclient "go.temporal.io/sdk/client"

	temporal "github.com/nuonco/nuon/pkg/temporal/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type memBlobstore struct {
	mu    sync.Mutex
	blobs map[string][]byte
}

var _ blobstore.Service = (*memBlobstore)(nil)

func newMemBlobstore() blobstore.Service {
	return &memBlobstore{blobs: map[string][]byte{}}
}

func (m *memBlobstore) Upload(ctx context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blobs[key] = append([]byte(nil), data...)
	return nil
}

func (m *memBlobstore) Download(ctx context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.blobs[key]
	if !ok {
		return nil, fmt.Errorf("blob %s not found", key)
	}
	return append([]byte(nil), data...), nil
}

func (m *memBlobstore) UploadStream(ctx context.Context, key string, r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	if err := m.Upload(ctx, key, data); err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func (m *memBlobstore) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	data, err := m.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *memBlobstore) GetMetadata(ctx context.Context, key string) (int64, string, error) {
	data, err := m.Download(ctx, key)
	if err != nil {
		return 0, "", err
	}
	return int64(len(data)), "application/octet-stream", nil
}

type mockWorkflowRun struct{}

func (m *mockWorkflowRun) GetID() string                                       { return "mock-workflow-id" }
func (m *mockWorkflowRun) GetRunID() string                                    { return "mock-run-id" }
func (m *mockWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error { return nil }
func (m *mockWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options tclient.WorkflowRunGetOptions) error {
	return nil
}

func newQueueMockTC(t *testing.T) *temporal.MockClient {
	mockTC := temporal.NewMockClient(gomock.NewController(t))
	mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(&mockWorkflowRun{}, nil).AnyTimes()
	return mockTC
}
