package monitor

import (
	"context"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/settings"
)

func TestRunnerImageConfig(t *testing.T) {
	const digest = "sha256:3f9a0c1e6b2d4a7f8c5e9b1d0a2c4e6f8b0d2f4a6c8e0b2d4f6a8c0e2b4d6f8a"
	verified := func(context.Context, *settings.Settings) (string, error) { return digest, nil }
	rejected := func(context.Context, *settings.Settings) (string, error) {
		return "", errors.New("no matching signatures")
	}
	unexpected := func(context.Context, *settings.Settings) (string, error) {
		t.Fatal("verify must not run when verification is disabled")
		return "", nil
	}

	tests := []struct {
		name    string
		mode    string
		verify  verifyImageFn
		wantTag string
		wantErr bool
	}{
		{name: "unset keeps the tag", mode: "", verify: unexpected, wantTag: "0.19.1200"},
		{name: "disabled keeps the tag", mode: ImageVerificationDisabled, verify: unexpected, wantTag: "0.19.1200"},
		{name: "warn pins a verified digest", mode: ImageVerificationWarn, verify: verified, wantTag: "0.19.1200@" + digest},
		{name: "warn falls back to the tag", mode: ImageVerificationWarn, verify: rejected, wantTag: "0.19.1200"},
		{name: "enforce pins a verified digest", mode: ImageVerificationEnforce, verify: verified, wantTag: "0.19.1200@" + digest},
		{name: "enforce refuses an unverified image", mode: ImageVerificationEnforce, verify: rejected, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &settings.Settings{
				ContainerImageURL:              "public.ecr.aws/p7e3r5y0/runner",
				ContainerImageTag:              "0.19.1200",
				ContainerImageVerificationMode: tc.mode,
			}
			got, err := runnerImageConfig(context.Background(), zap.NewNop(), s, tc.verify)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "public.ecr.aws/p7e3r5y0/runner", got.ContainerImageURL)
			require.Equal(t, tc.wantTag, got.ContainerImageTag)
		})
	}
}

func TestVerifyRunnerImageRequiresIdentity(t *testing.T) {
	_, err := verifyRunnerImage(context.Background(), &settings.Settings{
		ContainerImageURL: "public.ecr.aws/p7e3r5y0/runner",
		ContainerImageTag: "main",
	})
	require.ErrorContains(t, err, "issuer or identity")
}

func TestResolveRunnerImageDigestFallsBackToAnonymous(t *testing.T) {
	srv := httptest.NewServer(registry.New())
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")

	ref, err := name.ParseReference(host+"/runner:main", name.WeakValidation)
	require.NoError(t, err)
	img, err := random.Image(64, 1)
	require.NoError(t, err)
	require.NoError(t, remote.Write(ref, img))
	want, err := img.Digest()
	require.NoError(t, err)

	dockerConfig := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dockerConfig, "config.json"),
		[]byte(`{"credHelpers":{"`+host+`":"nuon-missing-helper"}}`), 0o600))
	t.Setenv("DOCKER_CONFIG", dockerConfig)

	_, err = remote.Head(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	require.Error(t, err, "a broken credential helper should fail the keychain lookup")

	got, opts, err := resolveRunnerImageDigest(context.Background(), ref, &settings.Settings{})
	require.NoError(t, err)
	require.Equal(t, want.String(), got)
	require.NotEmpty(t, opts)
}
