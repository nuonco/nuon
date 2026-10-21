package oidc_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
)

func TestHTTPClientAddressGuard(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	tests := map[string]struct {
		allow   bool
		wantErr string
	}{
		"guard on refuses loopback": {allow: false, wantErr: "refusing to connect to non-public address"},
		"guard off allows loopback": {allow: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := oidc.NewHTTPClient(oidc.HTTPClientOptions{AllowPrivateAddresses: tt.allow})
			resp, err := c.Get(srv.URL)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, "ok", string(body))
		})
	}
}

func TestHTTPClientRedirectNotFollowed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/target", http.StatusFound)
	})
	mux.HandleFunc("/target", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := oidc.NewHTTPClient(oidc.HTTPClientOptions{AllowPrivateAddresses: true})
	resp, err := c.Get(srv.URL + "/start")
	if resp != nil {
		resp.Body.Close()
	}
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redirects are not followed")
}

func TestHTTPClientBodyLimit(t *testing.T) {
	tests := map[string]struct {
		body    string
		max     int64
		wantErr bool
	}{
		"under cap":         {body: "0123456789", max: 16},
		"exactly at cap":    {body: strings.Repeat("a", 16), max: 16},
		"over cap":          {body: strings.Repeat("a", 17), max: 16, wantErr: true},
		"far over cap":      {body: strings.Repeat("a", 4096), max: 16, wantErr: true},
		"default cap holds": {body: strings.Repeat("a", 4096), max: 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := oidc.NewHTTPClient(oidc.HTTPClientOptions{MaxBodyBytes: tt.max, AllowPrivateAddresses: true})
			resp, err := c.Get(srv.URL)
			require.NoError(t, err)
			defer resp.Body.Close()

			got, err := io.ReadAll(resp.Body)
			if tt.wantErr {
				require.ErrorIs(t, err, oidc.ErrResponseTooLarge)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.body, string(got))
		})
	}
}

func TestHTTPClientTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		_, _ = w.Write([]byte("late"))
	}))
	defer srv.Close()

	c := oidc.NewHTTPClient(oidc.HTTPClientOptions{Timeout: 100 * time.Millisecond, AllowPrivateAddresses: true})
	resp, err := c.Get(srv.URL)
	if resp != nil {
		resp.Body.Close()
	}
	require.Error(t, err)
}
