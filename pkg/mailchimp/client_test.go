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
		method   string
		path     string
		username string
		password string
		body     upsertListMemberRequest
	}

	var got recorded
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		var ok bool
		got.username, got.password, ok = r.BasicAuth()
		require.True(t, ok)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got.body))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := New(validator.New(),
		WithAPIKey("test-api-key"),
		WithServerPrefix("us21"),
		WithAudienceID("aud-123"),
		WithBaseURL(srv.URL),
	)
	require.NoError(t, err)

	require.NoError(t, c.UpsertListMember(context.Background(), "Test@Example.com"))

	assert.Equal(t, http.MethodPut, got.method)
	assert.Equal(t, "/lists/aud-123/members/55502f40dc8b7c769880b10874abc9d0", got.path)
	assert.Equal(t, "test-api-key", got.password)
	assert.Equal(t, "Test@Example.com", got.body.EmailAddress)
	assert.Equal(t, "subscribed", got.body.StatusIfNew)
}

func TestUpsertListMemberErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"title":"Invalid Resource"}`))
	}))
	defer srv.Close()

	c, err := New(validator.New(),
		WithAPIKey("test-api-key"),
		WithServerPrefix("us21"),
		WithAudienceID("aud-123"),
		WithBaseURL(srv.URL),
	)
	require.NoError(t, err)

	err = c.UpsertListMember(context.Background(), "test@example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "Invalid Resource")
}

func TestNewRequiresConfig(t *testing.T) {
	_, err := New(validator.New(), WithAPIKey("test-api-key"))
	require.Error(t, err)
}

func TestDisabledClient(t *testing.T) {
	c := NewDisabled()
	require.NoError(t, c.UpsertListMember(context.Background(), "test@example.com"))
}

func TestBaseURLDerivedFromServerPrefix(t *testing.T) {
	c, err := New(validator.New(),
		WithAPIKey("test-api-key"),
		WithServerPrefix("us21"),
		WithAudienceID("aud-123"),
	)
	require.NoError(t, err)
	assert.Equal(t, "https://us21.api.mailchimp.com/3.0", c.baseURL)
}
