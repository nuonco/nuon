package builder

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/pusher"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/uploader"
)

func (b *Builder) packageChart(log hclog.Logger) (string, error) {
	chart, err := loader.Load(b.chartDir)
	if err != nil {
		return "", fmt.Errorf("unable to load chart: %w", err)
	}
	log.Info("succesfully loaded chart")

	packagePath, err := chartutil.Save(chart, b.tmpDir)
	if err != nil {
		return "", fmt.Errorf("unable to package chart: %w", err)
	}
	log.Info("succesfully packaged chart")

	return packagePath, nil
}

func (b *Builder) pushChart(log hclog.Logger, packagePath string, accessInfo *ociv1.AccessInfo) error {
	ociPath := fmt.Sprintf("oci://%s", accessInfo.Auth.ServerAddress)

	registryClient, err := registry.NewClient()
	if err != nil {
		return fmt.Errorf("unable to get registry client: %w", err)
	}
	if err := registryClient.Login(accessInfo.Auth.ServerAddress,
		registry.LoginOptBasicAuth(accessInfo.Auth.Username, accessInfo.Auth.Password),
		registry.LoginOptInsecure(false)); err != nil {
		return fmt.Errorf("unable to login to registry: %w", err)
	}

	u := uploader.ChartUploader{
		Out: log.StandardWriter(&hclog.StandardLoggerOptions{}),
		Pushers: []pusher.Provider{
			{
				Schemes: []string{registry.OCIScheme},
				New:     pusher.NewOCIPusher,
			},
		},

		// NOTE(jm): do we need to add a registryClient on both of these?
		Options: []pusher.Option{
			pusher.WithRegistryClient(registryClient),
		},
		RegistryClient: registryClient,
	}

	// NOTE: the package path has to be equivalent to <install-id>-<version> as that is how helm unfurls the tag
	if err := u.UploadTo(packagePath, ociPath); err != nil {
		return fmt.Errorf("unable to upload chart: %w", err)
	}

	return nil
}
