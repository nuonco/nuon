package service

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	posthog "github.com/posthog/posthog-go"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateCLIEventRequest struct {
	Command    string `json:"command" validate:"required,max=256"`
	CLIVersion string `json:"cli_version" validate:"required,max=64"`
	Success    bool   `json:"success"`
	DurationMS int64  `json:"duration_ms" validate:"gte=0"`
	OS         string `json:"os" validate:"max=32"`
	Arch       string `json:"arch" validate:"max=32"`
	Agent      string `json:"agent" validate:"max=64"`
}

// @ID						CreateCLIEvent
// @Summary				Record a CLI command run for product analytics
// @Description.markdown	create_cli_event.md
// @Tags					general
// @Accept					json
// @Param					req	body	CreateCLIEventRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				202
// @Router					/v1/general/cli-events [post]
func (s *service) CreateCLIEvent(ctx *gin.Context) {
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateCLIEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := s.v.Struct(req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if req.CLIVersion == "development" {
		ctx.Status(http.StatusAccepted)
		return
	}

	distinctID := acct.Email
	if distinctID == "" {
		distinctID = acct.Subject
	}
	orgID := ctx.GetHeader("X-Nuon-Org-ID")
	if !slices.Contains(acct.OrgIDs, orgID) {
		orgID = ""
	}

	s.productAnalytics.Capture(distinctID, "cli_command", orgID, posthog.NewProperties().
		Set("command", req.Command).
		Set("cli_version", req.CLIVersion).
		Set("success", req.Success).
		Set("duration_ms", req.DurationMS).
		Set("os", req.OS).
		Set("arch", req.Arch).
		Set("agent", req.Agent))

	ctx.Status(http.StatusAccepted)
}
