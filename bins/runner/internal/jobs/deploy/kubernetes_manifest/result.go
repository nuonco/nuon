package kubernetes_manifest

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"

	types "github.com/nuonco/nuon/pkg/types/approvals"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	release "helm.sh/helm/v4/pkg/release/v1"
)

// TODO(fd): encrypt and such
func (h *handler) createAPIResultRequest(rel *release.Release, l *zap.Logger, planContents types.KubernetesManifestPlanContents) (*models.ServiceCreateRunnerJobExecutionResultRequest, error) {
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}

	// read plan contents into json
	byts, err := json.Marshal(planContents)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal plan contents")
	}

	// gzip
	l.Info("zipping kubernetes_manifest plan")
	var zippedBytes bytes.Buffer
	gzipWriter := gzip.NewWriter(&zippedBytes)
	gzipWriter.Write(byts)
	gzipWriter.Close()
	l.Debug("zipped kubernetes_manifest plan", zap.Int("bytes.zipped", len(zippedBytes.Bytes())))

	// base64 encrypt
	encodedString := base64.URLEncoding.EncodeToString(zippedBytes.Bytes())
	req.ContentsCompressed = encodedString
	req.ContentsDisplayCompressed = encodedString

	return req, nil
}
