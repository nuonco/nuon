package middleware

import (
	"context"
	"net/http"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserGetter struct {
	mock.Mock
}

func (m *mockUserGetter) GetUserByExternalID(ctx context.Context, externalID string) (*models.User, error) {
	m.Called(externalID)
	return &models.User{Email: "hello@nuon.co", ExternalID: externalID}, nil
}

func TestUserAuthMiddleware(t *testing.T) {
	tests := []struct {
		description string
		status      int
		message     string
		shouldCall  bool
		cb          gin.HandlerFunc
		beforeHook  gin.HandlerFunc
	}{

		{
			description: "should error when an invalid token is passed",
			status:      http.StatusBadRequest,
			cb: func(c *gin.Context) {
				assert.True(t, false, "should not be called")
			},
		},
		{
			description: "should error with an invalid token is passed",
			status:      http.StatusBadRequest,
			cb: func(c *gin.Context) {
				assert.True(t, false, "should not be called")
			},
		},
		{
			description: "should error when a bad token is set",
			status:      http.StatusBadRequest,
			beforeHook: func(c *gin.Context) {
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), jwtmiddleware.ContextKey{}, "nil"))
			},
			cb: func(c *gin.Context) {
				assert.True(t, false, "should not be called")
			},
		},
		{
			description: "should error when the email is empty in the claim",
			status:      http.StatusBadRequest,
			beforeHook: func(c *gin.Context) {
				customClaim := &CustomClaim{
					ExternalID: "",
				}

				token := &validator.ValidatedClaims{
					CustomClaims: customClaim,
				}
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), jwtmiddleware.ContextKey{}, token))
				c.Next()
			},
			cb: func(c *gin.Context) {
				assert.True(t, false, "should not be called")
			},
		},
		{
			description: "should return a valid email is in the token",
			status:      http.StatusOK,
			shouldCall:  true,
			beforeHook: func(c *gin.Context) {
				customClaim := &CustomClaim{
					ExternalID: "email@nuon.co",
				}

				token := &validator.ValidatedClaims{
					CustomClaims: customClaim,
				}
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), jwtmiddleware.ContextKey{}, token))
				c.Next()
			},
			cb: func(c *gin.Context) {
				assert.True(t, true, "should be called")
				c.JSON(http.StatusOK, "ok")
			},
		},
	}

	// execute tests
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			userSvc := new(mockUserGetter)
			userSvc.On("GetUserByExternalID", "email@nuon.co")

			router := gin.New()
			router.HandleMethodNotAllowed = true
			if test.beforeHook != nil {
				router.Use(test.beforeHook)
			}

			middleware := NewUserAuthMiddleware(userSvc)
			router.Use(middleware)

			if test.cb != nil {
				router.Use(test.cb)
			}

			w := PerformRequest(router, "GET", "/")
			assert.Equal(t, test.status, w.Code)

			if test.shouldCall {
				userSvc.AssertCalled(t, "GetUserByExternalID", "email@nuon.co")
			}
		})
	}
}
