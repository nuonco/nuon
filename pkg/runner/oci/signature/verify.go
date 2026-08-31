package signature

import (
	"context"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/cosign/v3/pkg/cosign"
	cosignremote "github.com/sigstore/cosign/v3/pkg/oci/remote"
	cosignsignature "github.com/sigstore/cosign/v3/pkg/signature"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/verify"

	signaturecfg "github.com/nuonco/nuon/pkg/oci/signature"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	runneroci "github.com/nuonco/nuon/pkg/runner/oci"
)

func Verify(ctx context.Context, cfg *configs.OCIRegistryRepository, digest string, verification *signaturecfg.Verification) error {
	if verification == nil || !verification.RequireSignature {
		return nil
	}

	accessInfo, err := runneroci.FetchAccessInfo(ctx, cfg)
	if err != nil {
		return fmt.Errorf("fetch registry credentials: %w", err)
	}

	nameOpts := []name.Option{name.WeakValidation}
	if accessInfo.Insecure {
		nameOpts = append(nameOpts, name.Insecure)
	}
	ref, err := name.ParseReference(accessInfo.RepositoryURI()+"@"+digest, nameOpts...)
	if err != nil {
		return fmt.Errorf("parse digest reference: %w", err)
	}

	remoteOpts := []remote.Option{remote.WithAuth(authn.Anonymous)}
	if accessInfo.Auth != nil && accessInfo.Auth.Username != "" {
		remoteOpts = []remote.Option{remote.WithAuth(authn.FromConfig(authn.AuthConfig{
			Username: accessInfo.Auth.Username,
			Password: accessInfo.Auth.Password,
		}))}
	}
	registryOpts := []cosignremote.Option{cosignremote.WithRemoteOptions(remoteOpts...)}

	var trustedRoot root.TrustedMaterial
	for _, authority := range verification.Authorities {
		if authority.Type == signaturecfg.AuthorityTypeSigstoreKeyless {
			trustedRoot, err = cosign.TrustedRoot()
			if err != nil {
				return fmt.Errorf("load Sigstore trusted root: %w", err)
			}
			break
		}
	}

	var authorityErrors []error
	for i, authority := range verification.Authorities {
		if err := verifyAuthority(ctx, ref, nameOpts, registryOpts, trustedRoot, authority); err == nil {
			return nil
		} else {
			authorityErrors = append(authorityErrors, fmt.Errorf("authority %d (%s): %w", i, authority.Type, err))
		}
	}

	return fmt.Errorf("image signature did not satisfy any configured authority: %w", errors.Join(authorityErrors...))
}

func verifyAuthority(ctx context.Context, ref name.Reference, nameOpts []name.Option, registryOpts []cosignremote.Option, trustedRoot root.TrustedMaterial, authority signaturecfg.Authority) error {
	checkOpts := &cosign.CheckOpts{
		RegistryClientOpts: registryOpts,
		ClaimVerifier:      cosign.SimpleClaimVerifier,
		TrustedMaterial:    trustedRoot,
		ExperimentalOCI11:  true,
	}

	switch authority.Type {
	case signaturecfg.AuthorityTypeSigstoreKeyless:
		checkOpts.Identities = []cosign.Identity{{
			Issuer:        authority.Issuer,
			Subject:       authority.Subject,
			SubjectRegExp: authority.SubjectRegexp,
		}}
	case signaturecfg.AuthorityTypeCosignKey:
		verifier, err := cosignsignature.LoadPublicKeyRaw([]byte(authority.PublicKey), crypto.SHA256)
		if err != nil {
			return fmt.Errorf("load public key: %w", err)
		}
		checkOpts.SigVerifier = verifier
		checkOpts.IgnoreTlog = true
	default:
		return fmt.Errorf("unsupported authority type %q", authority.Type)
	}

	if err := verifyBundleSignatures(ctx, ref, nameOpts, checkOpts); err == nil {
		return nil
	} else {
		bundleErr := err
		if _, _, err := cosign.VerifyImageSignatures(ctx, ref, checkOpts); err == nil {
			return nil
		} else {
			return errors.Join(fmt.Errorf("verify OCI bundle: %w", bundleErr), fmt.Errorf("verify Cosign signature: %w", err))
		}
	}
}

func verifyBundleSignatures(ctx context.Context, ref name.Reference, nameOpts []name.Option, checkOpts *cosign.CheckOpts) error {
	bundles, hash, err := cosign.GetBundles(ctx, ref, checkOpts.RegistryClientOpts, nameOpts...)
	if err != nil {
		return err
	}
	digest, err := hex.DecodeString(hash.Hex)
	if err != nil {
		return fmt.Errorf("decode image digest: %w", err)
	}

	var bundleErrors []error
	for _, bundle := range bundles {
		isImageSignature := bundle.GetMessageSignature() != nil
		if envelope := bundle.GetDsseEnvelope(); envelope != nil {
			statement := struct {
				PredicateType string `json:"predicateType"`
			}{}
			if err := json.Unmarshal(envelope.Payload, &statement); err != nil {
				bundleErrors = append(bundleErrors, fmt.Errorf("decode signature statement: %w", err))
				continue
			}
			isImageSignature = statement.PredicateType == "https://sigstore.dev/cosign/sign/v1"
		}
		if !isImageSignature {
			continue
		}
		if _, err := cosign.VerifyNewBundle(ctx, checkOpts, verify.WithArtifactDigest(hash.Algorithm, digest), bundle); err == nil {
			return nil
		} else {
			bundleErrors = append(bundleErrors, err)
		}
	}
	if len(bundleErrors) == 0 {
		return errors.New("no image signature bundles found")
	}
	return errors.Join(bundleErrors...)
}
