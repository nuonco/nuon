package middleware

import (
	"context"
	"fmt"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/api/internal/models"
)

// UserContext is the context key representing a user.
type UserContext struct{}

// UserIDContext is the context key representing a user id.
type UserIDContext struct{}

var (
	errNoClaim      error = fmt.Errorf("no claim found")
	errInvalidClaim error = fmt.Errorf("invalid claim err")
	errEmptyClaim   error = fmt.Errorf("claims were empty")
)

// parseJWT: parse the JWT token, returning a custom claim
func parseJWT(ctx context.Context) (*CustomClaim, error) {
	claim := ctx.Value(jwtmiddleware.ContextKey{})
	if claim == nil {
		return nil, errNoClaim
	}

	validatedClaims, ok := claim.(*validator.ValidatedClaims)
	if !ok {
		return nil, errInvalidClaim
	}

	customClaims, ok := validatedClaims.CustomClaims.(*CustomClaim)
	if !ok {
		return nil, errInvalidClaim
	}

	return customClaims, nil
}

type userAuthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type userGetterService interface {
	GetUserByExternalID(context.Context, string) (*models.User, error)
}

// NewUserAuthMiddleware returns a middleware that enforces the user is authenticated before proceeding
func NewUserAuthMiddleware(userSvc userGetterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := parseJWT(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadRequest, userAuthResponse{
				Status:  "bad-request",
				Message: err.Error(),
			})
			c.Abort()
			return
		}

		if claims.ExternalID == "" {
			fmt.Println("no external ID", claims)
			c.JSON(http.StatusBadRequest, userAuthResponse{
				Status:  "bad-request",
				Message: errEmptyClaim.Error(),
			})
			c.Abort()
			return
		}

		user, err := userSvc.GetUserByExternalID(c, claims.ExternalID)
		// TODO(jm): support properly parsing this for a not found, or database error. For now, assume not
		// found, and rely on health checks to surface other db problems.
		if err != nil {
			c.JSON(http.StatusNotFound, userAuthResponse{
				Status:  "not-found",
				Message: err.Error(),
			})
			c.Abort()
			return
		}

		// NOTE: set the object on both the gin context, as well as the underlying context as both can be used
		// downstream differently (ie: gqlgen uses the regular context)
		c.Set("user", user)
		c.Set("user_id", user.ID.String())

		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, UserContext{}, user)
		reqCtx = context.WithValue(reqCtx, UserIDContext{}, user.ID.String())
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}
