package apps

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const (
	blobGrantBatchSize    = 100
	blobDownloadWorkers   = 8
	maxBlobDownloadBytes  = int64(5 << 30)
	maxIndexDownloadBytes = int64(100 << 20)
)

type PullBundleOptions struct {
	File      string
	CacheDir  string
	Overwrite bool
}

type blobGrantFunc func(ctx context.Context, digests []string) (*models.ServiceBlobGrantsResponse, error)

type pullStats struct {
	TotalBlobs      int
	TotalBytes      int64
	CachedBlobs     int
	CachedBytes     int64
	DownloadedBlobs int
	DownloadedBytes int64
}

func (s *Service) PullBundle(ctx context.Context, appID, bundleID string, opts PullBundleOptions) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	grants := func(ctx context.Context, digests []string) (*models.ServiceBlobGrantsResponse, error) {
		return s.api.CreateReleasePackageBlobGrants(ctx, bundleID, digests)
	}
	stats, err := pullBundle(ctx, http.DefaultClient, grants, opts)
	if err != nil {
		return ui.PrintError(err)
	}
	ui.PrintLn(fmt.Sprintf("blobs: %d total (%s), %d cached (%s), %d downloaded (%s)",
		stats.TotalBlobs, formatBytes(stats.TotalBytes),
		stats.CachedBlobs, formatBytes(stats.CachedBytes),
		stats.DownloadedBlobs, formatBytes(stats.DownloadedBytes)))
	ui.PrintSuccess(fmt.Sprintf("pulled bundle to %s (checksum verified)", opts.File))
	return nil
}

func pullBundle(ctx context.Context, client *http.Client, grants blobGrantFunc, opts PullBundleOptions) (pullStats, error) {
	var stats pullStats
	if opts.File == "" {
		return stats, fmt.Errorf("destination file path is required")
	}
	if _, err := os.Stat(opts.File); err == nil && !opts.Overwrite {
		return stats, fmt.Errorf("destination already exists; use --overwrite to replace it")
	} else if err != nil && !os.IsNotExist(err) {
		return stats, fmt.Errorf("inspect destination: %w", err)
	}
	cacheDir, err := resolveBlobCacheDir(opts.CacheDir)
	if err != nil {
		return stats, err
	}
	cache := blobCache{dir: cacheDir}

	meta, err := grants(ctx, nil)
	if err != nil {
		return stats, fmt.Errorf("fetch bundle metadata: %w", err)
	}
	indexDigest, err := parseSHA256Digest(meta.OciIndexDigest)
	if err != nil {
		return stats, fmt.Errorf("bundle metadata has invalid OCI index digest: %w", err)
	}
	expectedChecksum := strings.TrimPrefix(meta.TransportChecksum, "sha256:")
	if len(expectedChecksum) != sha256.Size*2 {
		return stats, fmt.Errorf("bundle metadata has malformed transport checksum")
	}

	indexDesc := ocispec.Descriptor{MediaType: ocispec.MediaTypeImageIndex, Digest: indexDigest, Size: -1}
	if !cache.has(indexDesc) {
		if err := downloadBlobs(ctx, client, grants, cache, []ocispec.Descriptor{indexDesc}, maxIndexDownloadBytes); err != nil {
			return stats, err
		}
	}
	indexJSON, err := cache.read(indexDigest)
	if err != nil {
		return stats, err
	}
	frontier, err := bundle.Successors(ocispec.MediaTypeImageIndex, indexJSON)
	if err != nil {
		return stats, fmt.Errorf("parse bundle index: %w", err)
	}

	seen := map[digest.Digest]struct{}{indexDigest: {}}
	for len(frontier) > 0 {
		level := make([]ocispec.Descriptor, 0, len(frontier))
		for _, desc := range frontier {
			if _, ok := seen[desc.Digest]; ok {
				continue
			}
			if _, err := parseSHA256Digest(desc.Digest.String()); err != nil {
				return stats, fmt.Errorf("bundle references unsupported blob digest %q", desc.Digest)
			}
			seen[desc.Digest] = struct{}{}
			level = append(level, desc)
			stats.TotalBlobs++
			stats.TotalBytes += desc.Size
		}
		missing := make([]ocispec.Descriptor, 0, len(level))
		for _, desc := range level {
			if cache.has(desc) {
				stats.CachedBlobs++
				stats.CachedBytes += desc.Size
				continue
			}
			missing = append(missing, desc)
			stats.DownloadedBlobs++
			stats.DownloadedBytes += desc.Size
		}
		if err := downloadBlobs(ctx, client, grants, cache, missing, maxBlobDownloadBytes); err != nil {
			return stats, err
		}
		var next []ocispec.Descriptor
		for _, desc := range level {
			if !bundle.IsManifestMediaType(desc.MediaType) {
				continue
			}
			data, err := cache.read(desc.Digest)
			if err != nil {
				return stats, err
			}
			children, err := bundle.Successors(desc.MediaType, data)
			if err != nil {
				return stats, fmt.Errorf("traverse blob %s: %w", desc.Digest, err)
			}
			next = append(next, children...)
		}
		frontier = next
	}

	fetch := func(ctx context.Context, dgst digest.Digest) ([]byte, error) {
		data, err := cache.read(dgst)
		if err == nil && digest.FromBytes(data) == dgst {
			return data, nil
		}
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if err == nil {
			if removeErr := os.Remove(cache.path(dgst)); removeErr != nil {
				return nil, fmt.Errorf("remove corrupt cached blob %s: %w", dgst, removeErr)
			}
		}
		desc := ocispec.Descriptor{Digest: dgst, Size: -1}
		if err := downloadBlobs(ctx, client, grants, cache, []ocispec.Descriptor{desc}, maxBlobDownloadBytes); err != nil {
			return nil, err
		}
		return cache.read(dgst)
	}

	partialPath := opts.File + ".partial"
	partial, err := os.OpenFile(partialPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return stats, fmt.Errorf("open partial bundle: %w", err)
	}
	checksum, reassembleErr := bundle.Reassemble(ctx, partial, indexJSON, fetch)
	closeErr := partial.Close()
	if reassembleErr != nil {
		os.Remove(partialPath)
		return stats, fmt.Errorf("reassemble bundle: %w", reassembleErr)
	}
	if closeErr != nil {
		os.Remove(partialPath)
		return stats, fmt.Errorf("close partial bundle: %w", closeErr)
	}
	if !strings.EqualFold(checksum, expectedChecksum) {
		os.Remove(partialPath)
		return stats, fmt.Errorf("reassembled bundle checksum mismatch: expected sha256:%s, got sha256:%s", expectedChecksum, checksum)
	}
	if err := os.Rename(partialPath, opts.File); err != nil {
		os.Remove(partialPath)
		return stats, fmt.Errorf("commit bundle: %w", err)
	}
	return stats, nil
}

