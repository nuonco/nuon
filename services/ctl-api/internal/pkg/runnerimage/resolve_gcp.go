package runnerimage

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/oauth2/google"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// ResolveGCPImageDigest resolves the manifest digest for an `imageURL:tag` pair
// in Google Artifact Registry. Uses ambient Application Default Credentials,
// which work for ctl-api running on GCP with workload identity. Returns digests
// in the `sha256:...` form.
func ResolveGCPImageDigest(ctx context.Context, imageURL, tag string) (string, error) {
	if !isGARURL(imageURL) {
		return "", fmt.Errorf("not a GCP Artifact Registry url: %s", imageURL)
	}

	ts, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("unable to get GCP token source: %w", err)
	}
	token, err := ts.Token()
	if err != nil {
		return "", fmt.Errorf("unable to get GCP access token: %w", err)
	}

	repo, err := remote.NewRepository(imageURL)
	if err != nil {
		return "", fmt.Errorf("unable to create repository client: %w", err)
	}
	host := strings.SplitN(imageURL, "/", 2)[0]
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(host, auth.Credential{
			Username: "oauth2accesstoken",
			Password: token.AccessToken,
		}),
	}

	desc, err := repo.Resolve(ctx, tag)
	if err != nil {
		return "", fmt.Errorf("unable to resolve %s:%s: %w", imageURL, tag, err)
	}
	digest := desc.Digest.String()
	if digest == "" {
		return "", fmt.Errorf("empty digest for %s:%s", imageURL, tag)
	}
	return digest, nil
}

func isGARURL(imageURL string) bool {
	imageURL = strings.TrimSpace(imageURL)
	return strings.Contains(imageURL, "-docker.pkg.dev/") || strings.HasPrefix(imageURL, "gcr.io/")
}
