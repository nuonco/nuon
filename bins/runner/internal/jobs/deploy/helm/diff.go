package helm

import (
	"bytes"

	"github.com/databus23/helm-diff/v3/diff"
	"github.com/databus23/helm-diff/v3/manifest"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	release "helm.sh/helm/v4/pkg/release/v1"
	"k8s.io/client-go/rest"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/zapwriter"
)

const (
	// NOTE(jm): this can also be template or simple
	defaultOutputFormat string = "diff"
)

func (h *handler) logDiff(l *zap.Logger, old, new map[string]*manifest.MappingResult) error {
	out := zapwriter.New(l, zapcore.DebugLevel, "helm-diff")

	opts := &diff.Options{
		OutputFormat:  "simple",
		OutputContext: 10,
		FindRenames:   0.5,
		ShowSecrets:   false,
	}

	changesExist := diff.Manifests(old, new, opts, out)
	if !changesExist {
		l.Info("no changes found")
	}
	return nil
}

func (h *handler) diffReport(currentSpec, targetSpec map[string]*manifest.MappingResult) (*diff.Report, error) {
	opts := diff.Options{
		StripTrailingCR: true,
		ShowSecrets:     false,
	}

	report, err := diff.ManifestReport(currentSpec, targetSpec, &opts)
	if err != nil {
		// l.Error("unable to get diff ", err)
		return nil, errors.Wrap(err, "unable to calculate diff")
	}

	return report, nil
}

func (h *handler) diff(currentSpec, targetSpec map[string]*manifest.MappingResult) ([]byte, error) {
	opts := &diff.Options{
		OutputFormat:  "simple",
		OutputContext: 10,
		FindRenames:   0.5,
		ShowSecrets:   false,
	}

	// Use the buffer to capture the diff output
	var buffer bytes.Buffer
	changesExist := diff.Manifests(currentSpec, targetSpec, opts, &buffer)

	if !changesExist {
		return make([]byte, 0), nil
	}

	return buffer.Bytes(), nil
}

// getDiff compares old and new manifest bytes and returns the diff as a string.
// release and taregt are old and new manifest bytes.
func (h *handler) getDiff(l *zap.Logger, kubeCfg *rest.Config, release, target *release.Release, namespace string) ([]byte, *diff.Report, error) {
	actionConfig, err := helm.ActionConfigV3(l, kubeCfg, namespace)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get helm action config to calculate diff")
	}

	var releaseManifest, targetManifest []byte
	if release != nil {
		releaseManifest = []byte(release.Manifest)
	}
	if target != nil {
		targetManifest = []byte(target.Manifest)
	}

	releaseResources, targetResources, err := manifest.Generate(
		actionConfig,
		releaseManifest,
		targetManifest,
	)

	currentSpecs := make(map[string]*manifest.MappingResult)
	if releaseResources != nil && release != nil {
		currentSpecs = manifest.Parse(string(releaseResources), release.Namespace, false)
	}

	newSpec := make(map[string]*manifest.MappingResult)
	if targetResources != nil && target != nil {
		newSpec = manifest.Parse(string(targetResources), target.Namespace, false)
	}

	diff, err := h.diff(currentSpecs, newSpec)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to generate diff")
	}
	diffReport, err := h.diffReport(currentSpecs, newSpec)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to generate diff report")
	}

	return diff, diffReport, nil
}
