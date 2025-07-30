package helm

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	release "helm.sh/helm/v4/pkg/release/v1"
)

// TODO(jm): pull out the helm resources and their statuses from the release, and write them to the api
func (h *handler) createAPIResultRequest(l *zap.Logger, rel *release.Release, helmPlanContents HelmPlanContents) (*models.ServiceCreateRunnerJobExecutionResultRequest, error) {
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		// JUST FOR NOW: the plan is going into both
		Success: true,
	}
	// if the helm plant contents is empty, the request should be empty
	if helmPlanContents.Diff == "" && helmPlanContents.Op == "" {
		return req, nil
	}
	// otherwise, the plan is provided in both of the relevant fields. we must gzip for storage and b64 encode for transit.
	// 1. marshall
	byts, err := json.Marshal(helmPlanContents)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshall helm plan contents")
	}
	// 2. gzip
	l.Info("zipping helm plan")
	var zippedBytes bytes.Buffer
	gzipWriter := gzip.NewWriter(&zippedBytes)
	gzipWriter.Write(byts)
	gzipWriter.Close()
	l.Debug("zipped helm plan", zap.Int("bytes.zipped", len(zippedBytes.Bytes())))

	// 3. base64 encode (urlsafe)
	l.Info("base64-encoding helm plan")
	encodedString := base64.URLEncoding.EncodeToString(zippedBytes.Bytes())
	l.Debug("base64-encoded helm plan", zap.Int("bytes.b64", len(encodedString)))

	req.ContentsCompressed = encodedString
	req.ContentsDisplayCompressed = encodedString
	return req, nil
}

// NOTE(fd): we gzip and base64 encode the payloads
