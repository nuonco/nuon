package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var (
	ErrMissingCloudAccount  = errors.New("cloud account identifier is required")
	ErrCloudAccountMismatch = errors.New("cloud account does not match the install")
)

// CheckObservedCloudAccount rejects a verified caller whose payload names a different
// cloud account than the install is pinned to.
//
// Separate from caller verification on purpose: a valid credential proves *who* is
// calling, this proves the call is about the account we expect. A mismatch means either
// the install was created against the wrong account or a credential escaped into a
// different one, and neither should be allowed to overwrite stack outputs.
func CheckObservedCloudAccount(install *app.Install, props map[string]any, required bool) error {
	if required {
		var payloadKey, expected string
		switch install.CloudPlatform {
		case app.CloudPlatformAWS:
			payloadKey, expected = "account_id", install.ExpectedAccountID
		case app.CloudPlatformAzure:
			payloadKey, expected = "subscription_id", install.ExpectedSubscriptionID
		case app.CloudPlatformGCP:
			payloadKey, expected = "project_id", install.ExpectedProjectID
		}

		observed, _ := props[payloadKey].(string)
		if payloadKey == "" || observed == "" || expected == "" {
			return fmt.Errorf("%w: %s install %s", ErrMissingCloudAccount, install.CloudPlatform, install.ID)
		}

		if install.CloudPlatform == app.CloudPlatformGCP {
			if err := checkGCPServiceAccountProjects(expected, props); err != nil {
				return err
			}
		}
	}

	for _, field := range []struct {
		payloadKey string
		expected   string
	}{
		{"account_id", install.ExpectedAccountID},
		{"subscription_id", install.ExpectedSubscriptionID},
		{"project_id", install.ExpectedProjectID},
	} {
		observed, _ := props[field.payloadKey].(string)
		if observed == "" || field.expected == "" || strings.EqualFold(observed, field.expected) {
			continue
		}

		return fmt.Errorf("%w: payload %s %s does not match the account expected for this install",
			ErrCloudAccountMismatch, field.payloadKey, observed)
	}

	return nil
}

func checkGCPServiceAccountProjects(projectID string, props map[string]any) error {
	suffix := "@" + strings.ToLower(projectID) + ".iam.gserviceaccount.com"
	check := func(key, email string) error {
		if email == "" || strings.HasSuffix(strings.ToLower(email), suffix) {
			return nil
		}

		return fmt.Errorf("%w: payload %s %s is not in the project expected for this install",
			ErrCloudAccountMismatch, key, email)
	}

	for key, value := range props {
		switch {
		case key == "runner_service_account_email" || strings.HasSuffix(key, "_sa_email"):
			email, ok := value.(string)
			if !ok && value != nil {
				return fmt.Errorf("%w: payload %s must be a string", ErrCloudAccountMismatch, key)
			}
			if err := check(key, email); err != nil {
				return err
			}
		case strings.HasSuffix(key, "_sa_emails"):
			emails, err := gcpServiceAccountEmails(value)
			if err != nil {
				return fmt.Errorf("%w: payload %s: %w", ErrCloudAccountMismatch, key, err)
			}
			for name, email := range emails {
				if err := check(key+"."+name, email); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func gcpServiceAccountEmails(value any) (map[string]string, error) {
	var raw map[string]any
	switch v := value.(type) {
	case nil:
		return nil, nil
	case map[string]any:
		raw = v
	case string:
		if v == "" {
			return nil, nil
		}
		if err := json.Unmarshal([]byte(v), &raw); err != nil {
			return nil, errors.New("must be an object of strings")
		}
	default:
		return nil, errors.New("must be an object of strings")
	}

	emails := make(map[string]string, len(raw))
	for name, email := range raw {
		str, ok := email.(string)
		if !ok {
			return nil, errors.New("must be an object of strings")
		}
		emails[name] = str
	}

	return emails, nil
}
