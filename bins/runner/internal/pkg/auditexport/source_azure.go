package auditexport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	azruntime "github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

const azureAuditExportSecretName = "runner-audit-export"

type azureSecretClient interface {
	GetSecret(context.Context, string, string, *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
}

type azureClientFactory interface {
	New(string) (azureSecretClient, error)
}

type azureFactory struct{}

func newAzureFactory() azureClientFactory { return azureFactory{} }

func (azureFactory) New(vaultURL string) (azureSecretClient, error) {
	credential, err := azidentity.NewManagedIdentityCredential(nil)
	if err != nil {
		return nil, err
	}
	return azsecrets.NewClient(vaultURL, credential, nil)
}

type azureConfigSource struct {
	factory  azureClientFactory
	vaultURL string
	name     string
}

func newAzureConfigSource(factory azureClientFactory, installID string) configSource {
	vaultName := installID
	if len(vaultName) > 24 {
		vaultName = vaultName[:24]
	}
	return &azureConfigSource{
		factory:  factory,
		vaultURL: "https://" + vaultName + ".vault.azure.net",
		name:     azureAuditExportSecretName,
	}
}

func (s *azureConfigSource) Watch(ctx context.Context, interval time.Duration) <-chan configUpdate {
	var client azureSecretClient
	return watchConfig(ctx, interval, func() configUpdate {
		if client == nil {
			var err error
			client, err = s.factory.New(s.vaultURL)
			if err != nil {
				client = nil
				return configUpdate{
					state:         configSourceInitializationFailed,
					err:           err,
					errorIdentity: azureErrorIdentity(err),
				}
			}
		}

		result, err := client.GetSecret(ctx, s.name, "", nil)
		return newAzureConfigUpdate(result, err)
	}, nil)
}

func newAzureConfigUpdate(result azsecrets.GetSecretResponse, err error) configUpdate {
	if err != nil {
		state := configLookupFailed
		var responseError *azcore.ResponseError
		if errors.As(err, &responseError) {
			switch {
			case responseError.StatusCode == http.StatusNotFound:
				state = configNotFound
			case responseError.StatusCode == http.StatusForbidden && azureSecretDisabled(responseError):
				state = configUnavailable
			}
		}
		return configUpdate{state: state, err: err, errorIdentity: azureErrorIdentity(err)}
	}
	if result.Value == nil {
		return configUpdate{state: configAvailable}
	}
	return configUpdate{state: configAvailable, value: *result.Value}
}

func azureSecretDisabled(responseError *azcore.ResponseError) bool {
	if responseError.ErrorCode == "SecretDisabled" {
		return true
	}
	if responseError.RawResponse == nil {
		return false
	}
	payload, err := azruntime.Payload(responseError.RawResponse)
	if err != nil {
		return false
	}
	var envelope struct {
		Error struct {
			InnerError struct {
				Code string `json:"code"`
			} `json:"innererror"`
		} `json:"error"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return false
	}
	return envelope.Error.InnerError.Code == "SecretDisabled"
}

func azureErrorIdentity(err error) string {
	var responseError *azcore.ResponseError
	if errors.As(err, &responseError) {
		return fmt.Sprintf("%T:%d:%s", responseError, responseError.StatusCode, responseError.ErrorCode)
	}
	var authenticationError *azidentity.AuthenticationFailedError
	if errors.As(err, &authenticationError) {
		statusCode := 0
		if authenticationError.RawResponse != nil {
			statusCode = authenticationError.RawResponse.StatusCode
		}
		return fmt.Sprintf("%T:%d", authenticationError, statusCode)
	}
	return fmt.Sprintf("%T:%s", err, err.Error())
}
