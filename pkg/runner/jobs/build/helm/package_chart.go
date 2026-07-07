package helm

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	"net/http"
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
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/downloader"
	"helm.sh/helm/v4/pkg/getter"
	"helm.sh/helm/v4/pkg/registry"
	"helm.sh/helm/v4/pkg/repo/v1"

	"github.com/nuonco/nuon/pkg/runner/log"
)

func (h *handler) packageChart(l *zap.Logger) (string, error) {

	if h.state.plan.Src != nil && h.state.plan.Src.URL != "" {
		chartDir := h.state.workspace.Source().AbsPath()
		dstDir := h.state.arch.TmpDir()

		packagePath, err := h.loadAndPackageChart(l, chartDir, dstDir)
		if err != nil {
			return "", err
		}
		h.state.chartPath = chartDir
		return packagePath, nil
	}

	l.Info("packaging chart from repo config", zap.String("chart", h.state.cfg.HelmRepoConfig.Chart), zap.String("url", h.state.cfg.HelmRepoConfig.RepoURL))
	return h.packageChartFromRepoConfig(l)

}

func (h *handler) packageChartFromRepoConfig(l *zap.Logger) (string, error) {
	l.Debug("initializing helm settings for repo")
	settings, err := h.initHelmSettingsForRepo(l)
	if err != nil {
		return "", fmt.Errorf("unable to initialize helm settings: %w", err)
	}

	repoName := fmt.Sprintf("repo-%d", time.Now().Unix())

	l.Debug("adding helm repo", zap.String("repo_name", repoName), zap.String("repo_url", h.state.cfg.HelmRepoConfig.RepoURL))
	if err := h.addHelmRepo(l, settings, repoName, h.state.cfg.HelmRepoConfig.RepoURL); err != nil {
		l.Error("unable to add helm repo", zap.String("repo_name", repoName), zap.String("repo_url", h.state.cfg.HelmRepoConfig.RepoURL), zap.Error(err))
		return "", fmt.Errorf("unable to add helm repo: %w", err)
	}

	destDir := h.state.workspace.Root()

	l.Debug("pulling helm chart", zap.String("repo_name", repoName), zap.String("chart", h.state.cfg.HelmRepoConfig.Chart))
	chartPath, err := h.pullHelmChart(l, settings, repoName, h.state.cfg.HelmRepoConfig.Chart, destDir)
	if err != nil {
		return "", fmt.Errorf("unable to pull helm chart: %w", err)
	}

	l.Info("successfully pulled chart from repo", zap.String("chart_path", chartPath))

	dstDir := h.state.arch.TmpDir()

	packagePath, err := h.loadAndPackageChart(l, chartPath, dstDir)
	if err != nil {
		return "", err
	}
	h.state.chartPath = chartPath
	return packagePath, nil
}

// loadAndPackageChart loads a chart from the given directory, handles dependencies, and packages it
func (h *handler) loadAndPackageChart(l *zap.Logger, chartDir, dstDir string) (string, error) {
	chart, err := loader.Load(chartDir)
	if err != nil {
		return "", fmt.Errorf("unable to load chart: %w", err)
	}
	l.Info("successfully loaded chart", zap.String("chart_dir", chartDir), zap.String("dst_dir", dstDir))

	dependencies := chart.Metadata.Dependencies
	if len(dependencies) > 0 {
		l.Info("dependencies: chart has dependencies", zap.String("chart_dir", chartDir), zap.Int("count", len(dependencies)))

		// OCI and file:// deps are resolved without a repo entry; registering them errors.
		repoURLs := map[string]struct{}{}
		for _, dep := range dependencies {
			if dep.Repository == "" || registry.IsOCI(dep.Repository) || strings.HasPrefix(dep.Repository, "file://") {
				continue
			}
			repoURLs[dep.Repository] = struct{}{}
		}

		// An update error is not fatal: a chart pulled with deps already vendored
		// stays valid. verifyDependenciesPresent below decides if the package is complete.
		if err := h.addDependencyReposAndUpdate(l, chartDir, repoURLs); err != nil {
			l.Warn("dependency update reported an error; verifying vendored dependencies", zap.Error(err))
		}

		chart, err = loader.Load(chartDir)
		if err != nil {
			return "", fmt.Errorf("unable to load chart with dependencies: %w", err)
		}

		if err := verifyDependenciesPresent(dependencies, chart); err != nil {
			return "", err
		}
	}

	// package the chart
	packagePath, err := chartutil.Save(chart, dstDir)
	if err != nil {
		return "", fmt.Errorf("unable to package chart: %w", err)
	}
	l.Info("successfully packaged chart", zap.String("path", packagePath))

	return packagePath, nil
}

// initHelmSettingsForRepo initializes helm settings for repository operations
func (h *handler) initHelmSettingsForRepo(l *zap.Logger) (*cli.EnvSettings, error) {
	settings := cli.New()
	settings.BurstLimit = 10
	settings.QPS = 5

	return settings, nil
}

// addHelmRepo adds a helm repository
func (h *handler) addHelmRepo(l *zap.Logger, settings *cli.EnvSettings, repoName, repoURL string) error {
	hcLog := log.NewHClog(l)
	lw := hcLog.StandardWriter(&hclog.StandardLoggerOptions{})

	l.Info("adding helm repository", zap.String("repo_name", repoName), zap.String("repo_url", repoURL))
	err := h.addRepo(l, lw, settings, "", repoName, repoURL)
	if err != nil {
		l.Error("failed to add helm repository", zap.String("repo_name", repoName), zap.String("repo_url", repoURL), zap.Error(err))
	}
	return err
}

