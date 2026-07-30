package service

import (
	"testing"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func TestJWTExpiryRequirement(t *testing.T) {
	tests := []struct {
		name   string
		expiry int64
		valid  bool
	}{
		{name: "missing exp", valid: false},
		{name: "exp present", expiry: 1, valid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &validator.ValidatedClaims{RegisteredClaims: validator.RegisteredClaims{Expiry: tt.expiry}}
			err := requireJWTExpiry(claims)
			if got := err == nil; got != tt.valid {
				t.Fatalf("requireJWTExpiry() error = %v; validity = %t, want %t", err, got, tt.valid)
			}
		})
	}
}
