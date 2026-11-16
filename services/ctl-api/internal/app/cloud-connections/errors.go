package cloudconnections

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/aws/smithy-go"
	"golang.org/x/oauth2"

	"github.com/nuonco/nuon/pkg/redact"
)

func IsAccessDenied(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "AccessDenied" || apiErr.ErrorCode() == "AccessDeniedException")
}

func AssumeRoleErrorMessage(err error) string {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidIdentityToken" {
		return "AWS could not validate Nuon's identity token. Create the OIDC provider for this issuer first (step 1). " + VerificationErrorMessage(err)
	}
	if IsAccessDenied(err) {
		return "Nuon OIDC identity is not trusted by this role. " + VerificationErrorMessage(err)
	}
	return VerificationErrorMessage(err)
}

func VerificationErrorMessage(err error) string {
	var azureErr *acrVerificationError
	if errors.As(err, &azureErr) {
		return azureErr.Error()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "Verification failed: request timed out"
	}
	if errors.Is(err, context.Canceled) {
		return "Verification failed: request canceled"
	}
	if err == nil {
		return "Verification failed: unable to complete the cloud verification request"
	}
	var diagnostic string
	var apiErr smithy.APIError
	var authErr *azidentity.AuthenticationFailedError
	var oauthErr *oauth2.RetrieveError
	switch {
	case errors.As(err, &apiErr):
		diagnostic = strings.TrimSuffix(apiErr.ErrorCode()+": "+apiErr.ErrorMessage(), ": ")
	case errors.As(err, &authErr) && authErr.RawResponse != nil:
		diagnostic = fmt.Sprintf("Microsoft Entra returned status %d", authErr.RawResponse.StatusCode)
		if message := authErr.Error(); strings.Contains(message, "{") {
			if detail := verificationProviderError(message[strings.Index(message, "{"):]); detail != "" {
				diagnostic += ": " + detail
			}
		}
	case errors.As(err, &oauthErr):
		diagnostic = strings.TrimSuffix(oauthErr.ErrorCode+": "+oauthErr.ErrorDescription, ": ")
		if diagnostic == "" && oauthErr.Response != nil {
			diagnostic = fmt.Sprintf("OAuth token request returned status %d", oauthErr.Response.StatusCode)
		}
	default:
		diagnostic = err.Error()
		if i := strings.Index(diagnostic, "oauth2/google: status code "); i >= 0 {
			parts := strings.SplitN(diagnostic[i:], ": ", 3)
			diagnostic = parts[1]
			if len(parts) == 3 {
				if detail := verificationProviderError(parts[2]); detail != "" {
					diagnostic += ": " + detail
				}
			}
		}
	}
	diagnostic = redact.DiagnosticSecrets(diagnostic)
	diagnostic = strings.Join(strings.Fields(diagnostic), " ")
	if len(diagnostic) > 2048 {
		diagnostic = strings.ToValidUTF8(diagnostic[:2048], "") + "..."
	}
	if diagnostic == "" {
		diagnostic = "unable to complete the cloud verification request"
	}
	return "Verification failed: " + diagnostic
}

func verificationProviderError(body string) string {
	if len(body) > 16*1024 {
		return ""
	}
	var failure struct {
		Error            json.RawMessage `json:"error"`
		ErrorDescription string          `json:"error_description"`
	}
	if json.NewDecoder(strings.NewReader(body)).Decode(&failure) != nil {
		return ""
	}
	var code string
	if json.Unmarshal(failure.Error, &code) == nil && code != "" {
		return strings.TrimSuffix(code+": "+failure.ErrorDescription, ": ")
	}
	var googleError struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if json.Unmarshal(failure.Error, &googleError) == nil && googleError.Status != "" {
		return strings.TrimSuffix(googleError.Status+": "+googleError.Message, ": ")
	}
	return ""
}
