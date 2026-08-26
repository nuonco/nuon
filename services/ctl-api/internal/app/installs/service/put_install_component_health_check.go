package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

const (
	customCheckProvider        = "custom"
	customCheckKind            = "CustomCheck"
	maxCustomCheckMessageBytes = 1024
	maxCustomCheckDetailsBytes = 16 * 1024

	// maxCustomCheckStaleAfter must not exceed the evaluator's retention
	// window, or it would promise to remember reports the evaluator drops.
	maxCustomCheckStaleAfter = 60 * time.Minute
)

var validCustomCheckStatuses = map[string]bool{
	string(app.InstallComponentResourceHealthHealthy):   true,
	string(app.InstallComponentResourceHealthDegraded):  true,
	string(app.InstallComponentResourceHealthUnhealthy): true,
	string(app.InstallComponentResourceHealthUnknown):   true,
}

// validateCustomCheckStatus limits status to the four verdicts a custom
// check can report; "progressing" is runner-only and doesn't apply here.
func validateCustomCheckStatus(status string) error {
	if !validCustomCheckStatuses[status] {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid status %q", status),
			Description: "status must be one of: healthy, degraded, unhealthy, unknown",
		}
	}
	return nil
}

// customCheckNameRe keeps check names safe to render across ClickHouse,
// dashboards, and Slack, and stable as a key across reports.
var customCheckNameRe = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9._-]{0,98}[a-zA-Z0-9])?$`)

func validateCustomCheckName(name string) error {
	if !customCheckNameRe.MatchString(name) {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid check name %q", name),
			Description: "check name must be 1-100 characters of letters, digits, dots, dashes, or underscores, starting and ending with a letter or digit",
		}
	}
	return nil
}

type PutInstallComponentHealthCheckRequest struct {
	Status  string          `json:"status" validate:"required"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details,omitempty" swaggertype:"object"`

	// StaleAfter is how long this report stays trustworthy, e.g. "30m"; past
	// it the check reads as unknown. Defaults to 5m — set higher for slower pushers.
	StaleAfter string `json:"stale_after,omitempty"`
}

func (r *PutInstallComponentHealthCheckRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	if _, err := r.staleAfter(); err != nil {
		return err
	}
	return validateCustomCheckStatus(r.Status)
}

func (r *PutInstallComponentHealthCheckRequest) staleAfter() (time.Duration, error) {
	raw := strings.TrimSpace(r.StaleAfter)
	if raw == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, stderr.ErrUser{
			Err:         fmt.Errorf("invalid stale_after %q: %w", raw, err),
			Description: "stale_after must be a duration like 30m",
		}
	}
	if d <= 0 || d > maxCustomCheckStaleAfter {
		return 0, stderr.ErrUser{
			Err:         fmt.Errorf("stale_after %s out of range", d),
			Description: fmt.Sprintf("stale_after must be between 1s and %s", maxCustomCheckStaleAfter),
		}
	}
	return d, nil
}

// @ID						PutInstallComponentHealthCheck
// @Summary				report a custom component health check
// @Description			Lets an external system (a vendor's CI, a Datadog monitor webhook, a custom action) report a named health signal for a component. The report is written as a resource observation with provider "custom", so it flows through the same live explorer, evaluator, alerting, and timeline as runner-reported resources. Requires the component-health feature.
// @Param					install_id		path	string									true	"install ID"
// @Param					component_id	path	string									true	"component ID"
// @Param					check_name		path	string									true	"check name"
// @Param					req				body	PutInstallComponentHealthCheckRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallComponentResourceState
// @Router					/v1/installs/{install_id}/components/{component_id}/health/checks/{check_name} [put]
func (s *service) PutInstallComponentHealthCheck(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")
	checkName := ctx.Param("check_name")
	if err := validateCustomCheckName(checkName); err != nil {
		ctx.Error(err)
		return
	}

	var req PutInstallComponentHealthCheckRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireComponentHealthFeature(ctx, org); err != nil {
		ctx.Error(err)
		return
	}

	row, err := s.putInstallComponentHealthCheck(ctx, org.ID, installID, componentID, checkName, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to report component health check: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, row)
}

func (s *service) putInstallComponentHealthCheck(ctx context.Context, orgID, installID, componentID, checkName string, req PutInstallComponentHealthCheckRequest) (*app.InstallComponentResourceState, error) {
	ic, err := s.findInstallComponent(ctx, orgID, installID, componentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install component: %w", err)
	}

	details := ""
	if len(req.Details) > 0 {
		details = string(req.Details)
	}

	staleAfter, err := req.staleAfter()
	if err != nil {
		return nil, err
	}

	row := app.InstallComponentResourceState{
		OrgID:              orgID,
		InstallID:          installID,
		InstallComponentID: ic.ID,
		ComponentID:        ic.ComponentID,
		Source:             app.InstallComponentResourceSourceComponent,
		Provider:           customCheckProvider,
		Kind:               customCheckKind,
		Name:               checkName,
		Health:             req.Status,
		Message:            boundCustomCheckMessage(req.Message),
		Details:            boundCustomCheckDetails(details),
		ObservedAt:         time.Now(),
		StaleAfterSeconds:  uint32(staleAfter.Seconds()),
	}

	if res := s.chDB.WithContext(ctx).Create(&row); res.Error != nil {
		return nil, fmt.Errorf("unable to write custom health check: %w", res.Error)
	}

	return &row, nil
}

// boundCustomCheckMessage caps a caller-supplied message so a single report
// can't blow up storage; the cut respects UTF-8 rune boundaries.
func boundCustomCheckMessage(msg string) string {
	if len(msg) <= maxCustomCheckMessageBytes {
		return msg
	}
	return strings.ToValidUTF8(msg[:maxCustomCheckMessageBytes], "")
}

// boundCustomCheckDetails mirrors runners/service.boundDetails: truncating
// mid-JSON would corrupt the blob, so it's replaced by a marker instead.
func boundCustomCheckDetails(details string) string {
	if len(details) > maxCustomCheckDetailsBytes {
		return `{"_truncated":true}`
	}
	return details
}
