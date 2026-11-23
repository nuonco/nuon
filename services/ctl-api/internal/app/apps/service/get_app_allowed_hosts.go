package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/urlpolicy"
)

type AppAllowedHostsCategory struct {
	Restricted bool     `json:"restricted"`
	Entries    []string `json:"entries"`
}

type AppAllowedHosts struct {
	Mode       string                             `json:"mode"`
	Categories map[string]AppAllowedHostsCategory `json:"categories"`
}

func newAppAllowedHosts(policy *urlpolicy.Policy) AppAllowedHosts {
	out := AppAllowedHosts{
		Mode:       string(policy.Mode()),
		Categories: make(map[string]AppAllowedHostsCategory, len(urlpolicy.Categories)),
	}
	for _, category := range urlpolicy.Categories {
		entries := policy.Entries(category)
		if entries == nil {
			entries = []string{}
		}
		out.Categories[string(category)] = AppAllowedHostsCategory{
			Restricted: policy.Restricted(category),
			Entries:    entries,
		}
	}
	return out
}

// @ID						GetAppAllowedHosts
// @Summary				get the allowed hosts for an app
// @Description.markdown	get_app_allowed_hosts.md
// @Param					app_id	path	string	true	"app ID"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	AppAllowedHosts
// @Router					/v1/apps/{app_id}/allowed-hosts [get]
func (s *service) GetAppAllowedHosts(ctx *gin.Context) {
	if _, err := s.appByNameOrID(ctx, ctx.Param("app_id")); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, newAppAllowedHosts(s.cfg.URLPolicy))
}