// pullHelmChart pulls a chart from a repository
func (h *handler) pullHelmChart(l *zap.Logger, settings *cli.EnvSettings, repoName, chartName, destDir string) (string, error) {
	hcLog := log.NewHClog(l)
	lw := hcLog.StandardWriter(&hclog.StandardLoggerOptions{})

	// Create registry client
	opts := []registry.ClientOption{
		registry.ClientOptDebug(true),
		registry.ClientOptEnableCache(false),
		registry.ClientOptWriter(lw),
		registry.ClientOptPlainHTTP(),
	}
	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return "", fmt.Errorf("failed to create registry client: %w", err)
	}

	// Create action configuration and set registry client
	cfg := &action.Configuration{
		RegistryClient: registryClient,
	}

	pull := action.NewPull(action.WithConfig(cfg))
	pull.Settings = settings
	pull.DestDir = destDir
	pull.Untar = true // Extract the chart after downloading

	// Construct the chart reference as repoName/chartName
	chartRef := fmt.Sprintf("%s/%s", repoName, chartName)
	if h.state.cfg.HelmRepoConfig.Version != "" {
		chartRef = fmt.Sprintf("%s/%s", repoName, chartName)
		pull.Version = h.state.cfg.HelmRepoConfig.Version
	}

	l.Info("pulling helm chart", zap.String("chart_ref", chartRef), zap.String("dest_dir", destDir))

	// Run the pull operation
	_, err = pull.Run(chartRef)
	if err != nil {
		return "", fmt.Errorf("failed to pull chart: %w", err)
	}

	// When Untar is true, the chart is extracted to destDir/chartName
	chartPath := filepath.Join(destDir, chartName)

	return chartPath, nil
}

// Repositories that have been permanently deleted and no longer work
var deprecatedRepos = map[string]string{
	"//kubernetes-charts.storage.googleapis.com":           "https://charts.helm.sh/stable",
	"//kubernetes-charts-incubator.storage.googleapis.com": "https://charts.helm.sh/incubator",
}

func (h *handler) addRepo(l *zap.Logger, out io.Writer, settings *cli.EnvSettings, chartDir, name, repository string) error {

	// Set default HTTP client with timeout globally for getters
	http.DefaultClient.Timeout = 30 * time.Second
	http.DefaultTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if out == nil {
		return fmt.Errorf("out parameter is nil")
	}
	if settings == nil {
		return fmt.Errorf("settings parameter is nil")
	}
	if settings.RepositoryConfig == "" {
		return fmt.Errorf("settings.RepositoryConfig is empty")
	}

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
	repoConfigDir := filepath.Dir(settings.RepositoryConfig)
	err := os.MkdirAll(repoConfigDir, os.ModePerm)
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

	lockCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		return err
	}
	if !locked {
		// Could not acquire file lock within timeout, continue anyway
	}

	b, err := os.ReadFile(settings.RepositoryConfig)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}
	c := repo.Entry{
		Name: name,
		URL:  repository,
	}

	// Check if the repo name is legal
	if strings.Contains(name, "/") {
		return errors.Errorf("repository name (%s) contains '/', please specify a different name without '/'", name)
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return err
	}

	// Add panic recovery around DownloadIndexFile
	defer func() {
		if r := recover(); r != nil {
			panic(r) // re-panic after logging
		}
	}()

	_, err = r.DownloadIndexFile()

	if err != nil {
		l.Error("helm repository index download failed", zap.String("repository", repository), zap.Error(err))
		return errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", repository)
	}

	f.Update(&c)

	if err := f.WriteFile(settings.RepositoryConfig, 0600); err != nil {
		return err
	}
	fmt.Fprintf(out, "%q has been added to your repositories\n", name)
	return nil
}

// repoNameForURL returns a stable, unique repo name per URL so distinct repos
// never collide on a name and the entry is idempotent across builds.
func repoNameForURL(url string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(url))
	return fmt.Sprintf("dep-repo-%x", h.Sum32())
}

// verifyDependenciesPresent fails if any declared dependency was not vendored,
// which would otherwise package an incomplete chart (commonly dropping CRDs).
func verifyDependenciesPresent(declared []*chartv2.Dependency, c *chartv2.Chart) error {
	loaded := map[string]struct{}{}
	for _, dep := range c.Dependencies() {
		loaded[dep.Name()] = struct{}{}
	}

	var missing []string
	for _, dep := range declared {
		if _, ok := loaded[dep.Name]; ok {
			continue
		}
		if dep.Alias != "" {
			if _, ok := loaded[dep.Alias]; ok {
				continue
			}
		}
		missing = append(missing, fmt.Sprintf("%s (repository %q)", dep.Name, dep.Repository))
	}
	if len(missing) > 0 {
		return fmt.Errorf("chart dependencies were not packaged: %s; verify the dependency repositories are reachable", strings.Join(missing, ", "))
	}
	return nil
}

func (h *handler) addDependencyReposAndUpdate(l *zap.Logger, chartDir string, repoURLs map[string]struct{}) error {
	hcLog := log.NewHClog(l)
	lw := hcLog.StandardWriter(&hclog.StandardLoggerOptions{})

	settings := cli.New()
	settings.BurstLimit = 10
	settings.QPS = 5

	// addRepo failures are non-fatal: the caller verifies dependencies were vendored.
	for url := range repoURLs {
		if err := h.addRepo(l, lw, settings, chartDir, repoNameForURL(url), url); err != nil {
			l.Warn("unable to add dependency repo", zap.String("repo_url", url), zap.Error(err))
		}
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
	if err := man.Update(); err != nil {
		return fmt.Errorf("unable to update chart dependencies: %w", err)
	}
	client.List(chartDir, lw)
	return nil
}
