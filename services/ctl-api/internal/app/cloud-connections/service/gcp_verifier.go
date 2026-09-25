package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnectionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type gcpVerifier struct {
	issuer *oidcissuer.Issuer
	client *http.Client
}

type gcpNegativeProbeStage string

const (
	gcpNegativeProbeDeniedAtSTS           gcpNegativeProbeStage = "sts"
	gcpNegativeProbeDeniedAtImpersonation gcpNegativeProbeStage = "impersonation"
	gcpNegativeProbeSucceeded             gcpNegativeProbeStage = "succeeded"
)

func (v *gcpVerifier) Verify(ctx context.Context, connection *app.CloudConnection, options VerifyOptions) (VerificationResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	federated, err := cloudconnectionshelpers.GCPFederatedToken(ctx, v.issuer, connection, subject(connection))
	if err != nil {
		return verificationFailure("Nuon OIDC identity is not trusted by this Workload Identity Provider."), nil
	}
	accessToken, err := cloudconnectionshelpers.GCPImpersonatedToken(ctx, v.client, federated.AccessToken, connection.Principal)
	if err != nil {
		return verificationFailure("The Nuon identity cannot impersonate the configured service account."), nil
	}
	if err := v.verifyIdentity(ctx, accessToken.AccessToken, connection.Principal); err != nil {
		return verificationFailure("The exchanged identity does not match the configured service account."), nil
	}
	stage, err := v.negativeProbe(ctx, connection)
	if err != nil {
		return VerificationResult{}, err
	}
	if err := evaluateGCPNegativeProbe(stage); err != nil {
		return verificationFailure(err.Error()), nil
	}

	capabilities := []app.CloudConnectionCapability{}
	if connection.HasCapability(app.CloudConnectionCapabilityImages) && v.probeRepositories(ctx, accessToken.AccessToken, connection.TargetID, options.Repositories) == nil {
		capabilities = append(capabilities, app.CloudConnectionCapabilityImages)
	}
	if len(capabilities) == 0 {
		return VerificationResult{Status: app.CloudConnectionStatusError, Message: "The service account lacks access to the requested Artifact Registry repositories.", Capabilities: capabilities}, nil
	}
	return VerificationResult{Status: app.CloudConnectionStatusVerified, Message: "Cloud connection verified.", Capabilities: capabilities}, nil
}

func (v *gcpVerifier) negativeProbe(ctx context.Context, connection *app.CloudConnection) (gcpNegativeProbeStage, error) {
	federated, err := cloudconnectionshelpers.GCPFederatedToken(ctx, v.issuer, connection, "org:foreign:connection:foreign")
	if err != nil {
		if cloudconnectionshelpers.GCPAccessDenied(err) {
			return gcpNegativeProbeDeniedAtSTS, nil
		}
		return "", fmt.Errorf("probe foreign subject at STS: %w", err)
	}
	if _, err := cloudconnectionshelpers.GCPImpersonatedToken(ctx, v.client, federated.AccessToken, connection.Principal); err != nil {
		if cloudconnectionshelpers.GCPAccessDenied(err) {
			return gcpNegativeProbeDeniedAtImpersonation, nil
		}
		return "", fmt.Errorf("probe foreign subject at service account impersonation: %w", err)
	}
	return gcpNegativeProbeSucceeded, nil
}

func evaluateGCPNegativeProbe(stage gcpNegativeProbeStage) error {
	if stage == gcpNegativeProbeDeniedAtSTS || stage == gcpNegativeProbeDeniedAtImpersonation {
		return nil
	}
	return fmt.Errorf("the service account trust policy accepts a foreign Nuon connection subject")
}

func (v *gcpVerifier) verifyIdentity(ctx context.Context, token, email string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://oauth2.googleapis.com/tokeninfo?access_token="+url.QueryEscape(token), nil)
	if err != nil {
		return err
	}
	response, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("token info returned %s", response.Status)
	}
	var info struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(response.Body).Decode(&info); err != nil {
		return err
	}
	if !strings.EqualFold(info.Email, email) {
		return fmt.Errorf("token email %q does not match %q", info.Email, email)
	}
	return nil
}

func (v *gcpVerifier) probeRepositories(ctx context.Context, token, projectID string, repositories []string) error {
	paths := []string{"projects/" + projectID + "/locations/-/repositories?pageSize=1"}
	if len(repositories) > 0 {
		paths = paths[:0]
		for _, repository := range repositories {
			parts := strings.SplitN(strings.Trim(repository, "/"), "/", 2)
			if len(parts) != 2 {
				return fmt.Errorf("repository %q must use location/name format", repository)
			}
			paths = append(paths, "projects/"+projectID+"/locations/"+parts[0]+"/repositories/"+parts[1])
		}
	}
	for _, path := range paths {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://artifactregistry.googleapis.com/v1/"+path, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		response, err := v.client.Do(req)
		if err != nil {
			return err
		}
		_, _ = io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("Artifact Registry probe returned %s", response.Status)
		}
	}
	return nil
}
