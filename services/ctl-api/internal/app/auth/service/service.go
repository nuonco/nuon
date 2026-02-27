package service

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	awsidentity "github.com/nuonco/nuon/pkg/identity/aws"
	azureidentity "github.com/nuonco/nuon/pkg/identity/azure"
	gcpidentity "github.com/nuonco/nuon/pkg/identity/gcp"
	"github.com/nuonco/nuon/pkg/identity/jwks"
	localidentity "github.com/nuonco/nuon/pkg/identity/local"
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

	// Identity validators for runner authentication (using pkg/identity)
	jwksCache      *jwks.Cache
	awsValidator   *IdentityValidator
	gcpValidator   *IdentityValidator
	azureValidator *IdentityValidator
	localValidator *IdentityValidator
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// Note: /v1/auth/me is registered in accountsservice so it's available in PublicServicesModule
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	auth := api.Group("/v1/runner-auth")
	{
		auth.POST("/aws", s.RunnerAuthAWS)   // Existing presigned auth (backward compatibility)
		auth.POST("/oidc", s.RunnerAuthOIDC) // New OIDC auth endpoint
	}
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
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

	// Device code flow for CLI authentication
	api.GET("/device/code", s.DeviceCodePage)
	api.POST("/device/code/approve", s.DeviceCodeApprove)
	api.GET("/device/token", s.DeviceCodeToken)

	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) (*service, error) {
	// Initialize JWKS cache for OIDC token validation (5-minute TTL)
	jwksCache := jwks.NewCache(5 * time.Minute)

	s := &service{
		cfg:        params.Cfg,
		l:          params.L,
		v:          params.V,
		db:         params.DB,
		mw:         params.MW,
		acctClient: params.AcctClient,
		jwksCache:  jwksCache,
	}

	// Initialize identity validators for runner authentication
	audience := params.Cfg.RunnerAPIURL
	if audience == "" {
		audience = "https://api.nuon.co" // Default audience
	}

	// Create pkg/identity validators
	awsValidator := awsidentity.NewValidator(params.L, jwksCache)
	gcpValidator := gcpidentity.NewValidator(params.L, jwksCache, audience)
	azureValidator := azureidentity.NewValidator(params.L, jwksCache, audience)
	localValidator := localidentity.NewValidator(&localidentity.ValidatorConfig{
		Logger:   params.L,
		AllowAny: false, // Don't allow any token in production
	})

	// Wrap with install-specific validation
	s.awsValidator = NewIdentityValidator(params.L, params.DB, awsValidator)
	s.gcpValidator = NewIdentityValidator(params.L, params.DB, gcpValidator)
	s.azureValidator = NewIdentityValidator(params.L, params.DB, azureValidator)
	s.localValidator = NewIdentityValidator(params.L, params.DB, localValidator)

	// Validate required configs
	if s.cfg.RootDomain == "" {
		return nil, fmt.Errorf("nuon_root_domain is required")
	}

	// Validate required secrets
	if s.cfg.NuonAuthSessionKey == "" {
		return nil, fmt.Errorf("nuon_auth_session_key is required")
	}

	// NOTE(fd): an empty env var `""` produces [""] via StringToSliceHookFunc so we must
	// filter out the empty strings
	for _, domain := range s.cfg.NuonAuthAllowedDomains {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			s.allowedDomains = append(s.allowedDomains, strings.ToLower(domain))
		}
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

	s.l.Info("allowed domains configured",
		zap.Strings("domains", s.allowedDomains))

	s.l.Info("auth service initialized",
		zap.String("provider_type", string(defaultIP.ProviderType)),
		zap.String("provider_id", defaultIP.ID))

	return s, nil
}
