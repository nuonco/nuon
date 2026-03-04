package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type ConnectHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewConnectHandler(cfg *internal.Config, l *zap.Logger) *ConnectHandler {
	return &ConnectHandler{cfg: cfg, l: l}
}

func (h *ConnectHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/connect", h.Handle)
	return nil
}

func (h *ConnectHandler) Handle(c *gin.Context) {
	installationID := c.Query("installation_id")
	orgID := c.Query("state")

	fallback := fmt.Sprintf("/%s/apps", orgID)

	token, err := c.Cookie(authCookie)
	if err != nil || token == "" {
		c.Redirect(http.StatusFound, fallback)
		return
	}

	client, err := nuon.New(nuon.WithURL(h.cfg.APIUrl), nuon.WithAuthToken(token))
	if err != nil {
		h.l.Error("failed to create nuon client", zap.Error(err))
		c.Redirect(http.StatusFound, fallback)
		return
	}

	connection, err := client.CreateVCSConnectionCallback(c.Request.Context(), &models.ServiceCreateConnectionCallbackRequest{
		GithubInstallID: &installationID,
		OrgID:           &orgID,
	})
	if err != nil {
		h.l.Error("vcs connection callback failed", zap.Error(err))
		c.Redirect(http.StatusFound, fallback)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/%s/apps?vcs-connected=%s", orgID, connection.ID))
}
