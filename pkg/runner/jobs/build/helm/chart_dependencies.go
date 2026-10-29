package helm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"helm.sh/helm/v4/pkg/chart/common"
	"helm.sh/helm/v4/pkg/chart/loader/archive"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"sigs.k8s.io/yaml"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

// Validate every file:// target before attempting any refresh or vendored
// fallback. Missing in-repository targets may be vendored; escaping ones may not.
func localDependencyCharts(source *sourcefs.FS, chartDir string, deps []*chartv2.Dependency) (map[string]*chartv2.Chart, error) {
	charts := map[string]*chartv2.Chart{}
	for _, dep := range deps {
		if !strings.HasPrefix(dep.Repository, "file://") {
			continue
		}
		local := strings.TrimPrefix(dep.Repository, "file://")
		if local == "" || filepath.IsAbs(local) || strings.ContainsAny(local, "\\:\x00") {
			return nil, fmt.Errorf("dependency %q must use a repository-relative file:// path", dep.Name)
		}
		target := filepath.Join(chartDir, local)
		if _, err := source.Relative(target); err != nil {
			return nil, err
		}
		resolved, err := source.Resolve(target)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("dependency %q escapes chart source: %w", dep.Name, err)
		}
		if err == nil {
			chart, err := loadChart(source, resolved)
			if err != nil {
				return nil, fmt.Errorf("unable to load local dependency %q: %w", dep.Name, err)
			}
			charts[dep.Repository] = chart
		}
	}
	return charts, nil
}

// Helm's Manager and internal resolver require host paths. Resolve with Helm's
// exported in-memory index/tag APIs instead, committing no changes on failure.
func refreshChartDependencies(ctx context.Context, client *http.Client, locals map[string]*chartv2.Chart, chart *chartv2.Chart) error {
	refreshed := map[string]*chartv2.Chart{}
	locked := make([]*chartv2.Dependency, 0, len(chart.Metadata.Dependencies))
	fetcher := newChartRepositoryFetcher(client)
	for _, dep := range chart.Metadata.Dependencies {
		constraint, err := semver.NewConstraint(dep.Version)
		if err != nil {
			return fmt.Errorf("invalid version constraint for %q: %w", dep.Name, err)
		}
		var child *chartv2.Chart
		switch {
		case dep.Repository == "":
			for _, candidate := range chart.Dependencies() {
				if candidate.Name() == dep.Name || (dep.Alias != "" && candidate.Name() == dep.Alias) {
					child = candidate
					break
				}
			}
			if child == nil {
				return fmt.Errorf("local dependency %q is not vendored", dep.Name)
			}
		case strings.HasPrefix(dep.Repository, "file://"):
			child = locals[dep.Repository]
			if child == nil {
				return fmt.Errorf("local dependency %q is unavailable", dep.Name)
			}
		default:
			var data []byte
			data, err = fetcher.fetch(ctx, dep.Repository, dep.Name, dep.Version)
			if err == nil {
				files, archiveErr := chartArchiveFiles(bytes.NewReader(data))
				err = archiveErr
				if err == nil {
					child, err = loadChartFiles(files)
				}
			}
		}
		if err != nil {
			return fmt.Errorf("unable to refresh dependency %q: %w", dep.Name, err)
		}
		if child.Name() != dep.Name && child.Name() != dep.Alias {
			return fmt.Errorf("dependency %q returned chart %q", dep.Name, child.Name())
		}
		version, err := semver.NewVersion(child.Metadata.Version)
		if err != nil || !constraint.Check(version) {
			return fmt.Errorf("dependency %q version %q does not satisfy %q", dep.Name, child.Metadata.Version, dep.Version)
		}
		if previous := refreshed[child.Name()]; previous != nil && previous.Metadata.Version != child.Metadata.Version {
			return fmt.Errorf("different versions of dependency %q cannot share a packaged chart path", child.Name())
		}
		refreshed[child.Name()] = child
		versionString := child.Metadata.Version
		if dep.Repository == "" {
			versionString = dep.Version // Helm retains constraints for vendored-only dependencies.
		}
		locked = append(locked, &chartv2.Dependency{Name: dep.Name, Repository: dep.Repository, Version: versionString})
	}
	// Match Helm internal/resolver.HashReq without its filesystem-bound resolver.
	data, err := json.Marshal([2][]*chartv2.Dependency{chart.Metadata.Dependencies, locked})
	if err != nil {
		return err
	}
	lock := &chartv2.Lock{Generated: time.Now(), Digest: fmt.Sprintf("sha256:%x", sha256.Sum256(data)), Dependencies: locked}
	if chart.Lock != nil && chart.Lock.Digest == lock.Digest {
		lock = chart.Lock
	}
	// Helm prunes obsolete archives, not directory-backed children. Reload the
	// retained raw files to preserve that distinction even when names differ
	// from their source paths, without changing the vendored set on failure.
	var retainedFiles []*archive.BufferedFile
	for _, file := range chart.Raw {
		if path.Dir(file.Name) == "charts" && strings.HasSuffix(file.Name, ".tgz") {
			continue
		}
		retainedFiles = append(retainedFiles, &archive.BufferedFile{Name: file.Name, ModTime: file.ModTime, Data: file.Data})
	}
	retained, err := loadChartFiles(retainedFiles)
	if err != nil {
		return err
	}
	if chart.Metadata.APIVersion == chartv2.APIVersionV1 {
		data, err := yaml.Marshal(lock)
		if err != nil {
			return err
		}
		files := make([]*common.File, 0, len(chart.Files)+1)
		for _, file := range chart.Files {
			if file.Name != "requirements.lock" {
				files = append(files, file)
			}
		}
		chart.Files = append(files, &common.File{Name: "requirements.lock", ModTime: lock.Generated, Data: data})
	}
	children := make([]*chartv2.Chart, 0, len(chart.Dependencies())+len(refreshed))
	for _, child := range retained.Dependencies() {
		if refreshed[child.Name()] == nil {
			children = append(children, child)
		}
	}
	// Preserve declaration order and include a shared aliased chart only once.
	for _, dep := range chart.Metadata.Dependencies {
		for _, name := range []string{dep.Name, dep.Alias} {
			if child := refreshed[name]; child != nil {
				children = append(children, child)
				delete(refreshed, name)
			}
		}
	}
	chart.SetDependencies(children...)
	chart.Lock = lock
	return nil
}
