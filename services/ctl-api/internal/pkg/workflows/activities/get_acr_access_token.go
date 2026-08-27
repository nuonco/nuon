package activities

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/azure/acr"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GetACRAccessTokenRequest names the credential rather than carrying it.
//
// ClientSecretName and ClientCertificateName are AppSecret names, resolved
// against the DB inside the activity. Activity inputs are recorded in Temporal
// history, so passing the vendor's long-lived secret here would persist it;
// passing a name persists nothing. The minted refresh token does land in
// history as the activity result, but it expires in 60 minutes — the same
// exposure GAR tokens already have.
type GetACRAccessTokenRequest struct {
	// ComponentID rather than an app ID: it is present on every build plan
	// without depending on a preload, and the activity has a DB handle anyway.
	ComponentID string
	LoginServer string
	TenantID    string
	ClientID    string

	ClientSecretName      string
	ClientCertificateName string
}

type ACRAccessToken struct {
	Username string
	Password string
}

// GetACRAccessToken mints a registry refresh token for a vendor-owned ACR.
//
// Like GetGARAccessToken this lives in the shared activity set because both the
// components namespace (pulling a source image for a build) and the installs
// namespace (resolving a sandbox artifact) schedule it.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) GetACRAccessToken(ctx context.Context, req *GetACRAccessTokenRequest) (*ACRAccessToken, error) {
	l := temporalzap.GetActivityLogger(ctx)

	cfg := &azurecredentials.Config{
		TenantID: req.TenantID,
		ClientID: req.ClientID,
	}

	switch {
	case req.ClientSecretName != "" && req.ClientCertificateName != "":
		return nil, fmt.Errorf("azure_acr sets both client_secret and client_certificate; set exactly one")

	case req.ClientSecretName != "":
		secret, err := a.appSecretValue(ctx, req.ComponentID, req.ClientSecretName)
		if err != nil {
			return nil, err
		}
		cfg.ClientSecret = secret

	case req.ClientCertificateName != "":
		raw, err := a.appSecretValue(ctx, req.ComponentID, req.ClientCertificateName)
		if err != nil {
			return nil, err
		}

		// Certificates are stored base64-encoded (AppSecretConfigFmtBase64), but
		// accept a bare PEM too rather than failing on a reasonable mistake.
		//
		// Whitespace is stripped before decoding because base64 is routinely
		// line-wrapped and DecodeString rejects embedded newlines; without this
		// a wrapped certificate falls through to the raw branch and fails with
		// a misleading "not a PEM" error. A bare PEM cannot decode as base64
		// (its "-----" delimiters are outside the alphabet), so the fallback
		// stays unambiguous.
		pem := decodeCertificate(raw)
		if !strings.Contains(string(pem), "-----BEGIN") {
			return nil, errors.New(
				"the app secret named by azure_acr.client_certificate_name is not a PEM certificate; " +
					"expected a base64-encoded PEM",
			)
		}
		cfg.ClientCertificatePEM = pem
	}

	// No app registration means the registry is expected to be reachable by
	// whatever ambient identity this process has — the same-tenant case that
	// worked before any of this existed. Partial config is a different thing
	// and worth rejecting, because silently falling back would surface as an
	// unexplained 401 against a registry the caller believes it configured.
	if !cfg.HasAppRegistrationCredentials() {
		if cfg.ClientID != "" || cfg.TenantID != "" ||
			req.ClientSecretName != "" || req.ClientCertificateName != "" {
			return nil, fmt.Errorf(
				"azure_acr for registry %q is partially configured: set tenant_id, client_id, and exactly one of client_secret or client_certificate, or none of them to use ambient credentials",
				req.LoginServer,
			)
		}

		cfg = nil
	}

	l.Info("minting acr refresh token",
		zap.String("registry", req.LoginServer),
		zap.String("identity", identityDescription(cfg)),
	)

	token, err := acr.GetRepositoryToken(ctx, cfg, req.LoginServer, l)
	if err != nil {
		return nil, fmt.Errorf("unable to get acr token for %s as %s: %w", req.LoginServer, identityDescription(cfg), err)
	}

	return &ACRAccessToken{
		Username: acr.DefaultACRUsername,
		Password: token,
	}, nil
}

// appSecretValue resolves an AppSecret by name within the component's app. A
// missing secret is a config error the vendor can act on, so it is reported as
// such rather than surfacing later as an opaque registry 401.
func (a *Activities) appSecretValue(ctx context.Context, componentID, name string) (string, error) {
	var component app.Component
	if err := a.db.WithContext(ctx).Where(app.Component{
		ID: componentID,
	}).First(&component).Error; err != nil {
		return "", fmt.Errorf("unable to resolve component %s for azure_acr credentials: %w", componentID, err)
	}

	var secret app.AppSecret
	if err := a.db.WithContext(ctx).Where(app.AppSecret{
		AppID: component.AppID,
		Name:  name,
	}).First(&secret).Error; err != nil {
		// The name is not echoed back. The overwhelmingly likely cause of a
		// miss is that someone put the credential itself in the config field
		// instead of the name of an app secret, and repeating it here would
		// copy it into logs and workflow history — the exact thing the
		// name-only design exists to avoid.
		return "", fmt.Errorf(
			"unable to resolve the app secret named by azure_acr for app %s: %w; "+
				"this field takes the name of an app secret, not the credential itself "+
				"(create one with: nuon apps variables create --name <name> --value <credential>)",
			component.AppID, err,
		)
	}

	if secret.Value == "" {
		return "", fmt.Errorf("the app secret named by azure_acr for app %s is empty", component.AppID)
	}

	return secret.Value, nil
}

func identityDescription(cfg *azurecredentials.Config) string {
	if cfg == nil {
		return "ambient credentials"
	}
	return cfg.String()
}

func stripWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, s)
}

// decodeCertificate returns the PEM bytes for a stored certificate, accepting a
// bare PEM or any of the base64 flavours a vendor might reasonably produce.
//
// Whitespace is stripped first because base64 is routinely line-wrapped and the
// decoders reject embedded newlines; without that a wrapped certificate would
// fall through to the raw branch and fail with a misleading "not a PEM" error.
// Padded and unpadded, standard and URL alphabets are all tried, so the failure
// message only ever means "this is not a certificate" rather than "this is the
// wrong flavour of base64".
//
// A bare PEM cannot decode as base64 — its "-----" delimiters are outside every
// alphabet — so falling back to the raw bytes stays unambiguous.
func decodeCertificate(raw string) []byte {
	compact := stripWhitespace(raw)

	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		if decoded, err := enc.DecodeString(compact); err == nil {
			return decoded
		}
	}

	return []byte(raw)
}
