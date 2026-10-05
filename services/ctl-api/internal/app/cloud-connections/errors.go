package cloudconnections

import (
	"context"
	"errors"

	"github.com/aws/smithy-go"
)

func IsAccessDenied(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "AccessDenied" || apiErr.ErrorCode() == "AccessDeniedException")
}

func AssumeRoleErrorMessage(err error) string {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidIdentityToken" {
		return "AWS could not validate Nuon's identity token. Create the OIDC provider for this issuer first (step 1)."
	}
	if IsAccessDenied(err) {
		return "Nuon OIDC identity is not trusted by this role."
	}
	return VerificationErrorMessage(err)
}

func VerificationErrorMessage(err error) string {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return "Verification failed: " + apiErr.ErrorCode()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "Verification failed: request timed out"
	}
	if errors.Is(err, context.Canceled) {
		return "Verification failed: request canceled"
	}
	return "Verification failed: unable to complete the AWS verification request"
}
