package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type fakeMailchimp struct {
	enabled bool
	err     error
	emails  []string
}

func (f *fakeMailchimp) Enabled() bool { return f.enabled }

func (f *fakeMailchimp) UpsertListMember(_ context.Context, email string) error {
	f.emails = append(f.emails, email)
	return f.err
}

func TestMarketingConsentSubscribe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctlAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/general/current-user" || r.Header.Get("Authorization") != "Bearer good-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"acct-1","email":"user@example.com"}`))
	}))
	defer ctlAPI.Close()

	tests := map[string]struct {
		mailchimp  *fakeMailchimp
		token      string
		wantStatus int
		wantBody   string
		wantEmails []string
	}{
		"mailchimp not configured": {
			mailchimp:  &fakeMailchimp{},
			token:      "good-token",
			wantStatus: http.StatusOK,
			wantBody:   `{"synced":false}`,
		},
		"subscribes the signed-in user": {
			mailchimp:  &fakeMailchimp{enabled: true},
			token:      "good-token",
			wantStatus: http.StatusOK,
			wantBody:   `{"synced":true}`,
			wantEmails: []string{"user@example.com"},
		},
		"no auth cookie": {
			mailchimp:  &fakeMailchimp{enabled: true},
			wantStatus: http.StatusUnauthorized,
		},
		"mailchimp error": {
			mailchimp:  &fakeMailchimp{enabled: true, err: errors.New("mailchimp is down")},
			token:      "good-token",
			wantStatus: http.StatusBadGateway,
			wantEmails: []string{"user@example.com"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			h := &MarketingConsentHandler{
				cfg:       &internal.Config{APIUrl: ctlAPI.URL},
				l:         zap.NewNop(),
				mailchimp: tt.mailchimp,
			}
			r := gin.New()
			assert.NoError(t, h.RegisterRoutes(r))

			req := httptest.NewRequest(http.MethodPost, "/api/marketing-consent", nil)
			if tt.token != "" {
				req.AddCookie(&http.Cookie{Name: authCookie, Value: tt.token})
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, rr.Body.String())
			}
			assert.Equal(t, tt.wantEmails, tt.mailchimp.emails)
		})
	}
}
