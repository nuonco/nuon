package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// TestHandlerBlocksReadOnlyOAuthWrites drives real HTTP requests through this
// middleware and the error-rendering middleware against a real token/account
// row, proving the write-block itself (not just its pure helpers). It stops
// at this middleware's own gin.Context.Abort() and does not mount the org
// permission layer (CanPerform), so it does not exercise require.Read/Write
// or org-scoped authz.
func TestHandlerBlocksReadOnlyOAuthWrites(t *testing.T) {
	tests.SkipIfNotIntegration(t)

	cfg, err := tests.LoadDBConfig()
	require.NoError(t, err)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	acct := &app.Account{
		Email:       fmt.Sprintf("nuo1118-%d@example.com", time.Now().UnixNano()),
		Subject:     fmt.Sprintf("nuo1118-subject-%d", time.Now().UnixNano()),
		AccountType: app.AccountTypeAuth,
	}
	require.NoError(t, db.Create(acct).Error)
	t.Cleanup(func() {
		db.Unscoped().Where("account_id = ?", acct.ID).Delete(&app.Token{})
		db.Unscoped().Delete(acct)
	})

	readOnlyToken := createOAuthToken(t, db, acct.ID, app.RoleTypeOrgReadOnly)
	adminToken := createOAuthToken(t, db, acct.ID, app.RoleTypeOrgAdmin)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(stderr.New(zap.NewNop(), &internal.Config{}).Handler())
	m := &middleware{
		l:          zap.NewNop(),
		db:         db,
		acctClient: account.New(account.Params{DB: db}),
	}
	engine.Use(m.Handler())
	engine.Any("/*any", func(c *gin.Context) { c.Status(http.StatusOK) })

	cases := []struct {
		name   string
		token  string
		method string
		want   int
	}{
		{"read-only GET", readOnlyToken, http.MethodGet, http.StatusOK},
		{"read-only POST", readOnlyToken, http.MethodPost, http.StatusForbidden},
		{"read-only PUT", readOnlyToken, http.MethodPut, http.StatusForbidden},
		{"read-only PATCH", readOnlyToken, http.MethodPatch, http.StatusForbidden},
		{"read-only DELETE", readOnlyToken, http.MethodDelete, http.StatusForbidden},
		{"admin POST", adminToken, http.MethodPost, http.StatusOK},
		{"admin PUT", adminToken, http.MethodPut, http.StatusOK},
		{"admin PATCH", adminToken, http.MethodPatch, http.StatusOK},
		{"admin DELETE", adminToken, http.MethodDelete, http.StatusOK},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/v1/apps", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)
			require.Equal(t, tt.want, rec.Code, rec.Body.String())
		})
	}
}

func createOAuthToken(t *testing.T, db *gorm.DB, accountID string, role app.RoleType) string {
	t.Helper()

	tok := &app.Token{
		AccountID:   accountID,
		CreatedByID: accountID,
		Name:        "nuo1118-" + string(role),
		Role:        string(role),
		Token:       fmt.Sprintf("nuo1118-%s-%d", role, time.Now().UnixNano()),
		TokenType:   app.TokenTypeOAuth,
		ExpiresAt:   time.Now().Add(time.Hour),
		IssuedAt:    time.Now(),
		Issuer:      "nuo1118-test",
	}
	require.NoError(t, db.Create(tok).Error)
	return tok.Token
}
