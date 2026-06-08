package containerimage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry"
	"github.com/nuonco/nuon/pkg/oci/updatepolicy"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	srcCfg := h.state.cfg.RepoCfg
	dstCfg := h.state.regCfg

	// When the user has set an `update_policy` semver constraint on the
	// component, pick the concrete tag to resolve by listing tags from
	// the source registry and semver-selecting the highest match. The
	// constraint is recorded as the SourceRef so we keep a faithful
	// record of what the user asked for; ResolvedTag records the
	// concrete tag we actually pulled.
	//
	// When unset, use h.state.cfg.Tag literally.
	srcTag := h.state.cfg.Tag
	policy := h.state.cfg.UpdatePolicy
	if policy != "" {
		l.Info(fmt.Sprintf("resolving image source %s with update_policy %q", h.state.cfg.Image, policy))
		tags, terr := h.ociResolve.Tags(ctx, srcCfg)
		if terr != nil {
			h.writeErrorResult(ctx, "list tags for update_policy", terr)
			return fmt.Errorf("unable to list source tags: %w", terr)
		}
		selected, serr := updatepolicy.SelectHighestMatching(tags, policy)
		if serr != nil {
			h.writeErrorResult(ctx, "select tag for update_policy", serr)
			return fmt.Errorf("unable to select tag for update_policy %q: %w", policy, serr)
		}
		l.Info(fmt.Sprintf("update_policy %q selected tag %q from %d source tags", policy, selected, len(tags)))
		srcTag = selected
	}

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
	resultReq.SourceRef, resultReq.SourceImage, resultReq.ResolvedTag = buildSourceIdentity(h.state.cfg.Image, h.state.cfg.Tag, policy, srcTag)
	resultReq.SourceDigest = resolvedDigest
	resultReq.SourceMediaType = desc.MediaType
	resultReq.ResolvedAt = resolvedAt.Format(time.RFC3339Nano)
	resultReq.NoOp = noOp

	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}

// buildSourceIdentity computes the SourceRef, SourceImage, and ResolvedTag
// strings from the planner-provided image+tag.
//
//   - SourceRef is what the user wrote: "<image>:<tag>" for tag-based refs,
//     "<image>@<digest>" for digest-pinned refs, or "<image>:<update_policy>"
//     when an update_policy semver constraint is set.
//   - SourceImage is always the image repository portion (no tag, no digest).
//   - ResolvedTag is the tag the runner actually pulled — for update_policy
//     this is the semver-selected tag; for plain tag-based refs it is the
//     literal tag; empty for digest-pinned refs.
//
// A "@sha256:..." substring in either Image or Tag indicates a digest-pinned
// ref. The plan currently splits these inconsistently (the planner copies the
// Tag field straight from the user spec) so we accept both shapes.
//
// When updatePolicy is non-empty it takes precedence over Tag for the
// SourceRef field — the user's "intent" was the constraint, not the tag we
// selected on this particular build — and selectedTag is recorded as the
// ResolvedTag.
func buildSourceIdentity(image, tag, updatePolicy, selectedTag string) (sourceRef, sourceImage, resolvedTag string) {
	if updatePolicy != "" {
		sourceImage = image
		sourceRef = fmt.Sprintf("%s:%s", image, updatePolicy)
		resolvedTag = selectedTag
		return sourceRef, sourceImage, resolvedTag
	}
	switch {
	case strings.Contains(tag, "@sha256:"):
		// User wrote `nginx@sha256:...` and the planner kept that as the tag.
		sourceImage = image
		sourceRef = tag
		if image == "" {
			// Best-effort split: take the part before "@sha256:" as the image.
			sourceImage = strings.SplitN(tag, "@sha256:", 2)[0]
		}
		resolvedTag = ""
	case strings.HasPrefix(tag, "sha256:"):
		// Tag is a bare digest. Build the canonical "<image>@<digest>" form.
		sourceImage = image
		sourceRef = fmt.Sprintf("%s@%s", image, tag)
		resolvedTag = ""
	case strings.Contains(image, "@sha256:"):
		// Digest baked into image; tag (if any) is ignored.
		sourceImage = strings.SplitN(image, "@sha256:", 2)[0]
		sourceRef = image
		resolvedTag = ""
	default:
		sourceImage = image
		sourceRef = fmt.Sprintf("%s:%s", image, tag)
		resolvedTag = tag
	}
	return sourceRef, sourceImage, resolvedTag
}