// downloadBlobs requests grants for the given descriptors in batches and
// downloads them concurrently into the cache, verifying each blob's digest.
func downloadBlobs(ctx context.Context, client *http.Client, grants blobGrantFunc, cache blobCache, descs []ocispec.Descriptor, maxBytes int64) error {
	if len(descs) == 0 {
		return nil
	}
	sizes := make(map[digest.Digest]int64, len(descs))
	digests := make([]string, 0, len(descs))
	for _, desc := range descs {
		digests = append(digests, desc.Digest.String())
		sizes[desc.Digest] = desc.Size
	}
	for start := 0; start < len(digests); start += blobGrantBatchSize {
		end := min(start+blobGrantBatchSize, len(digests))
		resp, err := grants(ctx, digests[start:end])
		if err != nil {
			return fmt.Errorf("request blob grants: %w", err)
		}
		granted := map[string]*models.ServiceBlobGrantItem{}
		for _, grant := range resp.Grants {
			granted[grant.Digest] = grant
		}
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(blobDownloadWorkers)
		for _, requested := range digests[start:end] {
			grant, ok := granted[requested]
			if !ok {
				return fmt.Errorf("no grant returned for blob %s", requested)
			}
			g.Go(func() error {
				dgst, err := parseSHA256Digest(grant.Digest)
				if err != nil {
					return err
				}
				expectedSize := sizes[dgst]
				if expectedSize >= 0 && grant.Size >= 0 && grant.Size != expectedSize {
					return fmt.Errorf("blob %s size mismatch: bundle declares %d bytes, store has %d", dgst, expectedSize, grant.Size)
				}
				return downloadBlob(gctx, client, cache, dgst, grant.URL, maxBytes)
			})
		}
		if err := g.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func downloadBlob(ctx context.Context, client *http.Client, cache blobCache, dgst digest.Digest, url string, maxBytes int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create blob download request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("blob download failed: %w", ctx.Err())
		}
		return fmt.Errorf("blob %s download failed", dgst)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("blob %s download returned unexpected HTTP status %d", dgst, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return fmt.Errorf("blob %s download interrupted: %w", dgst, err)
	}
	if int64(len(data)) > maxBytes {
		return fmt.Errorf("blob %s exceeds %d byte limit", dgst, maxBytes)
	}
	if computed := digest.FromBytes(data); computed != dgst {
		return fmt.Errorf("blob digest mismatch: expected %s, got %s", dgst, computed)
	}
	return cache.put(dgst, data)
}

type blobCache struct {
	dir string
}

func (c blobCache) path(dgst digest.Digest) string {
	return filepath.Join(c.dir, "sha256", dgst.Encoded())
}

func (c blobCache) has(desc ocispec.Descriptor) bool {
	info, err := os.Stat(c.path(desc.Digest))
	if err != nil {
		return false
	}
	return desc.Size < 0 || info.Size() == desc.Size
}

func (c blobCache) read(dgst digest.Digest) ([]byte, error) {
	return os.ReadFile(c.path(dgst))
}

func (c blobCache) put(dgst digest.Digest, data []byte) error {
	path := c.path(dgst)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create blob cache directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+dgst.Encoded()+".tmp-*")
	if err != nil {
		return fmt.Errorf("create blob cache temp file: %w", err)
	}
	_, writeErr := tmp.Write(data)
	closeErr := tmp.Close()
	if writeErr != nil || closeErr != nil {
		os.Remove(tmp.Name())
		if writeErr != nil {
			return fmt.Errorf("write cached blob %s: %w", dgst, writeErr)
		}
		return fmt.Errorf("close cached blob %s: %w", dgst, closeErr)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("commit cached blob %s: %w", dgst, err)
	}
	return nil
}

func resolveBlobCacheDir(dir string) (string, error) {
	if dir == "" {
		home, err := homedir.Dir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory for blob cache: %w", err)
		}
		return filepath.Join(home, ".config", "nuon", "bundle-blobs"), nil
	}
	expanded, err := homedir.Expand(dir)
	if err != nil {
		return "", fmt.Errorf("resolve blob cache directory: %w", err)
	}
	return expanded, nil
}

func parseSHA256Digest(value string) (digest.Digest, error) {
	dgst, err := digest.Parse(value)
	if err != nil {
		return "", fmt.Errorf("invalid blob digest %q: %w", value, err)
	}
	if dgst.Algorithm() != digest.SHA256 {
		return "", fmt.Errorf("unsupported blob digest algorithm %q", dgst.Algorithm())
	}
	return dgst, nil
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
