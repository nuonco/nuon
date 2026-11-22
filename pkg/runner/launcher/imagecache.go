package launcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ImageCache tracks the action images this host has pulled and reclaims them
// when the volume runs low. Action images are content-addressed and shared by
// every run of the same image, so they are deliberately NOT deleted after a run
// — deleting one would re-pull on the next step and could drop an image another
// job is about to start. Collection instead runs while the action job loop is
// idle, and never touches an image a job still holds a lease on.
//
// Host layout under root:
//
//	lock         held while pulling or collecting, so the two can never overlap
//	leases/<id>  one file per in-flight job execution, holding the image ref
//	images/<key> one record per pulled image, holding the ref; mtime is last use
type ImageCache struct {
	root string

	// docker and free are fields so tests can drive collection without a docker
	// daemon or a full disk.
	docker func(ctx context.Context, args ...string) ([]byte, error)
	free   func(path string) (uint64, error)

	spaceDirOnce sync.Once
	spaceDir     string
}

const (
	defaultImageCacheRoot = "/opt/nuon/action-images"

	// minFreeBytes is the free-space floor on the cache volume. Collection runs
	// below it and stops as soon as the volume is back above it, so there is
	// always room for another multi-GiB pull and its extraction.
	minFreeBytes uint64 = 8 << 30

	// maxImageIdle retires images nothing has used in a week even when the
	// volume has room, so a long-lived runner doesn't hold images for actions
	// that were since reconfigured or deleted.
	maxImageIdle = 7 * 24 * time.Hour

	// maxLeaseAge bounds a lease so a process that died mid-job cannot pin an
	// image forever. It is far above any action timeout.
	maxLeaseAge = 24 * time.Hour

	// lockWait bounds how long a caller waits for the cache lock before giving
	// up. Collection retries on its next idle pass; a pull fails the step with a
	// clear error rather than hanging past its timeout.
	lockWait          = 30 * time.Second
	lockRetryInterval = 250 * time.Millisecond
)

func NewImageCache() *ImageCache {
	dockerPath := "docker"
	if p, err := exec.LookPath("docker"); err == nil {
		dockerPath = p
	}

	// run-local wires this loop on a developer machine, where /opt isn't
	// writable. The cache only has to be a stable path the process can write.
	root := defaultImageCacheRoot
	if err := os.MkdirAll(root, 0o755); err != nil {
		root = filepath.Join(os.TempDir(), "nuon-action-images")
	}

	return &ImageCache{
		root: root,
		docker: func(ctx context.Context, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, dockerPath, args...).CombinedOutput()
		},
		free: freeBytes,
	}
}

