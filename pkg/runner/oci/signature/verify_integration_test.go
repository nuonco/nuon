package signature

import (
	"context"
	"os"
	"testing"

	signaturecfg "github.com/nuonco/nuon/pkg/oci/signature"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

func TestVerifyPublicKeylessImage(t *testing.T) {
	if os.Getenv("NUON_INTEGRATION") == "" {
		t.Skip("set NUON_INTEGRATION to run registry verification")
	}

	cfg := &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypePublicOCI,
		Repository:   "cgr.dev/chainguard/static",
		OCIAuth:      &configs.OCIRegistryAuth{},
	}
	verification := &signaturecfg.Verification{
		RequireSignature: true,
		Authorities: []signaturecfg.Authority{{
			Type:    signaturecfg.AuthorityTypeKeyless,
			Issuer:  "https://token.actions.githubusercontent.com",
			Subject: "https://github.com/chainguard-images/images/.github/workflows/release.yaml@refs/heads/main",
		}},
	}

	err := Verify(context.Background(), cfg, "sha256:96d02f455d5a73b817c0602910748609cf8471b1cc9522f78c75cedb1f67d072", verification)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}
