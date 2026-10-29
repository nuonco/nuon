package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/mailchimp"
	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type MarketingConsentHandler struct {
	cfg       *internal.Config
	l         *zap.Logger
	mailchimp mailchimp.Client
}

func NewMarketingConsentHandler(cfg *internal.Config, l *zap.Logger) (*MarketingConsentHandler, error) {
	mc := mailchimp.NewDisabled()
	if cfg.MailchimpAPIKey != "" && cfg.MailchimpAudienceID != "" {
		c, err := mailchimp.New(validator.New(),
			mailchimp.WithAPIKey(cfg.MailchimpAPIKey),
			mailchimp.WithServerPrefix(cfg.MailchimpServerPrefix),
			mailchimp.WithAudienceID(cfg.MailchimpAudienceID),
		)
		if err != nil {
			return nil, err
		}
		mc = c
	}
	return &MarketingConsentHandler{cfg: cfg, l: l, mailchimp: mc}, nil
}

func (h *MarketingConsentHandler) RegisterRoutes(e *gin.Engine) error {
	e.POST("/api/marketing-consent", h.Subscribe)
	return nil
}

func (h *MarketingConsentHandler) Subscribe(c *gin.Context) {
	if !h.mailchimp.Enabled() {
		c.JSON(http.StatusOK, gin.H{"synced": false})
		return
	}

	token, err := c.Cookie(authCookie)
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not signed in"})
		return
	}

	client, err := nuon.New(nuon.WithURL(h.cfg.ServerAPIURL()), nuon.WithAuthToken(token))
	if err != nil {
		h.l.Error("failed to create nuon client", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create client"})
		return
	}

	me, err := client.GetCurrentUser(c.Request.Context())
	if err != nil || me.Email == "" {
		h.l.Warn("failed to get current user", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to get current user"})
		return
	}

	if err := h.mailchimp.UpsertListMember(c.Request.Context(), me.Email); err != nil {
		h.l.Warn("failed to upsert mailchimp audience member", zap.String("account_id", me.ID), zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to subscribe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"synced": true})
}
