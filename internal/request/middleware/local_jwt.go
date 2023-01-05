package middleware

import (
	"context"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// localJWTMiddleware is a middleware that allows us to inject what looks like a valid JWT claim locally, into the
// request stack. This handler will create a token and add it to the context of the request, just like we would see from
// a normal JWT.
type localJWTMiddleware struct {
	email string
}

func (l *localJWTMiddleware) InjectJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customClaim := &CustomClaim{
			ExternalID: l.email,
		}

		token := &validator.ValidatedClaims{
			CustomClaims: customClaim,
		}

		r = r.Clone(context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, token))
		next.ServeHTTP(w, r)
	})
}

// NewLocalJWTMiddleware is a middleware that creates a fake jwt token and puts it on the context for this request with
// the provided email
func NewLocalJWTMiddleware(email string) *localJWTMiddleware {
	return &localJWTMiddleware{
		email: email,
	}
}
