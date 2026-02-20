package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type OrgsHandler struct {
	l *zap.Logger
}

func NewOrgsHandler(l *zap.Logger) *OrgsHandler {
	return &OrgsHandler{l: l}
}

func (h *OrgsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs", h.GetOrgs)
	e.GET("/api/orgs/:orgId", h.GetOrg)
	e.GET("/api/orgs/:orgId/accounts", h.GetOrgAccounts)
	e.GET("/api/orgs/:orgId/features", h.GetOrgFeatures)
	return nil
}

func (h *OrgsHandler) GetOrgs(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	orgs, _, err := client.GetOrgs(c.Request.Context(), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, orgs)
}

func (h *OrgsHandler) GetOrg(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	// SetOrgID from route param for this request
	client.SetOrgID(c.Param("orgId"))

	org, err := client.GetOrg(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, org)
}

func (h *OrgsHandler) GetOrgAccounts(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	client.SetOrgID(c.Param("orgId"))

	invites, _, err := client.GetOrgInvites(c.Request.Context(), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, invites)
}

func (h *OrgsHandler) GetOrgFeatures(c *gin.Context) {
	// Features endpoint — returns empty array for now as features
	// are a future capability.
	respondJSON(c, http.StatusOK, []any{})
}

// paginationFromQuery extracts pagination params from query string.
func paginationFromQuery(c *gin.Context) *models.GetPaginatedQuery {
	q := &models.GetPaginatedQuery{}
	if limit := c.Query("limit"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			q.Limit = v
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if v, err := strconv.Atoi(offset); err == nil {
			q.Offset = v
		}
	}
	return q
}
