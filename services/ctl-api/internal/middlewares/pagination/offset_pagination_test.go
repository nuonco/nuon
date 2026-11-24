package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/pagination"
)

func TestOffsetPaginationHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := map[string]struct {
		method      string
		query       string
		expected    *pagination.PaginationQuery
		expectError bool
	}{
		"defaults": {
			method:   http.MethodGet,
			query:    "",
			expected: &pagination.PaginationQuery{Offset: 0, Limit: 10, Page: 0},
		},
		"explicit offset and limit": {
			method:   http.MethodGet,
			query:    "offset=20&limit=5",
			expected: &pagination.PaginationQuery{Offset: 20, Limit: 5, Page: 0},
		},
		"page derives offset": {
			method:   http.MethodGet,
			query:    "page=3&limit=5",
			expected: &pagination.PaginationQuery{Offset: 10, Limit: 5, Page: 3},
		},
		"zero offset is allowed": {
			method:   http.MethodGet,
			query:    "offset=0",
			expected: &pagination.PaginationQuery{Offset: 0, Limit: 10, Page: 0},
		},
		"zero page is allowed": {
			method:   http.MethodGet,
			query:    "page=0",
			expected: &pagination.PaginationQuery{Offset: 0, Limit: 10, Page: 0},
		},
		"limit at max is allowed": {
			method:   http.MethodGet,
			query:    "limit=100",
			expected: &pagination.PaginationQuery{Offset: 0, Limit: 100, Page: 0},
		},
		"non numeric values fall back to defaults": {
			method:   http.MethodGet,
			query:    "offset=abc&limit=abc&page=abc",
			expected: &pagination.PaginationQuery{Offset: 0, Limit: 10, Page: 0},
		},
		"limit over max is rejected": {
			method:      http.MethodGet,
			query:       "limit=101",
			expectError: true,
		},
		"negative offset is rejected": {
			method:      http.MethodGet,
			query:       "offset=-1",
			expectError: true,
		},
		"negative limit is rejected": {
			method:      http.MethodGet,
			query:       "limit=-1",
			expectError: true,
		},
		"negative page is rejected": {
			method:      http.MethodGet,
			query:       "page=-1",
			expectError: true,
		},
		"negative values are ignored for non GET requests": {
			method:   http.MethodPost,
			query:    "offset=-1&limit=-1&page=-1",
			expected: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				handlerCalled bool
				got           *pagination.PaginationQuery
				errs          []*gin.Error
			)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Next()
				errs = c.Errors
			})
			router.Use(middleware{}.Handler())
			router.Handle(tt.method, "/", func(c *gin.Context) {
				handlerCalled = true
				got = cctx.OffsetPaginationFromContext(c)
				c.Status(http.StatusOK)
			})

			target := "/"
			if tt.query != "" {
				target += "?" + tt.query
			}
			req := httptest.NewRequest(tt.method, target, nil)
			router.ServeHTTP(httptest.NewRecorder(), req)

			if tt.expectError {
				assert.False(t, handlerCalled)
				require.Len(t, errs, 1)
				var userErr stderr.ErrUser
				assert.ErrorAs(t, errs[0].Err, &userErr)
				return
			}

			require.True(t, handlerCalled)
			assert.Empty(t, errs)
			assert.Equal(t, tt.expected, got)
		})
	}
}
