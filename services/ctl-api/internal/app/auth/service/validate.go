package service

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Response headers for validated requests
const (
	HeaderNuonAuthUser    = "X-Nuon-Auth-User"
	HeaderNuonAuthEmail   = "X-Nuon-Auth-Email"
	HeaderNuonAuthSuccess = "X-Nuon-Auth-Success"
	HeaderNuonAuthClaims  = "X-Nuon-Auth-Claims"
)

var (
	errNoJWT   = errors.New("no JWT found in request")
	errNoUser  = errors.New("no user found in JWT")
	errBadHost = errors.New("host not authorized")
)

// Validate handles the /validate endpoint.
// This is called by reverse proxies (nginx, etc.) to validate requests.
// It checks the auth cookie and returns headers with user information.
func (s *service) Validate(c *gin.Context) {
	s.l.Debug("/validate")

	// Try to find the JWT from cookie or Authorization header
	jwt := s.findJWT(c)
	if jwt == "" {
		s.sendValidateError(c, errNoJWT)
		return
	}

	// Parse and validate the JWT
	claims, err := s.parseJWT(jwt)
	if err != nil {
		s.sendValidateError(c, err)
		return
	}

	// Ensure we have user info
	if claims.Email == "" && claims.Username == "" {
		s.sendValidateError(c, errNoUser)
		return
	}

	// TODO: Validate the host against allowed domains
	// host := c.Request.Host
	// if !s.isHostAllowed(host, claims) {
	//     s.sendValidateError(c, fmt.Errorf("%w: %s", errBadHost, host))
	//     return
	// }

	// Set response headers with user information
	c.Header(HeaderNuonAuthSuccess, "true")
	c.Header(HeaderNuonAuthUser, claims.Username)
	c.Header(HeaderNuonAuthEmail, claims.Email)

	// TODO: Add custom claims headers if configured
	// if claims.CustomClaims != nil {
	//     s.setCustomClaimsHeaders(c, claims.CustomClaims)
	// }

	s.l.Debug("validate success",
		zap.String("user", claims.Username),
		zap.String("email", claims.Email))

	c.Status(http.StatusOK)
}

// findJWT looks for the JWT in the cookie or Authorization header.
func (s *service) findJWT(c *gin.Context) string {
	// First, try the cookie
	if cookie, err := s.getCookie(c); err == nil && cookie != "" {
		return cookie
	}

	// Then, try the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Support "Bearer <token>" format
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		return authHeader
	}

	// Finally, try a custom header
	if token := c.GetHeader(NuonAuthCookieName); token != "" {
		return token
	}

	return ""
}

// sendValidateError sends an appropriate error response for validation failures.
func (s *service) sendValidateError(c *gin.Context, err error) {
	s.l.Debug("validate failed", zap.Error(err))

	// TODO: Support public access mode where unauthenticated requests are allowed
	// if s.cfg.PublicAccess {
	//     c.Header(HeaderNuonAuthUser, "")
	//     c.Status(http.StatusOK)
	//     return
	// }

	c.Header(HeaderNuonAuthSuccess, "false")
	c.Status(http.StatusUnauthorized)
}
