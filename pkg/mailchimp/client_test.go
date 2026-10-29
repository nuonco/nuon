package mailchimp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertListMember(t *testing.T) {
	type recorded struct {
		method string
		path   string
		auth   string
		body   map[string]any
	}

	var got recorded
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.auth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got.body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"email_address":"Test@Example.com","status":"subscribed"}`))
	}))
	defer srv.Close()

	c, err := New(validator.New(),
		WithAPIKey("test-api-key"),
		WithAudienceID("aud-123"),
		WithBaseURL(srv.URL),
	)
	require.NoError(t, err)

	require.NoError(t, c.UpsertListMember(context.Background(), "Test@Example.com"))

	assert.Equal(t, http.MethodPut, got.method)
	assert.Equal(t, "/3.0/lists/aud-123/members/55502f40dc8b7c769880b10874abc9d0", got.path)
	assert.Equal(t, "Bearer test-api-key", got.auth)
	assert.Equal(t, "Test@Example.com", got.body["email_address"])
	assert.Equal(t, "subscribed", got.body["status_if_new"])
	assert.NotContains(t, got.body, "status", "an upsert must never overwrite an existing member's status")
}

func TestUpsertListMemberErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"title":"Invalid Resource"}`))
	}))
	defer srv.Close()

	c, err := New(validator.New(),
		WithAPIKey("test-api-key"),
		WithAudienceID("aud-123"),
		WithBaseURL(srv.URL),
	)
	require.NoError(t, err)

	err = c.UpsertListMember(context.Background(), "test@example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestNewRequiresConfig(t *testing.T) {
	_, err := New(validator.New(), WithAPIKey("test-api-key"))
	require.Error(t, err)
}

func TestDisabledClient(t *testing.T) {
	c := NewDisabled()
	assert.False(t, c.Enabled())
	require.NoError(t, c.UpsertListMember(context.Background(), "test@example.com"))
}

func TestBaseURL(t *testing.T) {
	c, err := New(validator.New(),
		WithAPIKey("test-api-key"),
		WithServerPrefix("us21"),
		WithAudienceID("aud-123"),
	)
	require.NoError(t, err)
	assert.True(t, c.Enabled())
	assert.Equal(t, "https://us21.api.mailchimp.com", c.baseURL)

	c, err = New(validator.New(),
		WithAPIKey("test-api-key"),
		WithAudienceID("aud-123"),
	)
	require.NoError(t, err)
	assert.Empty(t, c.baseURL, "no server prefix leaves the SDK's default host")
}
