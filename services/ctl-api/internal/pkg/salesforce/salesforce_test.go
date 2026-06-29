package salesforce

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func TestEnabled(t *testing.T) {
	assert.False(t, New(&internal.Config{}).Enabled())
	assert.True(t, New(&internal.Config{SFTrialEndpoint: "https://example.com/hook"}).Enabled())
}

func TestSendTrialSignup(t *testing.T) {
	var received TrialSignup
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := New(&internal.Config{SFTrialEndpoint: srv.URL})
	signup := TrialSignup{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.com",
		Notes:     "Created via CLI. Org: acme",
		Subject:   TrialSignupSubject,
	}

	require.NoError(t, client.SendTrialSignup(context.Background(), signup))
	assert.Equal(t, signup, received)
}

func TestSendTrialSignupErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := New(&internal.Config{SFTrialEndpoint: srv.URL})
	err := client.SendTrialSignup(context.Background(), TrialSignup{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestSendTrialSignupDisabled(t *testing.T) {
	client := New(&internal.Config{})
	require.NoError(t, client.SendTrialSignup(context.Background(), TrialSignup{}))
}
