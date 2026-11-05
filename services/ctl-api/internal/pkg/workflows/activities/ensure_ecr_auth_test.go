package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/runner/oci"
)

const ecrRepositoryURI = "111122223333.dkr.ecr.us-west-2.amazonaws.com/org1/app1"

func rewrittenCfg(loginServer string) *configs.OCIRegistryRepository {
	return &configs.OCIRegistryRepository{
		Plugin:       "oci",
		RegistryType: configs.OCIRegistryTypePrivateOCI,
		Repository:   ecrRepositoryURI,
		LoginServer:  loginServer,
		OCIAuth:      &configs.OCIRegistryAuth{Username: "AWS", Password: "token"},
	}
}

func TestScopeECRPullLimitsTheAssumeToTheRepository(t *testing.T) {
	cfg := ecrSandboxCfg()
	require.NoError(t, scopeECRPull(cfg))
	require.Contains(t, cfg.ECRAuth.AssumeRole.SessionPolicy, "arn:aws:ecr:us-west-2:111122223333:repository/org1/app1")
	require.NotContains(t, cfg.ECRAuth.AssumeRole.SessionPolicy, "ecr:PutImage")
}

func TestScopeECRPullRejectsANonECRRepository(t *testing.T) {
	cfg := ecrSandboxCfg()
	cfg.Repository = "registry.example.com/org/app"
	require.Error(t, scopeECRPull(cfg))
}

func TestRewrittenECRConfigKeepsRepositoryURI(t *testing.T) {
	info, err := oci.FetchAccessInfo(context.Background(), rewrittenCfg("111122223333.dkr.ecr.us-west-2.amazonaws.com"))
	require.NoError(t, err)
	require.Equal(t, ecrRepositoryURI, info.RepositoryURI())
	require.Equal(t, "AWS", info.Auth.Username)
}

func TestECRLoginServerWithSchemeDoublesTheHost(t *testing.T) {
	info, err := oci.FetchAccessInfo(context.Background(), rewrittenCfg("https://111122223333.dkr.ecr.us-west-2.amazonaws.com"))
	require.NoError(t, err)
	require.NotEqual(t, ecrRepositoryURI, info.RepositoryURI(), "ProxyEndpoint carries a scheme, so it has to be trimmed")
}
