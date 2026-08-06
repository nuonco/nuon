package apps

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestDownloadBundleFullTransfer(t *testing.T) {
	content := []byte("complete air-gap bundle")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("presigned request included authorization header")
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "bundle.tar")
	err := downloadBundle(context.Background(), server.Client(), testGrant(server.URL, content, true), DownloadBundleOptions{File: destination})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContent(t, destination, content)
	if _, err := os.Stat(destination + ".partial"); !os.IsNotExist(err) {
		t.Fatalf("partial file still exists: %v", err)
	}
}

func TestDownloadBundleResumesWithPartialContent(t *testing.T) {
	content := []byte("resumable bundle content")
	partialSize := 8
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRange := fmt.Sprintf("bytes=%d-", partialSize)
		if r.Header.Get("Range") != wantRange {
			t.Errorf("Range = %q, want %q", r.Header.Get("Range"), wantRange)
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", partialSize, len(content)-1, len(content)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(content[partialSize:])
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "bundle.tar")
	if err := os.WriteFile(destination+".partial", content[:partialSize], 0o600); err != nil {
		t.Fatal(err)
	}
	if err := downloadBundle(context.Background(), server.Client(), testGrant(server.URL, content, true), DownloadBundleOptions{File: destination}); err != nil {
		t.Fatal(err)
	}
	assertFileContent(t, destination, content)
}

func TestDownloadBundleRestartsAfterCorruptResume(t *testing.T) {
	content := []byte("resumable bundle content")
	partialSize := 8
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("Range") != "" {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", partialSize, len(content)-1, len(content)))
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write(content[partialSize:])
			return
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "bundle.tar")
	if err := os.WriteFile(destination+".partial", []byte("corrupt!"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := downloadBundle(context.Background(), server.Client(), testGrant(server.URL, content, true), DownloadBundleOptions{File: destination}); err != nil {
		t.Fatal(err)
	}
	assertFileContent(t, destination, content)
	if requests != 2 {
		t.Fatalf("requests = %d, want resume and one full restart", requests)
	}
}

func TestDownloadBundleRestartsWhenServerIgnoresRange(t *testing.T) {
	content := []byte("server sends the whole bundle")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") == "" {
			t.Error("expected a range request")
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "bundle.tar")
	if err := os.WriteFile(destination+".partial", []byte("server"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := downloadBundle(context.Background(), server.Client(), testGrant(server.URL, content, true), DownloadBundleOptions{File: destination}); err != nil {
		t.Fatal(err)
	}
	assertFileContent(t, destination, content)
}

func TestDownloadBundlePreservesPartialOnVerificationFailure(t *testing.T) {
	content := []byte("bundle bytes")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(content)
	}))
	defer server.Close()

	tests := []struct {
		name  string
		grant *models.ServiceDownloadGrantResponse
		match string
	}{
		{name: "wrong size", grant: testGrant(server.URL, append(content, 'x'), true), match: "size mismatch"},
		{name: "wrong checksum", grant: testGrant(server.URL, []byte("bundle BYTEs"), true), match: "checksum mismatch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "bundle.tar")
			err := downloadBundle(context.Background(), server.Client(), tt.grant, DownloadBundleOptions{File: destination})
			if err == nil || !strings.Contains(err.Error(), tt.match) {
				t.Fatalf("error = %v, want %q", err, tt.match)
			}
			assertFileContent(t, destination+".partial", content)
			if _, err := os.Stat(destination); !os.IsNotExist(err) {
				t.Fatalf("destination exists after failed verification: %v", err)
			}
		})
	}
}

func TestDownloadBundleDestinationOverwrite(t *testing.T) {
	content := []byte("new bundle")
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write(content)
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "bundle.tar")
	if err := os.WriteFile(destination, []byte("old bundle"), 0o600); err != nil {
		t.Fatal(err)
	}
	grant := testGrant(server.URL, content, false)
	if err := downloadBundle(context.Background(), server.Client(), grant, DownloadBundleOptions{File: destination}); err == nil || !strings.Contains(err.Error(), "--overwrite") {
		t.Fatalf("error = %v, want overwrite error", err)
	}
	if requests != 0 {
		t.Fatalf("made %d requests before rejecting destination", requests)
	}
	assertFileContent(t, destination, []byte("old bundle"))

	if err := downloadBundle(context.Background(), server.Client(), grant, DownloadBundleOptions{File: destination, Overwrite: true}); err != nil {
		t.Fatal(err)
	}
	assertFileContent(t, destination, content)
}

func TestVerifyAndCommitDoesNotClobberDestination(t *testing.T) {
	content := []byte("verified bundle")
	dir := t.TempDir()
	partial := filepath.Join(dir, "bundle.partial")
	destination := filepath.Join(dir, "bundle.tar")
	if err := os.WriteFile(partial, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte("racing writer"), 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	err := verifyAndCommit(partial, destination, int64(len(content)), fmt.Sprintf("%x", sum), false)
	if err == nil {
		t.Fatal("expected no-clobber commit to fail")
	}
	assertFileContent(t, destination, []byte("racing writer"))
	assertFileContent(t, partial, content)
}

func TestDownloadBundleDoesNotExposeGrantURLInErrors(t *testing.T) {
	secretURL := "http://127.0.0.1:1/bundle?secret=do-not-print"
	grant := testGrant(secretURL, []byte("bundle"), false)
	err := downloadBundle(context.Background(), &http.Client{}, grant, DownloadBundleOptions{File: filepath.Join(t.TempDir(), "bundle.tar")})
	if err == nil {
		t.Fatal("expected request error")
	}
	if strings.Contains(err.Error(), secretURL) || strings.Contains(err.Error(), "do-not-print") {
		t.Fatalf("error exposes grant URL: %v", err)
	}
}

func testGrant(url string, content []byte, supportsRange bool) *models.ServiceDownloadGrantResponse {
	checksum := sha256.Sum256(content)
	return &models.ServiceDownloadGrantResponse{
		URL:               url,
		Size:              int64(len(content)),
		SupportsRange:     supportsRange,
		TransportChecksum: fmt.Sprintf("sha256:%X", checksum),
	}
}

func assertFileContent(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("file content = %q, want %q", got, want)
	}
}
