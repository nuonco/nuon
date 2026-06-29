package containerimage

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry"
	"github.com/nuonco/nuon/pkg/oci/imageref"
	"github.com/nuonco/nuon/pkg/oci/updatepolicy"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	srcCfg := h.state.cfg.RepoCfg
	dstCfg := h.state.regCfg

	// imageref owns the rules that turn the planner-provided image+tag into
	// the pull ref and the recorded source identity, so they can never
	// disagree. When the user set an `update_policy` semver constraint, pick
	// the concrete tag by listing tags from the source registry and
	// semver-selecting the highest match; otherwise the pull ref comes
	// straight from the spec.
	spec := imageref.Spec{
		Image:        h.state.cfg.Image,
		Tag:          h.state.cfg.Tag,
		UpdatePolicy: h.state.cfg.UpdatePolicy,
	}
	var selected string
	if spec.UpdatePolicy != "" {
		l.Info(fmt.Sprintf("resolving image source %s with update_policy %q", spec.Image, spec.UpdatePolicy))
		tags, terr := h.ociResolve.Tags(ctx, srcCfg)
		if terr != nil {
			h.writeErrorResult(ctx, "list tags for update_policy", terr)
			return fmt.Errorf("unable to list source tags: %w", terr)
		}
		var serr error
		selected, serr = updatepolicy.SelectHighestMatching(tags, spec.UpdatePolicy)
		if serr != nil {
			h.writeErrorResult(ctx, "select tag for update_policy", serr)
			return fmt.Errorf("unable to select tag for update_policy %q: %w", spec.UpdatePolicy, serr)
		}
		l.Info(fmt.Sprintf("update_policy %q selected tag %q from %d source tags", spec.UpdatePolicy, selected, len(tags)))
	}
	srcTag := spec.PullRef(selected)

	// Resolve the upstream source ref to its manifest descriptor BEFORE
	// pulling/pushing any blobs. The descriptor's digest is the canonical
	// content address we use to:
	//   - decide whether this build is a no-op (digest matches the previous
	//     build's recorded SourceDigest)
	//   - record source identity on the ComponentBuild row for downstream
	//     drift detection and dependency-aware deploys.
	l.Info(fmt.Sprintf("resolving image source %s:%s", h.state.cfg.Image, srcTag))
	desc, err := h.ociResolve.Resolve(ctx, srcCfg, srcTag)
	if err != nil {
		h.writeErrorResult(ctx, "resolve image source", err)
		return fmt.Errorf("unable to resolve image source: %w", err)
	}

	resolvedDigest := string(desc.Digest)
	noOp := h.state.cfg.PreviousSourceDigest != "" && h.state.cfg.PreviousSourceDigest == resolvedDigest

	if noOp {
		l.Info(fmt.Sprintf(
			"upstream digest %s matches previous build, skipping copy (no-op build)",
			resolvedDigest,
		))
	} else {
		l.Info(fmt.Sprintf("copying image from %s:%s to %s", h.state.cfg.Image, srcTag, h.state.plan.DstTag))
		if _, err := h.ociCopy.Copy(ctx,
			srcCfg,
			srcTag,
			dstCfg,
			h.state.resultTag,
		); err != nil {
			h.writeErrorResult(ctx, "copy image", err)
			return err
		}
	}

	resolvedAt := time.Now().UTC()
	resultReq := registry.ToAPIResult(desc)
	src := spec.Identity(selected)
	resultReq.SourceRef = src.SourceRef
	resultReq.SourceImage = src.SourceImage
	resultReq.ResolvedTag = src.ResolvedTag
	resultReq.SourceDigest = resolvedDigest
	resultReq.SourceMediaType = desc.MediaType
	resultReq.ResolvedAt = resolvedAt.Format(time.RFC3339Nano)
	resultReq.NoOp = noOp

	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}
