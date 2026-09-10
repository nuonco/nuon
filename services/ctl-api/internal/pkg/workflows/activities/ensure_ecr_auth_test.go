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
