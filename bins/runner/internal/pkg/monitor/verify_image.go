package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	signaturecfg "github.com/nuonco/nuon/pkg/oci/signature"
	"github.com/nuonco/nuon/pkg/runner/oci/signature"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

const (
	ImageVerificationDisabled = "disabled"
	ImageVerificationWarn     = "warn"
	ImageVerificationEnforce  = "enforce"
)

type imageConfig struct {
	ContainerImageURL string
	ContainerImageTag string
}

// failedVerificationTTL spaces out retries of a digest that failed, so an unsigned image does not
// send every monitor loop to the registry and Sigstore.
const failedVerificationTTL = time.Hour

type verificationResult struct {
	err error
	at  time.Time
}

// verifiedImages caches verification results by digest reference and policy, so the monitor loop
// only reaches the registry and Sigstore when the tag moves.
var verifiedImages sync.Map

type verifyImageFn func(ctx context.Context, s *settings.Settings) (string, error)

// runnerImageConfig pins the runner image to the digest its tag resolves to once that digest's
// signature verifies. Enforce mode returns an error instead of an unverified image, so the image
// config the service runs from is left as it was.
func runnerImageConfig(ctx context.Context, l *zap.Logger, s *settings.Settings, verify verifyImageFn) (imageConfig, error) {
	cfg := imageConfig{ContainerImageURL: s.ContainerImageURL, ContainerImageTag: s.ContainerImageTag}
	switch s.ContainerImageVerificationMode {
	case "", ImageVerificationDisabled:
		return cfg, nil
	}

	digest, err := verify(ctx, s)
	if err == nil {
		cfg.ContainerImageTag = s.ContainerImageTag + "@" + digest
		return cfg, nil
	}
	if s.ContainerImageVerificationMode == ImageVerificationEnforce {
		return imageConfig{}, errors.Wrapf(err, "runner image %s:%s failed signature verification", s.ContainerImageURL, s.ContainerImageTag)
	}
	l.Warn("runner image failed signature verification",
		zap.String("image", s.ContainerImageURL+":"+s.ContainerImageTag),
		zap.Error(err))
	return cfg, nil
}

func verifyRunnerImage(ctx context.Context, s *settings.Settings) (string, error) {
	if s.ContainerImageURL == "" || s.ContainerImageTag == "" {
		return "", errors.New("runner image url or tag is not set")
	}
	if s.ContainerImageSignatureIssuer == "" || s.ContainerImageSignatureIdentityRegexp == "" {
		return "", errors.New("runner image signature issuer or identity is not set")
	}

	nameOpts := []name.Option{name.WeakValidation}
	ref, err := name.ParseReference(s.ContainerImageURL+":"+s.ContainerImageTag, nameOpts...)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse runner image reference")
	}
	remoteOpts := []remote.Option{remote.WithContext(ctx), remote.WithAuthFromKeychain(authn.DefaultKeychain)}
	desc, err := remote.Head(ref, remoteOpts...)
	if err != nil {
		return "", errors.Wrap(err, "unable to resolve runner image digest")
	}
	digest := desc.Digest.String()
	digestRef := ref.Context().Digest(digest)
	key := digestRef.String() + "|" + s.ContainerImageSignatureIssuer + "|" + s.ContainerImageSignatureIdentityRegexp
	if cached, ok := verifiedImages.Load(key); ok {
		result := cached.(verificationResult)
		if result.err == nil {
			return digest, nil
		}
		if time.Since(result.at) < failedVerificationTTL {
			return "", result.err
		}
	}

	err = signature.VerifyReference(ctx, digestRef, nameOpts, remoteOpts, &signaturecfg.Verification{
		RequireSignature: true,
		Authorities: []signaturecfg.Authority{{
			Type:          signaturecfg.AuthorityTypeKeyless,
			Issuer:        s.ContainerImageSignatureIssuer,
			SubjectRegexp: s.ContainerImageSignatureIdentityRegexp,
		}},
	})
	verifiedImages.Store(key, verificationResult{err: err, at: time.Now()})
	if err != nil {
		return "", err
	}
	return digest, nil
}
