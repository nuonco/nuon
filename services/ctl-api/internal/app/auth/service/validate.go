package service

import (
	"errors"
	"net/http"

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
	errNoToken = errors.New("no token found in request")
	errNoUser  = errors.New("no user found in token")
	errBadHost = errors.New("host not authorized")
)

// Validate handles the /validate endpoint.
// This is called by reverse proxies (nginx, etc.) to validate requests.
// It checks the auth cookie and returns headers with user information.
func (s *service) Validate(c *gin.Context) {
	s.l.Debug("/validate")

	// Try to find the token from cookie or Authorization header
	token := s.findToken(c)
	if token == "" {
		s.sendValidateError(c, errNoToken)
		return
	}

	// Validate the token and get account info
	tokenInfo, err := s.validateToken(token)
	if err != nil {
		s.sendValidateError(c, err)
		return
	}

	// Ensure we have user info
	if tokenInfo.Email == "" && tokenInfo.Username == "" {
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
	c.Header(HeaderNuonAuthUser, tokenInfo.Username)
	c.Header(HeaderNuonAuthEmail, tokenInfo.Email)

	s.l.Debug("validate success",
		zap.String("user", tokenInfo.Username),
		zap.String("email", tokenInfo.Email))

	c.Status(http.StatusOK)
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
