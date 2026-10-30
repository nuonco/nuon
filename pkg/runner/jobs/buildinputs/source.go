package buildinputs

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
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
		if !sourceidentity.IsCommit(plan.Src.Ref) || plan.Src.Ref != i.Commit || filepath.Clean(plan.Src.Path) != filepath.Clean(i.Root) {
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

func Report(ctx context.Context, client ResultClient, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest, i *sourceidentity.Identity) error {
	if i != nil {
		metadata, err := apiIdentity(i)
		if err != nil {
			return err
		}
		req.SourceIdentity = metadata
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
