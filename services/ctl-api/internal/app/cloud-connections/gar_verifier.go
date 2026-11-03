package cloudconnections

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (v *providerVerifier) verifyGAR(ctx context.Context, connection *app.CloudConnection, options VerifyOptions) (VerificationResult, error) {
	if connection.Preset != app.CloudConnectionPresetGAR {
		return VerificationResult{}, fmt.Errorf("GCP cloud connection requires the gar preset")
	}
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	source, err := GARTokenSource(ctx, v.issuer, connection, Subject(connection))
	if err != nil {
		return VerificationResult{}, err
	}
	deadline := v.now().Add(time.Minute)
	for {
		_, err = source.Token()
		if err == nil {
			break
		}
		if !garTrustDenied(err) {
			return VerificationResult{}, fmt.Errorf("gcp federation: %w", err)
		}
		if !options.RetryIAMPropagation || !v.now().Before(deadline) {
			return verificationFailure("GCP federation or service account impersonation failed. Review the identity provider and exact-subject binding in steps 1 and 2. " + VerificationErrorMessage(err)), nil
		}
		if err := v.sleep(ctx, 5*time.Second); err != nil {
			return VerificationResult{}, err
		}
	}
	foreign, err := GARTokenSource(ctx, v.issuer, connection, ForeignSubject)
	if err != nil {
		return VerificationResult{}, err
	}
	if _, err := foreign.Token(); err == nil {
		return verificationFailure("The GCP identity trust accepts a foreign Nuon connection subject. Review step 2."), nil
	} else if !garTrustDenied(err) {
		return VerificationResult{}, fmt.Errorf("probe foreign GCP subject: %w", err)
	}
	return VerificationResult{Status: app.CloudConnectionStatusVerified, Message: "GCP federation, service account impersonation, and foreign-subject rejection passed. Cloud IAM controls repository access; component builds check image pull permissions."}, nil
}

func garTrustDenied(err error) bool {
	message := err.Error()
	return strings.Contains(message, "rejected by the attribute condition") ||
		(strings.Contains(message, "PERMISSION_DENIED") && strings.Contains(message, "iam.serviceAccounts.getAccessToken"))
}
