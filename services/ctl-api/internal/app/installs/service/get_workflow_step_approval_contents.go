package service

import (
	"bytes"
	"compress/gzip"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetWorkflowStepApprovalContents
// @Summary				get a workflow step approval contents
// @Description.markdown	get_workflow_step_approval_contents.md
// @Param					workflow_id			path	string	true	"workflow id"
// @Param					step_id	path	string	true	"step id"
// @Param					approval_id			path	string	true	"approval id"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	interface{}
// @Header					200	{string}	Content-Encoding	"gzip"
// @Router					/v1/workflows/{workflow_id}/steps/{step_id}/approvals/{approval_id}/contents  [GET]
func (s *service) GetWorkflowStepApprovalContents(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	workflowID := ctx.Param("workflow_id")
	stepID := ctx.Param("step_id")
	approvalID := ctx.Param("approval_id")

	_, err = s.getWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow step"))
		return
	}

	approval, err := s.getWorkflowStepApproval(ctx, org.ID, approvalID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow step approval"))
		return
	}

	// Create a buffer to hold the gzipped data
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	// Write the contents to the gzip writer
	_, err = gzipWriter.Write([]byte(approval.Contents))
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to gzip approval contents"))
		return
	}

	// Close the gzip writer to flush any remaining data
	err = gzipWriter.Close()
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to close gzip writer"))
		return
	}

	// Set the Content-Encoding header to indicate gzip compression
	ctx.Header("Content-Encoding", "gzip")

	// Return the gzipped bytes
	ctx.Data(http.StatusOK, "application/json", buf.Bytes())
}
