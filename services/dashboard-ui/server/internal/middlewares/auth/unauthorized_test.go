package auth

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
)

func TestIsUnauthorized(t *testing.T) {
	tests := map[string]struct {
		err  error
		want bool
	}{
		"typed 401":         {err: &operations.ValidateTokenUnauthorized{}, want: true},
		"wrapped typed 401": {err: fmt.Errorf("validate: %w", &operations.ValidateTokenUnauthorized{}), want: true},
		"api error 401":     {err: runtime.NewAPIError("unauthorized", nil, http.StatusUnauthorized), want: true},
		"api error 403":     {err: runtime.NewAPIError("forbidden", nil, http.StatusForbidden), want: false},
		"api error 500":     {err: runtime.NewAPIError("boom", nil, http.StatusInternalServerError), want: false},
		"api error 502":     {err: runtime.NewAPIError("bad gateway", nil, http.StatusBadGateway), want: false},
		"dns failure":       {err: errors.New(`dial tcp: lookup api.example.com: no such host`), want: false},
		"nil":               {err: nil, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsUnauthorized(tc.err))
		})
	}
}
