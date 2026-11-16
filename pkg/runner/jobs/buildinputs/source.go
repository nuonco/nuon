package buildinputs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/sourcefs"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type ResultClient interface {
	CreateJobExecutionResult(context.Context, string, string, *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error)
}

func Verify(plan *plantypes.BuildPlan, source *sourcefs.FS) error {
	i := plan.SourceIdentity
	if i == nil {
		return nil
	}
	if err := i.Validate(); err != nil {
		return err
	}
	if plan.Src != nil {
		if !sourceidentity.IsCommit(plan.Src.Ref) || plan.Src.Ref != i.Commit || sourceidentity.CleanRoot(plan.Src.Path) != sourceidentity.CleanRoot(i.Root) {
			return fmt.Errorf("build source selection differs from prepared snapshot")
		}
		return sourceidentity.Verify(source, i)
	}
	if plan.KubernetesManifestBuildPlan != nil && plan.KubernetesManifestBuildPlan.SourceType == "inline" {
		if i.Digest != sourceidentity.Hash(plan.KubernetesManifestBuildPlan.InlineManifest) {
			return fmt.Errorf("inline manifest differs from prepared input")
		}
	}
	return nil
}

func Claim(ctx context.Context, client ResultClient, jobID, executionID string, i *sourceidentity.Identity) error {
	if i == nil {
		return nil
	}
	metadata, err := apiIdentity(i)
	if err != nil {
		return err
	}
	_, err = client.CreateJobExecutionResult(ctx, jobID, executionID, &models.ServiceCreateRunnerJobExecutionResultRequest{SourceCapture: metadata})
	return err
}

// Report sends the result with the identity it was built from and the paths
// outside that identity's root which the build read from source.
func Report(ctx context.Context, client ResultClient, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest, i *sourceidentity.Identity, source *sourcefs.FS) error {
	if i != nil {
		metadata, err := apiIdentity(i)
		if err != nil {
			return err
		}
		req.SourceIdentity = metadata
		if source != nil && strings.HasPrefix(i.Digest, "git:") {
			reads := source.Reads()
			outside := sourceidentity.ExternalReads(i.Root, i.Digest, reads.Read, reads.Stat)
			if count := len(outside.Read) + len(outside.Stat); count > sourceidentity.MaxInputs {
				pkgctx.LoggerOrDefault(ctx, zap.NewNop()).Warn(fmt.Sprintf(
					"This build read %d paths outside its source directory %q. Nuon tracks at most %d, so this build will not be reused and the next run rebuilds it. Move build inputs into the source directory, or set the directory to a common parent.",
					count, i.Root, sourceidentity.MaxInputs))
				outside.Read = outside.Read[:min(len(outside.Read), sourceidentity.MaxInputs+1)]
				outside.Stat = outside.Stat[:min(len(outside.Stat), sourceidentity.MaxInputs+1-len(outside.Read))]
			}
			req.SourceReads = &models.SourceidentityReads{Read: outside.Read, Stat: outside.Stat}
		}
	}
	_, err := client.CreateJobExecutionResult(ctx, jobID, executionID, req)
	return err
}

func apiIdentity(i *sourceidentity.Identity) (*models.SourceidentityIdentity, error) {
	if err := i.Validate(); err != nil {
		return nil, err
	}
	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	var out models.SourceidentityIdentity
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
