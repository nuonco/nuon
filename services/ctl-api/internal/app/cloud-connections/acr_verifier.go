package cloudconnections

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/nuonco/nuon/pkg/azure/acr"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var acrForeignSubjectDeniedPattern = regexp.MustCompile(`\bAADSTS(?:70021|700213)\b`)
var acrIdentityDeniedPattern = regexp.MustCompile(`\bAADSTS(?:70021|700213|700016|70025|700024|500011|50020)\b`)

type acrVerificationError struct {
	message string
	cause   error
}

func (e *acrVerificationError) Error() string {
	return e.message + " " + VerificationErrorMessage(e.cause)
}
func (e *acrVerificationError) Unwrap() error { return e.cause }

func (v *providerVerifier) verifyACR(ctx context.Context, connection *app.CloudConnection, options VerifyOptions) (result VerificationResult, verifyErr error) {
	message := "Azure federation could not be checked. Review the existing application and service principal, exact issuer/subject/audience, and publicly reachable issuer discovery and JWKS in step 2, then retry."
	defer func() {
		if verifyErr != nil {
			verifyErr = &acrVerificationError{message: message, cause: verifyErr}
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	credential, err := ACRCredential(v.issuer, connection, Subject(connection))
	if err != nil {
		return VerificationResult{}, err
	}
	tokenOptions := policy.TokenRequestOptions{Scopes: []string{acr.RegistryScope}}
	deadline := v.now().Add(time.Minute)
	for {
		_, err = credential.GetToken(ctx, tokenOptions)
		if err == nil {
			break
		}
		if !acrEntraDenied(err, acrIdentityDeniedPattern) {
			return VerificationResult{}, fmt.Errorf("authenticate Azure connection: %w", err)
		}
		if !options.RetryIAMPropagation || !v.now().Before(deadline) {
			return verificationFailure("Microsoft Entra rejected this connection's identity. Review the existing application/service principal and exact issuer, subject, and api://AzureADTokenExchange audience in step 2. Issuer discovery and JWKS must be public. " + VerificationErrorMessage(err)), nil
		}
		if err := v.sleep(ctx, 5*time.Second); err != nil {
			return VerificationResult{}, err
		}
	}
	foreign, err := ACRCredential(v.issuer, connection, ForeignSubject)
	if err != nil {
		return VerificationResult{}, err
	}
	if _, err := foreign.GetToken(ctx, tokenOptions); err == nil {
		return verificationFailure("The Entra application accepts a foreign Nuon connection subject. Restrict the federated identity credential to this exact connection subject in step 2."), nil
	} else if !acrEntraDenied(err, acrForeignSubjectDeniedPattern) {
		return VerificationResult{}, fmt.Errorf("confirm foreign Azure subject rejection: %w", err)
	}
	message = "ACR token exchange could not be checked. Review the registry login server and Microsoft Entra identity, then retry."
	if _, err := acr.GetRepositoryTokenWithCredential(ctx, credential, connection.AzureTenantID, connection.AzureRegistryLoginServer); err != nil {
		return VerificationResult{}, fmt.Errorf("exchange Azure credential for ACR token: %w", err)
	}
	return VerificationResult{Status: app.CloudConnectionStatusVerified, Message: "Azure federation, foreign-subject rejection, and ACR token exchange passed. Cloud IAM controls repository access; component builds check image pull permissions."}, nil
}

func acrEntraDenied(err error, codePattern *regexp.Regexp) bool {
	var authErr *azidentity.AuthenticationFailedError
	return errors.As(err, &authErr) && authErr.RawResponse != nil &&
		(authErr.RawResponse.StatusCode == http.StatusBadRequest || authErr.RawResponse.StatusCode == http.StatusUnauthorized) && codePattern.MatchString(authErr.Error())
}
