package helm

import (
	"bytes"

	"github.com/databus23/helm-diff/v3/diff"
	"github.com/databus23/helm-diff/v3/manifest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

func (h *handler) getDiff(old map[string]*manifest.MappingResult, new map[string]*manifest.MappingResult) (string, error) {
	// Create a buffer to capture the diff output
	var buffer bytes.Buffer

	opts := &diff.Options{
		OutputFormat:  "simple",
		OutputContext: 10,
		FindRenames:   0.5,
		ShowSecrets:   false,
	}

	// Use the buffer to capture the diff output
	changesExist := diff.Manifests(old, new, opts, &buffer)

	if !changesExist {
		return "", nil
	}

	return buffer.String(), nil
}
