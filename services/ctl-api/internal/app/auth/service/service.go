package service

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

// Cookie and session names
const (
	NuonAuthCookieName  string = "X-Nuon-Auth"
	NuonAuthSessionName string = "nuon-auth-session"

	failCountLimit int = 6
)

//go:embed templates
var tmplFS embed.FS

type Params struct {
	fx.In

	V          *validator.Validate
	Cfg        *internal.Config
	DB         *gorm.DB `name:"psql"`
	MW         metrics.Writer
	L          *zap.Logger
	AcctClient *account.Client
}

type service struct {
	v          *validator.Validate
	l          *zap.Logger
	db         *gorm.DB
	mw         metrics.Writer
	cfg        *internal.Config
	acctClient *account.Client

	domain         string   // domain the service is served at
	allowedDomains []string // email domains that are allowed to use this service for auth
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	auth := api.Group("/v1/auth")
	{
		auth.GET("/me", s.GetAuthMe)
	}
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	auth := api.Group("/v1/auth")
	{
		auth.POST("/identity-providers", s.AdminCreateIdentityProvider)
		auth.PATCH("/identity-providers/:identity_provider_id", s.AdminPatchIdentityProvider)
	}
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	// Load HTML templates
	sub, err := fs.Sub(tmplFS, "templates")
	if err != nil {
		return err
	}
	api.LoadHTMLFS(http.FS(sub), "*.tmpl")

	// Register routes
	// Session management is handled via signed cookies in session.go
	api.GET("/login", s.Login)
	api.GET("/auth", s.Auth)
	api.GET("/auth/:state", s.AuthState)
	api.GET("/logout", s.Logout)
	api.GET("/success", s.Success)
	api.GET("/validate", s.Validate)
	api.GET("/", s.Index)

	return nil
}

func New(params Params) (*service, error) {
	s := &service{
		cfg:        params.Cfg,
		l:          params.L,
		v:          params.V,
		db:         params.DB,
		mw:         params.MW,
		acctClient: params.AcctClient,
	}

	// Validate required configs
	if s.cfg.RootDomain == "" {
		return nil, fmt.Errorf("nuon_root_domain is required")
	}

	// Validate required secrets
	if s.cfg.NuonAuthSessionKey == "" {
		return nil, fmt.Errorf("nuon_auth_session_key is required")
	}

	// Validate allowed domains (may not be empty)
	if len(s.cfg.NuonAuthAllowedDomains) == 0 {
		return nil, fmt.Errorf("nuon_auth_allowed_domains is required")
	}

	// configure domain name for the auth service.
	if s.cfg.RootDomain != "localhost" {
		// TODO: consider returning an error if the NuonRootDomain is localhost but the env is not dev
		s.domain = fmt.Sprintf("auth.%s", s.cfg.RootDomain)
	} else {
		s.domain = s.cfg.RootDomain
	}

	// Load and validate the default identity provider from env vars at startup.
	// This ensures the service won't start without valid provider configuration.
	// The config is validated inside getDefaultIdentityProvider() via cfg.Validate().
	// Providers are created dynamically at runtime via getProviderByType() or
	// createProviderFromIdentityProvider() when handling requests.
	defaultIP, err := s.getDefaultIdentityProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to load default identity provider: %w", err)
	}

	// Normalize allowed domains to lowercase
	for _, domain := range s.cfg.NuonAuthAllowedDomains {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			s.allowedDomains = append(s.allowedDomains, strings.ToLower(domain))
		}
	}

	s.l.Info("allowed domains configured",
		zap.Strings("domains", s.allowedDomains))

	s.l.Info("auth service initialized",
		zap.String("provider_type", string(defaultIP.ProviderType)),
		zap.String("provider_id", defaultIP.ID))

	return s, nil
}
