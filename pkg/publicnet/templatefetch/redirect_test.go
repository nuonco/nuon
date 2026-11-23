package templatefetch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/publicnet"
)

func TestTemplateFetchClient_RunsTheRedirectCheck(t *testing.T) {
	var targetHits atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetHits.Add(1)
		_, _ = io.WriteString(w, "template")
	}))
	defer target.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/stack.yaml", http.StatusFound)
	}))
	defer srv.Close()

	get := func(t *testing.T, check func(*url.URL) error) (*http.Response, error) {
		ctx := publicnet.WithRedirectCheck(context.Background(), check)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
		require.NoError(t, err)
		return NewLocalClient().Do(req)
	}

	t.Run("a refused target is not fetched", func(t *testing.T) {
		targetHits.Store(0)
		var checked []string
		_, err := get(t, func(u *url.URL) error {
			checked = append(checked, u.String())
			return errors.New("not on the list")
		})
		require.ErrorContains(t, err, "redirect refused: not on the list")
		require.True(t, IsRefused(err))
		require.Equal(t, []string{target.URL + "/stack.yaml"}, checked)
		require.Zero(t, targetHits.Load())
	})

	t.Run("an accepted target is fetched", func(t *testing.T) {
		resp, err := get(t, func(*url.URL) error { return nil })
		require.NoError(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "template", string(body))
	})

	t.Run("no check follows the redirect", func(t *testing.T) {
		resp, err := get(t, nil)
		require.NoError(t, err)
		resp.Body.Close()
	})
}
