package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const (
	sseWatchPollInterval     = 2 * time.Second
	sseTimelinePollInterval  = 3 * time.Second
	sseOrgStatusPollInterval = 5 * time.Second
	sseFinishedPollInterval  = 30 * time.Second
	sseErrorRetryDelay       = 5 * time.Second
	sseFinishedGracePeriod   = 2 * time.Minute
	sseMaxStreamLifetime     = 30 * time.Minute
)

var terminalStatuses = map[string]bool{
	"success":       true,
	"error":         true,
	"failed":        true,
	"cancelled":     true,
	"not-attempted": true,
}

// sseToken extracts the auth cookie. On failure it writes the 401 JSON
// response and returns ok=false.
func sseToken(c *gin.Context) (string, bool) {
	token, err := c.Cookie(authCookie)
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", false
	}
	return token, true
}

// sseAuth extracts the auth cookie and builds a nuon client scoped to the
// route's org. On failure it writes the JSON error response and returns
// ok=false. Must be called before runSSEStream, which flushes SSE headers.
func sseAuth(c *gin.Context, cfg *internal.Config, l *zap.Logger) (nuon.Client, string, bool) {
	token, ok := sseToken(c)
	if !ok {
		return nil, "", false
	}

	client, err := nuon.New(
		nuon.WithURL(cfg.APIUrl),
		nuon.WithAuthToken(token),
		nuon.WithOrgID(c.Param("orgId")),
	)
	if err != nil {
		l.Error("failed to create nuon client", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create client"})
		return nil, "", false
	}
	return client, token, true
}

type sseEvent struct {
	// Name is the SSE event name; empty means an unnamed bare "data:" event.
	Name string
	Data []byte
}

func marshalEvent(name string, v any) (sseEvent, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return sseEvent{}, err
	}
	return sseEvent{Name: name, Data: data}, nil
}

type sseFetchResult struct {
	// Events are emitted in order when their per-name hash changes. The
	// first event is the primary resource: finished detection keys off its
	// hash changing. Omitting an event name on a tick leaves its previous
	// hash intact, so a failed secondary fetch is skipped silently.
	Events   []sseEvent
	Finished bool
}

// errSSESilentRetry wraps fetch errors that should retry without emitting a
// fetch-error event to the client.
var errSSESilentRetry = errors.New("sse: silent retry")

type sseStreamConfig struct {
	Fetch        func(ctx context.Context) (sseFetchResult, error)
	ClientErrMsg string
	// PollInterval defaults to sseWatchPollInterval.
	PollInterval time.Duration
	// FinishedPollInterval defaults to sseFinishedPollInterval.
	FinishedPollInterval time.Duration
	// ErrorRetryDelay defaults to sseErrorRetryDelay.
	ErrorRetryDelay time.Duration
	// FinishedGracePeriod of zero means the stream never self-closes.
	FinishedGracePeriod time.Duration
	// MaxLifetime defaults to sseMaxStreamLifetime. The stream emits an
	// expired event and closes once exceeded; the client decides whether to
	// reconnect based on recent user activity.
	MaxLifetime time.Duration
	Log         *zap.Logger
}

// runSSEStream writes SSE headers and runs the poll loop: fetch, emit events
// whose hash changed, handle finished state and the grace-period close, send
// keepalives, and honor context cancellation at every wait.
func runSSEStream(c *gin.Context, cfg sseStreamConfig) {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = sseWatchPollInterval
	}
	if cfg.FinishedPollInterval == 0 {
		cfg.FinishedPollInterval = sseFinishedPollInterval
	}
	if cfg.ErrorRetryDelay == 0 {
		cfg.ErrorRetryDelay = sseErrorRetryDelay
	}
	if cfg.MaxLifetime == 0 {
		cfg.MaxLifetime = sseMaxStreamLifetime
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	ctx := c.Request.Context()
	hashes := map[string]string{}
	start := time.Now()
	var finishedAt time.Time

	sendEvent := func(ev sseEvent) {
		if ev.Name == "" {
			fmt.Fprintf(c.Writer, "data: %s\n\n", ev.Data)
		} else {
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", ev.Name, ev.Data)
		}
		c.Writer.Flush()
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if time.Since(start) >= cfg.MaxLifetime {
			sendEvent(sseEvent{Name: "expired", Data: []byte("true")})
			return
		}

		res, fetchErr := cfg.Fetch(ctx)

		// Emit partial results even when the fetch errored, so a failed
		// secondary fetch doesn't drop updates from the ones that succeeded.
		primaryChanged := false
		for i, ev := range res.Events {
			hash := sha256.Sum256(ev.Data)
			hashStr := hex.EncodeToString(hash[:])
			if hashStr == hashes[ev.Name] {
				continue
			}
			hashes[ev.Name] = hashStr
			sendEvent(ev)
			if i == 0 {
				primaryChanged = true
			}
		}

		if fetchErr != nil {
			if cfg.Log != nil {
				cfg.Log.Error("sse fetch failed", zap.String("path", c.FullPath()), zap.Error(fetchErr))
			}
			if !errors.Is(fetchErr, errSSESilentRetry) {
				b, _ := json.Marshal(map[string]string{"error": cfg.ClientErrMsg})
				sendEvent(sseEvent{Name: "fetch-error", Data: b})
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(cfg.ErrorRetryDelay):
			}
			continue
		}

		if cfg.FinishedGracePeriod > 0 && primaryChanged {
			if res.Finished {
				sendEvent(sseEvent{Name: "finished", Data: []byte("true")})
				if finishedAt.IsZero() {
					finishedAt = time.Now()
				}
			} else {
				finishedAt = time.Time{}
			}
		}

		if !finishedAt.IsZero() && time.Since(finishedAt) > cfg.FinishedGracePeriod {
			return
		}

		fmt.Fprintf(c.Writer, ": keepalive\n\n")
		c.Writer.Flush()

		interval := cfg.PollInterval
		if !finishedAt.IsZero() {
			interval = cfg.FinishedPollInterval
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
