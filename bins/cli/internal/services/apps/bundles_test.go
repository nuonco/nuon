package apps

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type bundleTestClient struct {
	nuon.Client
	get   func(context.Context) (*models.ServiceBundleResponse, error)
	grant func() (*models.ServiceDownloadGrantResponse, error)
}

func (c bundleTestClient) GetAppBundle(ctx context.Context, _, _ string) (*models.ServiceBundleResponse, error) {
	return c.get(ctx)
}

func (c bundleTestClient) CreateAppBundleDownloadGrant(context.Context, string, string) (*models.ServiceDownloadGrantResponse, error) {
	return c.grant()
}

func TestWaitBundle(t *testing.T) {
	for name, tc := range map[string]struct {
		statuses []string
		err      string
	}{
		"publish": {statuses: []string{"queued", "publishing", "active"}},
		"failed":  {statuses: []string{"error"}, err: "publication failed: acme failure"},
		"unknown": {statuses: []string{"unknown"}, err: "unknown status"},
		"timeout": {statuses: []string{"publishing"}, err: "publication continues"},
	} {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				calls := 0
				s := &Service{api: bundleTestClient{get: func(context.Context) (*models.ServiceBundleResponse, error) {
					idx := min(calls, len(tc.statuses)-1)
					calls++
					return &models.ServiceBundleResponse{ID: "abb1", Status: tc.statuses[idx], StatusDescription: "acme failure"}, nil
				}}}
				bundle, err := s.waitBundle(context.Background(), "app1", "abb1", 5*time.Second, func(*models.ServiceBundleResponse) {})
				if tc.err != "" {
					require.ErrorContains(t, err, tc.err)
				} else {
					require.NoError(t, err)
					require.Equal(t, "active", bundle.Status)
					require.Equal(t, 3, calls)
				}
			})
		})
	}
}

func TestWaitBundleCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := &Service{api: bundleTestClient{get: func(ctx context.Context) (*models.ServiceBundleResponse, error) { return nil, ctx.Err() }}}
	_, err := s.waitBundle(ctx, "app1", "abb1", time.Minute, func(*models.ServiceBundleResponse) {})
	require.ErrorIs(t, err, context.Canceled)
}

