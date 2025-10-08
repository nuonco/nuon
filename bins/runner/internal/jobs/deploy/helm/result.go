package helm

import (
	"encoding/json"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	release "helm.sh/helm/v4/pkg/release/v1"

	"github.com/powertoolsdev/mono/pkg/plans"
)

// TODO(jm): pull out the helm resources and their statuses from the release, and write them to the api
func (h *handler) createAPIResultRequest(l *zap.Logger, rel *release.Release, helmPlanContents HelmPlanContents) (*models.ServiceCreateRunnerJobExecutionResultRequest, error) {
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}

	if helmPlanContents.Diff == "" && helmPlanContents.Op == "" {
		return req, nil
	}

	jsonData, err := json.Marshal(helmPlanContents)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal helm plan contents")
	}

	encodedPlan, err := plans.CompressPlan(jsonData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compress plan")
	}

	l.Debug("base64-encoded helm plan", zap.Int("bytes.b64", len(encodedPlan)))

	req.ContentsCompressed = encodedPlan
	req.ContentsDisplayCompressed = encodedPlan
	return req, nil
}
