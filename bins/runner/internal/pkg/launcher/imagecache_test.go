package launcher

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

// newTestCache returns a cache rooted in a temp dir, with docker calls recorded
// instead of executed and a fixed amount of free space.
func newTestCache(t *testing.T, freeSpace uint64) (*ImageCache, *[]string) {
	t.Helper()

	var removed []string
	c := &ImageCache{
		root: t.TempDir(),
		docker: func(_ context.Context, args ...string) ([]byte, error) {
			if len(args) > 0 && args[0] == "info" {
				return nil, nil
			}
			removed = append(removed, strings.Join(args, " "))
			return nil, nil
		},
		free: func(string) (uint64, error) { return freeSpace, nil },
	}

	return c, &removed
}

// age backdates an image record so it looks least-recently-used.
func age(t *testing.T, c *ImageCache, image string, d time.Duration) {
	t.Helper()

	when := time.Now().Add(-d)
	if err := os.Chtimes(c.recordPath(image), when, when); err != nil {
		t.Fatalf("unable to backdate record: %v", err)
	}
}

func TestCollectGarbage(t *testing.T) {
	l := zap.NewNop()
	plentyOfSpace := minFreeBytes * 2

	t.Run("retains recent images when the volume has room", func(t *testing.T) {
		c, removed := newTestCache(t, plentyOfSpace)
		if err := c.Record("repo@sha256:aaa"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		c.CollectGarbage(context.Background(), l)

		if len(*removed) != 0 {
			t.Fatalf("expected no removals, got %v", *removed)
		}
		if _, err := os.Stat(c.recordPath("repo@sha256:aaa")); err != nil {
			t.Fatalf("expected record to survive: %v", err)
		}
	})

	t.Run("removes images under disk pressure", func(t *testing.T) {
		c, removed := newTestCache(t, minFreeBytes/2)
		if err := c.Record("repo@sha256:aaa"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		c.CollectGarbage(context.Background(), l)

		if len(*removed) != 1 || !strings.Contains((*removed)[0], "repo@sha256:aaa") {
			t.Fatalf("expected the image to be removed, got %v", *removed)
		}
		// no -f, so docker refuses while a container still uses the image
		if strings.Contains((*removed)[0], "-f") {
			t.Fatalf("removal must not force: %v", (*removed)[0])
		}
		if _, err := os.Stat(c.recordPath("repo@sha256:aaa")); !os.IsNotExist(err) {
			t.Fatalf("expected record to be pruned, got %v", err)
		}
	})

	t.Run("never removes a leased image", func(t *testing.T) {
		c, removed := newTestCache(t, minFreeBytes/2)
		if err := c.Record("repo@sha256:leased"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := c.Lease("exec-1", "repo@sha256:leased"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		c.CollectGarbage(context.Background(), l)

		if len(*removed) != 0 {
			t.Fatalf("expected leased image to be retained, got %v", *removed)
		}
	})

	t.Run("collects a leased image once released", func(t *testing.T) {
		c, removed := newTestCache(t, minFreeBytes/2)
		if err := c.Record("repo@sha256:aaa"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := c.Lease("exec-1", "repo@sha256:aaa"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		c.Unlease("exec-1")

		c.CollectGarbage(context.Background(), l)

		if len(*removed) != 1 {
			t.Fatalf("expected released image to be collected, got %v", *removed)
		}
	})

	t.Run("reaps a lease left by a crashed process", func(t *testing.T) {
		c, removed := newTestCache(t, minFreeBytes/2)
		if err := c.Record("repo@sha256:aaa"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := c.Lease("exec-crashed", "repo@sha256:aaa"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		when := time.Now().Add(-2 * maxLeaseAge)
		if err := os.Chtimes(c.leasePath("exec-crashed"), when, when); err != nil {
			t.Fatalf("unable to backdate lease: %v", err)
		}

		c.CollectGarbage(context.Background(), l)

		if len(*removed) != 1 {
			t.Fatalf("expected stale lease to stop pinning the image, got %v", *removed)
		}
		if _, err := os.Stat(c.leasePath("exec-crashed")); !os.IsNotExist(err) {
			t.Fatalf("expected stale lease to be reaped, got %v", err)
		}
	})

	t.Run("retires idle images even with free space", func(t *testing.T) {
		c, removed := newTestCache(t, plentyOfSpace)
		if err := c.Record("repo@sha256:cold"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		age(t, c, "repo@sha256:cold", 2*maxImageIdle)

		c.CollectGarbage(context.Background(), l)

		if len(*removed) != 1 || !strings.Contains((*removed)[0], "repo@sha256:cold") {
			t.Fatalf("expected the idle image to be retired, got %v", *removed)
		}
	})

	t.Run("stops once back above the free-space floor", func(t *testing.T) {
		free := minFreeBytes / 2
		var removed []string
		c := &ImageCache{
			root: t.TempDir(),
			docker: func(_ context.Context, args ...string) ([]byte, error) {
				if len(args) > 0 && args[0] == "info" {
					return nil, nil
				}
				removed = append(removed, strings.Join(args, " "))
				// the first removal frees enough to clear the floor
				free = minFreeBytes * 2
				return nil, nil
			},
			free: func(string) (uint64, error) { return free, nil },
		}

		for _, image := range []string{"repo@sha256:oldest", "repo@sha256:newer"} {
			if err := c.Record(image); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
		age(t, c, "repo@sha256:oldest", time.Hour)

		c.CollectGarbage(context.Background(), l)

		if len(removed) != 1 {
			t.Fatalf("expected collection to stop after one removal, got %v", removed)
		}
		// least-recently-used goes first
		if !strings.Contains(removed[0], "repo@sha256:oldest") {
			t.Fatalf("expected the coldest image to go first, got %v", removed[0])
		}
	})

	t.Run("keeps the record when docker refuses", func(t *testing.T) {
		c := &ImageCache{
			root: t.TempDir(),
			docker: func(_ context.Context, _ ...string) ([]byte, error) {
				return []byte("image is being used by running container abc"), os.ErrPermission
			},
			free: func(string) (uint64, error) { return minFreeBytes / 2, nil },
		}
		if err := c.Record("repo@sha256:inuse"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		c.CollectGarbage(context.Background(), l)

		if _, err := os.Stat(c.recordPath("repo@sha256:inuse")); err != nil {
			t.Fatalf("expected record to be kept so the image stays tracked: %v", err)
		}
	})

	t.Run("prunes the record when the image is already gone", func(t *testing.T) {
		c := &ImageCache{
			root: t.TempDir(),
			docker: func(_ context.Context, _ ...string) ([]byte, error) {
				return []byte("Error response from daemon: No such image: repo@sha256:gone"), os.ErrNotExist
			},
			free: func(string) (uint64, error) { return minFreeBytes / 2, nil },
		}
		if err := c.Record("repo@sha256:gone"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		c.CollectGarbage(context.Background(), l)

		if _, err := os.Stat(c.recordPath("repo@sha256:gone")); !os.IsNotExist(err) {
			t.Fatalf("expected the stale record to be pruned, got %v", err)
		}
	})
}

func TestImageCacheLease(t *testing.T) {
	c, _ := newTestCache(t, minFreeBytes)

	if err := c.Lease("", "repo@sha256:aaa"); err == nil {
		t.Fatal("expected an error for an empty lease id")
	}
	if err := c.Lease("exec-1", ""); err == nil {
		t.Fatal("expected an error for an empty image")
	}

	// Unlease is idempotent so cleanup can call it unconditionally
	c.Unlease("never-leased")
	c.Unlease("")
}

func TestImageCacheLock(t *testing.T) {
	t.Run("is reusable once released", func(t *testing.T) {
		c, _ := newTestCache(t, minFreeBytes)

		unlock, err := c.Lock(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		unlock()

		unlock, err = c.Lock(context.Background())
		if err != nil {
			t.Fatalf("unexpected error on relock: %v", err)
		}
		unlock()
	})

	t.Run("gives up instead of waiting on a held lock", func(t *testing.T) {
		c, _ := newTestCache(t, minFreeBytes)

		unlock, err := c.Lock(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer unlock()

		// A holder stuck mid-pull must never wedge the caller, which on the
		// collection side is the job loop's own goroutine.
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		start := time.Now()
		if _, err := c.Lock(ctx); err == nil {
			t.Fatal("expected the second attempt to fail while the lock is held")
		}
		if waited := time.Since(start); waited > 5*time.Second {
			t.Fatalf("expected a bounded wait, waited %s", waited)
		}
	})

	t.Run("collection skips a pass it cannot lock", func(t *testing.T) {
		c, removed := newTestCache(t, minFreeBytes/2)
		if err := c.Record("repo@sha256:aaa"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		unlock, err := c.Lock(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer unlock()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		c.CollectGarbage(ctx, zap.NewNop())

		if len(*removed) != 0 {
			t.Fatalf("expected collection to skip the pass, got %v", *removed)
		}
		if _, err := os.Stat(c.recordPath("repo@sha256:aaa")); err != nil {
			t.Fatalf("expected the record to survive a skipped pass: %v", err)
		}
	})
}
