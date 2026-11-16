package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestIsReadOnlyOAuthToken(t *testing.T) {
	tests := []struct {
		name  string
		token app.Token
		want  bool
	}{
		{name: "oauth read only", token: app.Token{TokenType: app.TokenTypeOAuth, Role: string(app.RoleTypeOrgReadOnly)}, want: true},
		{name: "oauth without role", token: app.Token{TokenType: app.TokenTypeOAuth}, want: true},
		{name: "oauth admin", token: app.Token{TokenType: app.TokenTypeOAuth, Role: string(app.RoleTypeOrgAdmin)}, want: false},
		{name: "static read only service account", token: app.Token{TokenType: app.TokenTypeStatic, Role: string(app.RoleTypeOrgReadOnly)}, want: false},
		{name: "nuon session", token: app.Token{TokenType: app.TokenTypeNuon}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isReadOnlyOAuthToken(&tt.token))
		})
	}
}

func TestIsReadMethod(t *testing.T) {
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		assert.True(t, isReadMethod(m), m)
	}
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		assert.False(t, isReadMethod(m), m)
	}
}
