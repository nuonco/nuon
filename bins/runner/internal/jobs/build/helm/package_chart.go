package helm

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/downloader"
	"helm.sh/helm/v4/pkg/getter"
	"helm.sh/helm/v4/pkg/registry"
	"helm.sh/helm/v4/pkg/repo/v1"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
)

func (h *handler) packageChart(l *zap.Logger) (string, error) {
	chartDir := h.state.workspace.Source().AbsPath()
	dstDir := h.state.arch.TmpDir()

	// load chart
	chart, err := loader.Load(chartDir)
	if err != nil {
		return "", fmt.Errorf("unable to load chart: %w", err)
	}
	l.Info("succesfully loaded chart", zap.String("chart_dir", chartDir), zap.String("dst_dir", dstDir))

	// check for dependencies
	dependencies := chart.Metadata.Dependencies
	dep_repos := map[string]string{}
	if len(dependencies) > 0 {
		l.Info("dependencies: chart has dependencies", zap.String("chart_dir", chartDir), zap.String("dst_dir", dstDir))
		// 1. add repos and update dependencies
		h.addDependencyReposAndUpdate(l, chartDir, dep_repos)
		// 2. reload the chart now that the deps are in place
		chart, err = loader.Load(chartDir)
		if err != nil {
			return "", fmt.Errorf("dependencies: unable to load chart with dependencies: %w", err)
		}
	}

	// package the chart
	packagePath, err := chartutil.Save(chart, dstDir)
	if err != nil {
		return "", fmt.Errorf("unable to package chart: %w", err)
	}
	l.Info("succesfully packaged chart", zap.String("path", packagePath))

	return packagePath, nil
}

/***
*
* cannibalized from the helm `repo add` and `dependency update` commands.
*
* NOTE(fd): we do not copy all of the code 1:1 so some functionality (e.g. handling deprecated hashes) is not included
*
***/
// Repositories that have been permanently deleted and no longer work
var deprecatedRepos = map[string]string{
	"//kubernetes-charts.storage.googleapis.com":           "https://charts.helm.sh/stable",
	"//kubernetes-charts-incubator.storage.googleapis.com": "https://charts.helm.sh/incubator",
}

func (h *handler) addRepo(l *zap.Logger, out io.Writer, settings *cli.EnvSettings, chartDir, name, repository string) error {
	// NOTE(fd): lifted from here: https://github.com/helm/helm/blob/main/cmd/helm/repo_add.go#L106
	fmt.Fprintf(out, "dependencies: preparing to add repo %s from %s", name, repository)
	// Block deprecated repos
	allowDeprecatedRepos := false // hoisted into a var in case we need to do logic here later
	if !allowDeprecatedRepos {    // we block deprecated reps by default for now
		for oldURL, newURL := range deprecatedRepos {
			if strings.Contains(repository, oldURL) {
				return fmt.Errorf("repo %q is no longer available; try %q instead", repository, newURL)
			}
		}
	}

	// Ensure the file directory exists as it is required for file locking
	err := os.MkdirAll(filepath.Dir(settings.RepositoryConfig), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Acquire a file lock for process synchronization
	repoFileExt := filepath.Ext(settings.RepositoryConfig)
	var lockPath string
	if len(repoFileExt) > 0 && len(repoFileExt) < len(settings.RepositoryConfig) {
		lockPath = strings.TrimSuffix(settings.RepositoryConfig, repoFileExt) + ".lock"
	} else {
		lockPath = settings.RepositoryConfig + ".lock"
	}
	fileLock := flock.New(lockPath)
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		return err
	}

	b, err := os.ReadFile(settings.RepositoryConfig)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}
	// due to the way this is implemented, it does not support git or private repos
	c := repo.Entry{
		Name: name,
		URL:  repository,
	}

	// Check if the repo name is legal
	if strings.Contains(name, "/") {
		return errors.Errorf("repository name (%s) contains '/', please specify a different name without '/'", name)
	}

	// we skip the check for the repo since it should not exist
	// this can break if we use dependencies w/ the same repo but that seems unlikely

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return err
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		return errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", repository)
	}

	f.Update(&c)

	if err := f.WriteFile(settings.RepositoryConfig, 0600); err != nil {
		return err
	}
	fmt.Fprintf(out, "%q has been added to your repositories\n", name)
	return nil
}

func (h *handler) addDependencyReposAndUpdate(l *zap.Logger, chartDir string, repos map[string]string) error {
	// NOTE(fd): lifted from here: https://github.com/helm/helm/blob/main/cmd/helm/dependency_update.go#L47
	hcLog := log.NewHClog(l)
	lw := hcLog.StandardWriter(&hclog.StandardLoggerOptions{})

	// make some settings
	settings := cli.New()
	settings.BurstLimit = 10
	settings.QPS = 5

	// add repos
	for url, name := range repos {
		h.addRepo(l, lw, settings, chartDir, name, url)
	}

	// make a helm client
	client := action.NewDependency()

	// Create a new registry client
	opts := []registry.ClientOption{
		registry.ClientOptDebug(true),
		registry.ClientOptEnableCache(false),
		registry.ClientOptWriter(lw),
		registry.ClientOptPlainHTTP(),
	}
	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return err
	}
	man := &downloader.Manager{
		Out:              lw,
		ChartPath:        chartDir,
		Keyring:          client.Keyring,
		SkipUpdate:       client.SkipRefresh,
		Getters:          getter.All(settings),
		RegistryClient:   registryClient,
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
		ContentCache:     settings.ContentCache,
		Debug:            true,
	}
	if client.Verify {
		man.Verify = downloader.VerifyAlways
	}

	// update dependencies
	l.Info("dependencies: preparing to to update dependencies", zap.String("settings", fmt.Sprintf("%+v", settings)))
	err = man.Update()
	if err == nil {
		client.List(chartDir, lw)
		return nil
	} else {
		l.Info("dependencies: failed to update dependencies", zap.String("error", fmt.Sprintf("%s", err)))
		return err
	}
}
