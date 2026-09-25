package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/distribution/reference"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createContainerImageBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.ContainerImagePullPlan, error) {
	srcRepo, err := p.getSourceRepository(
		ctx,
		bld.ComponentConfigConnection.ExternalImageComponentConfig,
		bld.ComponentConfigConnection.ComponentID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get source repository")
	}

	plan := &plantypes.ContainerImagePullPlan{
		Image:        bld.ComponentConfigConnection.ExternalImageComponentConfig.ImageURL,
		Tag:          bld.ComponentConfigConnection.ExternalImageComponentConfig.Tag,
		UpdatePolicy: bld.ComponentConfigConnection.ExternalImageComponentConfig.UpdatePolicy,
		Verification: bld.ComponentConfigConnection.ExternalImageComponentConfig.Verification,

		RepoCfg: srcRepo,
	}

	// Look up the most recent prior Active build's SourceDigest so the runner
	// can detect a no-op (upstream digest unchanged) and skip the artifact
	// push. Failure here is non-fatal — without the hint the runner just runs
	// a normal copy.
	prior, err := activities.AwaitGetPreviousActiveBuildSourceDigest(ctx, activities.GetPreviousActiveBuildSourceDigestRequest{
		ComponentID:    bld.ComponentConfigConnection.ComponentID,
		ExcludeBuildID: bld.ID,
	})
	if err != nil {
		if l, lerr := log.WorkflowLogger(ctx); lerr == nil {
			l.Warn("unable to look up previous source digest for build dedup",
				zap.String("component_id", bld.ComponentConfigConnection.ComponentID),
				zap.String("build_id", bld.ID),
				zap.Error(err))
		}
	} else if prior != nil {
		plan.PreviousSourceDigest = prior.SourceDigest
	}

	return plan, nil
}

func (b *Planner) normalizeRepository(repo string) (string, error) {
	ref, err := reference.ParseAnyReference(repo)
	if err != nil {
		return "", fmt.Errorf("invalid reference: %w", err)
	}

	named, err := reference.ParseDockerRef(ref.String())
	if err != nil {
		return "", fmt.Errorf("unable to parse docker ref: %w", err)
	}

	host := reference.Domain(named)
	if host == "docker.io" {
		// The normalized name parse above will turn short names like "foo/bar"
		// into "docker.io/foo/bar". We return "docker.io" and let oras-go
		// handle the mapping to "registry-1.docker.io" internally.
		// Using "index.docker.io" breaks the anonymous bearer token flow.
		return "docker.io", nil
	}

	// by default, if a reference is fully resolved, we just use the repository name
	return "", nil
}

func (b *Planner) getSourceRepository(ctx workflow.Context, cfg *app.ExternalImageComponentConfig, componentID string) (*configs.OCIRegistryRepository, error) {
	loginServer, err := b.normalizeRepository(cfg.ImageURL)
	if err != nil {
		return nil, errors.Wrap(err, "unable to normalize repository")
	}

	if cfg.AWSECRImageConfig != nil {
		var auth credentials.Config
		if err := workflow.ExecuteActivity(ctx, "GetCloudConnectionCredentials", &activities.GetCloudConnectionCredentialsRequest{ConnectionID: cfg.AWSECRImageConfig.CloudConnectionID, SessionName: "container-image-build"}).Get(ctx, &auth); err != nil {
			return nil, fmt.Errorf("get cloud connection credentials: %w", err)
		}

		return &configs.OCIRegistryRepository{
			RegistryType: configs.OCIRegistryTypeECR,
			Repository:   cfg.ImageURL,
			Region:       cfg.AWSECRImageConfig.AWSRegion,

			ECRAuth: &auth,
		}, nil
	}

	if cfg.GCPGARImageConfig != nil {
		garLoginServer := fmt.Sprintf("%s-docker.pkg.dev", cfg.GCPGARImageConfig.GCPRegion)
		repository := &configs.OCIRegistryRepository{
			RegistryType:             configs.OCIRegistryTypeGAR,
			Repository:               cfg.ImageURL,
			Region:                   cfg.GCPGARImageConfig.GCPRegion,
			LoginServer:              garLoginServer,
			ServiceAccountEmail:      cfg.GCPGARImageConfig.ServiceAccountEmail,
			WorkloadIdentityProvider: cfg.GCPGARImageConfig.WorkloadIdentityProvider,
			OrgID:                    cfg.GCPGARImageConfig.OrgID,
			CloudConnectionID:        cfg.GCPGARImageConfig.CloudConnectionID,
		}
		if connection := cfg.GCPGARImageConfig.CloudConnection; connection != nil && connection.AuthMode != app.CloudConnectionAuthModeLegacy {
			auth, err := activities.AwaitGetCloudConnectionGARAuth(ctx, &activities.GetCloudConnectionGARAuthRequest{ConnectionID: connection.ID})
			if err != nil {
				return nil, fmt.Errorf("get GCP cloud connection credentials: %w", err)
			}
			repository.OCIAuth = auth
		}
		return repository, nil
	}

	if cfg.AzureACRImageConfig != nil {
		acrCfg := &configs.OCIRegistryRepository{
			RegistryType: configs.OCIRegistryTypeACR,
			Repository:   cfg.ImageURL,
			LoginServer:  cfg.AzureACRImageConfig.RegistryURL,
			ACRAuth: &azurecredentials.Config{
				UseDefault: true,
			},
		}

		// Naming an app registration is what makes a registry in someone else's
		// tenant reachable; without one the ambient identity is all there is,
		// which only works when the registry shares our tenant.
		//
		// Any field being set is enough to attach it. A half-specified
		// registration is rejected at sync, but attaching it here too means a
		// config that slipped through fails loudly in the token activity rather
		// than quietly falling back to an identity that cannot see the registry.
		if acr := cfg.AzureACRImageConfig; acr.CloudConnectionID != "" || acr.ClientID != "" || acr.TenantID != "" ||
			acr.ClientSecretName != "" || acr.ClientCertificateName != "" {
			acrCfg.ACRAppRegistration = &configs.ACRAppRegistration{
				ComponentID:           componentID,
				ConnectionID:          acr.CloudConnectionID,
				TenantID:              acr.TenantID,
				ClientID:              acr.ClientID,
				ClientSecretName:      acr.ClientSecretName,
				ClientCertificateName: acr.ClientCertificateName,
			}
		}

		return acrCfg, nil
	}

	return &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypePublicOCI,

		Repository:  cfg.ImageURL,
		LoginServer: loginServer,

		OCIAuth: &configs.OCIRegistryAuth{},
	}, nil
}
