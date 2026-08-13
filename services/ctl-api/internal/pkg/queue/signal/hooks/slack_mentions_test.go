package hooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/hooks/slackrender"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

func fakeLookupServer(t *testing.T, users map[string]string, requests *atomic.Int64) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		id, ok := users[r.URL.Query().Get("email")]
		if !ok {
			w.Write([]byte(`{"ok":false,"error":"users_not_found"}`))
			return
		}
		w.Write([]byte(`{"ok":true,"user":{"id":"` + id + `"}}`))
	}))
}

func TestSlackUserResolver(t *testing.T) {
	t.Run("resolves and memoises hits", func(t *testing.T) {
		var requests atomic.Int64
		srv := fakeLookupServer(t, map[string]string{"a@b.co": "U123"}, &requests)
		defer srv.Close()

		r := newSlackUserResolver(slackclient.New(slackclient.WithBaseURL(srv.URL)), zap.NewNop())
		assert.Equal(t, "U123", r.resolve(context.Background(), "tok", "T1", "a@b.co"))
		assert.Equal(t, "U123", r.resolve(context.Background(), "tok", "T1", "a@b.co"))
		assert.Equal(t, int64(1), requests.Load())
	})

	t.Run("memoises misses", func(t *testing.T) {
		var requests atomic.Int64
		srv := fakeLookupServer(t, nil, &requests)
		defer srv.Close()

		r := newSlackUserResolver(slackclient.New(slackclient.WithBaseURL(srv.URL)), zap.NewNop())
		assert.Equal(t, "", r.resolve(context.Background(), "tok", "T1", "a@b.co"))
		assert.Equal(t, "", r.resolve(context.Background(), "tok", "T1", "a@b.co"))
		assert.Equal(t, int64(1), requests.Load())
	})

	t.Run("skips non-email values without HTTP", func(t *testing.T) {
		var requests atomic.Int64
		srv := fakeLookupServer(t, nil, &requests)
		defer srv.Close()

		r := newSlackUserResolver(slackclient.New(slackclient.WithBaseURL(srv.URL)), zap.NewNop())
		assert.Equal(t, "", r.resolve(context.Background(), "tok", "T1", "acct123"))
		assert.Equal(t, "", r.resolve(context.Background(), "tok", "T1", ""))
		assert.Equal(t, int64(0), requests.Load())
	})

	t.Run("nil client is a no-op", func(t *testing.T) {
		r := newSlackUserResolver(nil, nil)
		assert.Equal(t, "", r.resolve(context.Background(), "tok", "T1", "a@b.co"))
	})
}

func TestWithResolvedMentions(t *testing.T) {
	var requests atomic.Int64
	srv := fakeLookupServer(t, map[string]string{
		"creator@b.co":   "U1",
		"responder@b.co": "U2",
	}, &requests)
	defer srv.Close()

	hook := &SlackSignalLifecycleHook{l: zap.NewNop()}
	resolver := newSlackUserResolver(slackclient.New(slackclient.WithBaseURL(srv.URL)), zap.NewNop())
	install := &app.SlackInstallation{TeamID: "T1", BotAccessToken: "tok"}
	rendered := renderEvent{event: slackrender.Event{
		Workflow: slackrender.WorkflowRef{CreatedByEmail: "creator@b.co"},
		Approval: &slackrender.ApprovalRef{RespondedBy: "responder@b.co"},
	}}

	out := hook.withResolvedMentions(context.Background(), resolver, install, rendered)
	assert.Equal(t, map[string]string{
		"creator@b.co":   "U1",
		"responder@b.co": "U2",
	}, out.event.SlackUserIDByEmail)
	assert.Nil(t, rendered.event.SlackUserIDByEmail)
}
