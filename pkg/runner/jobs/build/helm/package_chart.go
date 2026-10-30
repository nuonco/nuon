package helm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"go.uber.org/zap"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"

	"github.com/nuonco/nuon/pkg/runner/jobs/buildinputs"
	"github.com/nuonco/nuon/pkg/sourceidentity"
)

func (h *handler) packageChart(ctx context.Context, l *zap.Logger) (string, error) {
	if h.state.plan.Src != nil && h.state.plan.Src.URL != "" {
		if err := buildinputs.Verify(h.state.plan, h.state.source); err != nil {
			return "", err
		}
		return h.loadAndPackageChart(ctx, l, h.state.workspace.Source().AbsPath())
	}
	return h.packageChartFromRepoConfig(ctx, l)
}

func (h *handler) packageChartFromRepoConfig(ctx context.Context, l *zap.Logger) (string, error) {
	cfg := h.state.cfg.HelmRepoConfig
	client := newChartRepositoryClient()
	defer client.CloseIdleConnections()
	version := cfg.Version
	if expected := h.state.plan.SourceIdentity; expected != nil && expected.Captured && expected.ChartVersion != "" {
		version = expected.ChartVersion
	}
	data, err := fetchRepositoryChart(ctx, client, cfg.RepoURL, cfg.Chart, version)
	if err != nil {
		return "", fmt.Errorf("unable to pull helm chart: %w", err)
	}
	dir, err := materializeChartArchive(h.state.workspace.FS(), bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	pulledSource, err := h.state.source.Sub(dir)
	if err != nil {
		return "", fmt.Errorf("unable to open pulled chart source: %w", err)
	}
	if err := h.state.source.Close(); err != nil {
		pulledSource.Close()
		return "", err
	}
	h.state.source = pulledSource
	if expected := h.state.plan.SourceIdentity; expected != nil {
		chart, err := loadChart(h.state.source, ".")
		if err != nil {
			return "", err
		}
		selected := *expected
		selected.ChartVersion = chart.Metadata.Version
		captured, err := sourceidentity.Capture(h.state.source, &selected)
		if err != nil {
			return "", err
		}
		if expected.Captured && !sourceidentity.Matches(expected, captured) {
			return "", fmt.Errorf("remote chart differs from captured source")
		}
		if err := buildinputs.Claim(ctx, h.apiClient, h.state.jobID, h.state.jobExecutionID, captured); err != nil {
			return "", err
		}
		h.state.plan.SourceIdentity = captured
	}
	return h.loadAndPackageChart(ctx, l, ".")
}

func (h *handler) loadAndPackageChart(ctx context.Context, l *zap.Logger, chartDir string) (string, error) {
	chart, err := loadChart(h.state.source, chartDir)
	if err != nil {
		return "", fmt.Errorf("unable to load chart: %w", err)
	}
	resolved, err := h.state.source.Resolve(chartDir)
	if err != nil {
		return "", err
	}
	info, err := h.state.source.Lstat(resolved)
	if err != nil {
		return "", err
	}
	// Helm updates directories only. Archives retain their own ignore semantics.
	if len(chart.Metadata.Dependencies) > 0 && info.IsDir() {
		locals, err := localDependencyCharts(h.state.source, resolved, chart.Metadata.Dependencies)
		if err != nil {
			return "", err
		}
		client := newChartRepositoryClient()
		err = refreshChartDependencies(ctx, client, locals, chart)
		client.CloseIdleConnections()
		if err != nil {
			if missing := verifyDependenciesPresent(chart.Metadata.Dependencies, chart); missing != nil {
				return "", fmt.Errorf("%w; dependency refresh failed: %v", missing, err)
			}
			// A failed update must not discard an already-vendored chart.
			l.Warn("dependency update reported an error; verifying vendored dependencies", zap.Error(err))
		}
	}
	if err := verifyDependenciesPresent(chart.Metadata.Dependencies, chart); err != nil {
		return "", err
	}
	h.state.chart, err = publishChart(h.state.arch.FS(), chart)
	if err != nil {
		return "", fmt.Errorf("unable to package chart: %w", err)
	}
	packagePath := filepath.Join(h.state.arch.BasePath(), defaultChartPackageFilename)
	l.Info("successfully packaged chart", zap.String("path", packagePath))
	return packagePath, nil
}

func publishChart(output afero.Fs, chart *chartv2.Chart) (*chartv2.Chart, error) {
	packaged, err := afero.TempFile(output, ".", ".helm-")
	if err != nil {
		return nil, err
	}
	defer output.Remove(packaged.Name())
	defer packaged.Close()
	if err := writeChartArchive(packaged, chart); err != nil {
		return nil, err
	}
	if _, err := packaged.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	files, err := chartArchiveFiles(packaged)
	if err != nil {
		return nil, fmt.Errorf("invalid packaged chart: %w", err)
	}
	chart, err = loadChartFiles(files)
	if err != nil {
		return nil, err
	}
	if _, err := packaged.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	// Exclusive creation refuses existing symlinks; display paths never authorize I/O.
	dest, err := output.OpenFile(defaultChartPackageFilename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	_, copyErr := io.Copy(dest, packaged)
	closeErr := dest.Close()
	if copyErr != nil || closeErr != nil {
		output.Remove(defaultChartPackageFilename)
		return nil, fmt.Errorf("unable to publish chart: write=%v close=%v", copyErr, closeErr)
	}
	return chart, nil
}

// verifyDependenciesPresent prevents incomplete packages (notably missing CRDs).
func verifyDependenciesPresent(declared []*chartv2.Dependency, c *chartv2.Chart) error {
	loaded := map[string]bool{}
	for _, dep := range c.Dependencies() {
		loaded[dep.Name()] = true
	}
	var missing []string
	for _, dep := range declared {
		if loaded[dep.Name] || (dep.Alias != "" && loaded[dep.Alias]) {
			continue
		}
		missing = append(missing, fmt.Sprintf("%s (repository %q)", dep.Name, dep.Repository))
	}
	if len(missing) > 0 {
		return fmt.Errorf("chart dependencies were not packaged: %s; verify the dependency repositories are reachable", strings.Join(missing, ", "))
	}
	return nil
}
