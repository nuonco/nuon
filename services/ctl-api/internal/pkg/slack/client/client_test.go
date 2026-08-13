package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersLookupByEmail(t *testing.T) {
	t.Run("resolves user id", func(t *testing.T) {
		var gotEmail, gotAuth string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/users.lookupByEmail", r.URL.Path)
			gotEmail = r.URL.Query().Get("email")
			gotAuth = r.Header.Get("Authorization")
			w.Write([]byte(`{"ok":true,"user":{"id":"U123"}}`))
		}))
		defer srv.Close()

		c := New(WithBaseURL(srv.URL))
		resp, err := c.UsersLookupByEmail(context.Background(), "xoxb-token", "a@b.co")
		require.NoError(t, err)
		assert.Equal(t, "U123", resp.User.ID)
		assert.Equal(t, "a@b.co", gotEmail)
		assert.Equal(t, "Bearer xoxb-token", gotAuth)
	})

	t.Run("users_not_found returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"ok":false,"error":"users_not_found"}`))
		}))
		defer srv.Close()

		c := New(WithBaseURL(srv.URL))
		_, err := c.UsersLookupByEmail(context.Background(), "xoxb-token", "a@b.co")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "users_not_found")
	})

	t.Run("non-2xx returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer srv.Close()

		c := New(WithBaseURL(srv.URL))
		_, err := c.UsersLookupByEmail(context.Background(), "xoxb-token", "a@b.co")
		require.Error(t, err)
	})
}