// Lock serializes image pulls against collection. The returned func releases it.
//
// The wait is a bounded poll rather than a blocking LOCK_EX: flock cannot
// observe context cancellation, so a holder stuck mid-pull would otherwise
// wedge whichever caller is waiting. On the collection side that caller is the
// job loop's own goroutine, which would stop claiming work with nothing to
// detect it.
func (c *ImageCache) Lock(ctx context.Context) (func(), error) {
	if err := os.MkdirAll(c.root, 0o755); err != nil {
		return nil, err
	}

	fh, err := os.OpenFile(filepath.Join(c.root, "lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, lockWait)
	defer cancel()

	for {
		err := syscall.Flock(int(fh.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() {
				syscall.Flock(int(fh.Fd()), syscall.LOCK_UN)
				fh.Close()
			}, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			fh.Close()
			return nil, err
		}

		select {
		case <-ctx.Done():
			fh.Close()
			return nil, errors.Wrap(ctx.Err(), "timed out waiting for the action image cache lock")
		case <-time.After(lockRetryInterval):
		}
	}
}

// Lease marks an image as in use by a job execution.
func (c *ImageCache) Lease(leaseID, image string) error {
	if leaseID == "" || image == "" {
		return errors.New("image lease requires both a lease id and an image")
	}
	if err := os.MkdirAll(filepath.Join(c.root, "leases"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.leasePath(leaseID), []byte(image), 0o600)
}

// Unlease drops a lease. The image stays on the host.
func (c *ImageCache) Unlease(leaseID string) {
	if leaseID == "" {
		return
	}
	_ = os.Remove(c.leasePath(leaseID))
}

// Record notes that an image is present on the host and was just used.
func (c *ImageCache) Record(image string) error {
	if image == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Join(c.root, "images"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.recordPath(image), []byte(image), 0o600)
}

// CollectGarbage removes cached action images that no job holds a lease on,
// once the volume is below its free-space floor or an image has gone unused for
// maxImageIdle. Callers should bound ctx: collection gives up rather than run
// long, since the caller is the job loop goroutine that would otherwise not be
// claiming work.
func (c *ImageCache) CollectGarbage(ctx context.Context, l *zap.Logger) {
	unlock, err := c.Lock(ctx)
	if err != nil {
		l.Warn("unable to lock action image cache for collection", zap.Error(err))
		return
	}
	defer unlock()

	probeDir := c.spaceProbeDir(ctx)

	leased := c.leasedRefs(l)

	records, err := c.records()
	if err != nil {
		l.Warn("unable to read action image cache records", zap.Error(err))
		return
	}
	if len(records) == 0 {
		return
	}

	freeBefore, err := c.free(probeDir)
	if err != nil {
		l.Warn("unable to read free space for action image cache", zap.Error(err))
		return
	}
	needSpace := freeBefore < minFreeBytes

	var removed, retained int
	for _, rec := range records {
		if ctx.Err() != nil {
			l.Info("stopping action image collection early",
				zap.Error(ctx.Err()),
				zap.Int("removed", removed),
			)
			break
		}

		if leased[rec.image] {
			retained++
			continue
		}

		expired := time.Since(rec.usedAt) > maxImageIdle
		if !expired && !needSpace {
			retained++
			continue
		}

		out, rmErr := c.docker(ctx, "image", "rm", rec.image)
		// No -f: docker refuses while a container still references the image,
		// which is exactly the outcome we want.
		if rmErr != nil && !strings.Contains(string(out), "No such image") {
			retained++
			l.Info("retained action image",
				zap.String("action.image", rec.image),
				zap.String("docker.output", strings.TrimSpace(string(out))),
			)
			continue
		}

		_ = os.Remove(rec.path)
		removed++
		l.Info("removed cached action image",
			zap.String("action.image", rec.image),
			zap.Time("last_used_at", rec.usedAt),
		)

		if needSpace {
			if free, err := c.free(probeDir); err == nil && free >= minFreeBytes {
				needSpace = false
			}
		}
	}

	freeAfter, err := c.free(probeDir)
	if err != nil {
		freeAfter = freeBefore
	}

	l.Info("collected action image cache",
		zap.Int("considered", len(records)),
		zap.Int("removed", removed),
		zap.Int("retained", retained),
		zap.Int("leased", len(leased)),
		zap.Uint64("bytes_reclaimed", freeAfter-min(freeAfter, freeBefore)),
		zap.Uint64("free_bytes", freeAfter),
	)
}

// spaceProbeDir returns the directory whose volume free space is measured
// against. Image layers live under docker's data root, which is the same volume
// as the cache on a default VM runner but not on one with a separate docker
// volume, where measuring the cache would read the wrong disk.
func (c *ImageCache) spaceProbeDir(ctx context.Context) string {
	c.spaceDirOnce.Do(func() {
		c.spaceDir = c.root

		out, err := c.docker(ctx, "info", "-f", "{{.DockerRootDir}}")
		if err != nil {
			return
		}
		if dir := strings.TrimSpace(string(out)); dir != "" {
			c.spaceDir = dir
		}
	})

	return c.spaceDir
}

type imageRecord struct {
	image  string
	path   string
	usedAt time.Time
}

// records returns the tracked images, least recently used first, so collection
// reclaims the coldest images before the ones a job is most likely to want.
func (c *ImageCache) records() ([]imageRecord, error) {
	entries, err := os.ReadDir(filepath.Join(c.root, "images"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	recs := make([]imageRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(c.root, "images", entry.Name())
		image, err := os.ReadFile(path)
		if err != nil || len(image) == 0 {
			_ = os.Remove(path)
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		recs = append(recs, imageRecord{
			image:  strings.TrimSpace(string(image)),
			path:   path,
			usedAt: info.ModTime(),
		})
	}

	sort.Slice(recs, func(i, j int) bool { return recs[i].usedAt.Before(recs[j].usedAt) })

	return recs, nil
}

// leasedRefs returns the images in-flight jobs are using, reaping leases left
// behind by a process that died mid-job.
func (c *ImageCache) leasedRefs(l *zap.Logger) map[string]bool {
	entries, err := os.ReadDir(filepath.Join(c.root, "leases"))
	if err != nil {
		return map[string]bool{}
	}

	leased := make(map[string]bool, len(entries))
	for _, entry := range entries {
		path := filepath.Join(c.root, "leases", entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if time.Since(info.ModTime()) > maxLeaseAge {
			l.Warn("reaping stale action image lease",
				zap.String("lease.id", entry.Name()),
				zap.Time("created_at", info.ModTime()),
			)
			_ = os.Remove(path)
			continue
		}
		image, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		leased[strings.TrimSpace(string(image))] = true
	}

	return leased
}

func (c *ImageCache) leasePath(leaseID string) string {
	return filepath.Join(c.root, "leases", cacheKey(leaseID))
}

func (c *ImageCache) recordPath(image string) string {
	return filepath.Join(c.root, "images", cacheKey(image))
}

// cacheKey hashes a ref or ID so it is safe to use as a filename.
func cacheKey(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:16])
}

func freeBytes(path string) (uint64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, err
	}
	return uint64(st.Bsize) * st.Bavail, nil
}
