package auditexport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type azureFactoryStub struct {
	vaultURL string
	client   azureSecretClient
}

func (f *azureFactoryStub) New(vaultURL string) (azureSecretClient, error) {
	f.vaultURL = vaultURL
	return f.client, nil
}

type azureSecretClientStub struct {
	name    string
	version string
	result  azsecrets.GetSecretResponse
}

func (c *azureSecretClientStub) GetSecret(_ context.Context, name, version string, _ *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	c.name = name
	c.version = version
	return c.result, nil
}

func TestNewAzureConfigUpdateClassifiesResults(t *testing.T) {
	value := "config"
	disabledError := &azcore.ResponseError{
		StatusCode: http.StatusForbidden,
		ErrorCode:  "Forbidden",
		RawResponse: &http.Response{
			StatusCode: http.StatusForbidden,
			Status:     "403 Forbidden",
			Body:       io.NopCloser(strings.NewReader("{\"error\":{\"code\":\"Forbidden\",\"innererror\":{\"code\":\"SecretDisabled\"}}}")),
		},
	}
	tests := []struct {
		name   string
		result azsecrets.GetSecretResponse
		err    error
		state  configUpdateState
		value  string
	}{
		{name: "available", result: azsecrets.GetSecretResponse{Secret: azsecrets.Secret{Value: &value}}, state: configAvailable, value: value},
		{name: "empty", result: azsecrets.GetSecretResponse{}, state: configAvailable},
		{name: "disabled", err: disabledError, state: configUnavailable},
		{name: "not found", err: &azcore.ResponseError{StatusCode: http.StatusNotFound, ErrorCode: "SecretNotFound"}, state: configNotFound},
		{name: "permission denied", err: &azcore.ResponseError{StatusCode: http.StatusForbidden, ErrorCode: "Forbidden"}, state: configLookupFailed},
		{name: "lookup failure", err: errors.New("lookup failed"), state: configLookupFailed},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			update := newAzureConfigUpdate(test.result, test.err)
			if update.state != test.state || update.value != test.value {
				t.Fatalf("expected state=%d value=%q, got state=%d value=%q", test.state, test.value, update.state, update.value)
			}
		})
	}
}

func TestAzureConfigSourceReadsLatestSecret(t *testing.T) {
	value := "config"
	client := &azureSecretClientStub{result: azsecrets.GetSecretResponse{Secret: azsecrets.Secret{Value: &value}}}
	factory := &azureFactoryStub{client: client}
	source := newAzureConfigSource(factory, "instabcdefghijklmnopqrstuv")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	update := <-source.Watch(ctx, time.Hour)
	if update.state != configAvailable || update.value != value {
		t.Fatalf("unexpected configuration update: %#v", update)
	}
	if factory.vaultURL != "https://instabcdefghijklmnopqrst.vault.azure.net" {
		t.Fatalf("unexpected vault URL: %q", factory.vaultURL)
	}
	if client.name != azureAuditExportSecretName || client.version != "" {
		t.Fatalf("unexpected secret request: name=%q version=%q", client.name, client.version)
	}
}

func TestConfigObservationDeduplicatesEquivalentAzureErrors(t *testing.T) {
	var observation configObservation
	first := &azcore.ResponseError{StatusCode: http.StatusForbidden, ErrorCode: "Forbidden"}
	second := &azcore.ResponseError{StatusCode: http.StatusForbidden, ErrorCode: "Forbidden"}

	if !observation.changed(configUpdate{state: configLookupFailed, err: first, errorIdentity: azureErrorIdentity(first)}) {
		t.Fatal("first Azure error was not emitted")
	}
	if observation.changed(configUpdate{state: configLookupFailed, err: second, errorIdentity: azureErrorIdentity(second)}) {
		t.Fatal("equivalent Azure error was emitted again")
	}
}

func TestAzureErrorIdentityDeduplicatesAuthenticationFailures(t *testing.T) {
	first := &azidentity.AuthenticationFailedError{RawResponse: &http.Response{StatusCode: http.StatusBadRequest}}
	second := &azidentity.AuthenticationFailedError{RawResponse: &http.Response{StatusCode: http.StatusBadRequest}}
	if azureErrorIdentity(first) != azureErrorIdentity(second) {
		t.Fatal("equivalent Azure authentication errors have different identities")
	}
}