func TestDownloadBundleArchive(t *testing.T) {
	data := []byte("acme portable bundle archive")
	checksum := fmt.Sprintf("%x", sha256.Sum256(data))
	for name, tc := range map[string]struct {
		offset      int
		existing    bool
		force       bool
		expired     bool
		ignoreRange bool
		corrupt     bool
		wrongRange  bool
		wrongGrant  bool
		wrongPart   bool
		err         string
	}{
		"fresh":                  {},
		"resume":                 {offset: 5},
		"complete partial":       {offset: len(data)},
		"refresh":                {offset: 5, expired: true},
		"range ignored restarts": {offset: 5, ignoreRange: true},
		"overwrite denied":       {existing: true, err: "already exists"},
		"force":                  {existing: true, force: true},
		"checksum":               {corrupt: true, err: "SHA-256 verification failed"},
		"invalid range":          {offset: 5, wrongRange: true, err: "unexpected status or range"},
		"grant changed":          {wrongGrant: true, err: "pinned bundle"},
		"partial identity":       {offset: 5, wrongPart: true, err: "different bundle"},
	} {
		t.Run(name, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "acme.tar.zst")
			bundle := &models.ServiceBundleResponse{ID: "abb1", AppID: "app1", Size: int64(len(data)), TransportChecksum: checksum}
			if tc.existing {
				require.NoError(t, os.WriteFile(destination, []byte("existing"), 0600))
			}
			if tc.offset > 0 {
				identity := bundleDownloadIdentity{AppID: bundle.AppID, BundleID: bundle.ID, Size: bundle.Size, Checksum: "sha256:" + checksum}
				if tc.wrongPart {
					identity.BundleID = "abb-other"
				}
				file, err := openBundlePartial(destination+".part", destination+".part.json", identity)
				require.NoError(t, err)
				_, err = file.Write(data[:tc.offset])
				require.NoError(t, err)
				require.NoError(t, file.Close())
			}
			requests, grants := 0, 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				require.Empty(t, r.Header.Get("Authorization"))
				require.Empty(t, r.Header.Get("X-Nuon-Org-ID"))
				if tc.expired && requests == 1 {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				offset := 0
				if tc.offset > 0 {
					require.Equal(t, fmt.Sprintf("bytes=%d-", tc.offset), r.Header.Get("Range"))
					if !tc.ignoreRange {
						offset = tc.offset
						start := offset
						if tc.wrongRange {
							start++
						}
						w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, len(data)-1, len(data)))
						w.WriteHeader(http.StatusPartialContent)
					}
				}
				body := append([]byte(nil), data[offset:]...)
				if tc.corrupt {
					body[0] = 'X'
				}
				_, _ = w.Write(body)
			}))
			defer server.Close()
			api := bundleTestClient{grant: func() (*models.ServiceDownloadGrantResponse, error) {
				grants++
				grant := &models.ServiceDownloadGrantResponse{URL: server.URL + "?secret=never-print", Size: bundle.Size, TransportChecksum: checksum, SupportsRange: true}
				if tc.wrongGrant {
					grant.Size++
				}
				return grant, nil
			}}
			result, err := downloadBundleArchive(context.Background(), api, server.Client(), bundle, destination, tc.force, func(int64, int64) {})
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				require.NotContains(t, err.Error(), "never-print")
				if tc.existing {
					got, err := os.ReadFile(destination)
					require.NoError(t, err)
					require.Equal(t, "existing", string(got))
				} else {
					require.NoFileExists(t, destination)
				}
				return
			}
			if tc.expired {
				require.Equal(t, 2, grants)
			}
			if tc.offset == len(data) {
				require.Zero(t, requests)
			}
			require.NoError(t, err)
			require.True(t, result.Verified)
			got, err := os.ReadFile(destination)
			require.NoError(t, err)
			require.Equal(t, data, got)
			require.NoFileExists(t, destination+".part")
			require.NoFileExists(t, destination+".part.json")
		})
	}
}

func TestDownloadInterruptionResumes(t *testing.T) {
	data := []byte("acme interrupted archive")
	checksum := fmt.Sprintf("%x", sha256.Sum256(data))
	requests := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			w.Header().Set("Content-Length", fmt.Sprint(len(data)))
			_, _ = w.Write(data[:5])
			return
		}
		require.Equal(t, "bytes=5-", r.Header.Get("Range"))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes 5-%d/%d", len(data)-1, len(data)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(data[5:])
	}))
	defer server.Close()
	api := bundleTestClient{grant: func() (*models.ServiceDownloadGrantResponse, error) {
		return &models.ServiceDownloadGrantResponse{URL: server.URL, Size: int64(len(data)), TransportChecksum: checksum, SupportsRange: true}, nil
	}}
	bundle := &models.ServiceBundleResponse{ID: "abb1", AppID: "app1", Size: int64(len(data)), TransportChecksum: checksum}
	destination := filepath.Join(t.TempDir(), "acme.tar.zst")
	_, err := downloadBundleArchive(context.Background(), api, server.Client(), bundle, destination, false, func(int64, int64) {})
	require.ErrorContains(t, err, "interrupted")
	require.NoFileExists(t, destination)
	partial, err := os.ReadFile(destination + ".part")
	require.NoError(t, err)
	require.Equal(t, data[:5], partial)
	_, err = downloadBundleArchive(context.Background(), api, server.Client(), bundle, destination, false, func(int64, int64) {})
	require.NoError(t, err)
	got, err := os.ReadFile(destination)
	require.NoError(t, err)
	require.Equal(t, data, got)
}
