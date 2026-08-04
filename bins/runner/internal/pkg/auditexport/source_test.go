package auditexport

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
)

func TestConfigObservationEmitsOnlyStateChanges(t *testing.T) {
	var observation configObservation
	sameError := errors.New("region unavailable")
	differentError := errors.New("credentials unavailable")
	tests := []struct {
		name    string
		update  configUpdate
		changed bool
	}{
		{name: "first error", update: configUpdate{state: configSourceInitializationFailed, err: sameError}, changed: true},
		{name: "same error", update: configUpdate{state: configSourceInitializationFailed, err: errors.New("region unavailable")}, changed: false},
		{name: "different error", update: configUpdate{state: configSourceInitializationFailed, err: differentError}, changed: true},
		{name: "first success", update: configUpdate{state: configAvailable, value: "config-a"}, changed: true},
		{name: "unchanged success", update: configUpdate{state: configAvailable, value: "config-a"}, changed: false},
		{name: "changed contents", update: configUpdate{state: configAvailable, value: "config-b"}, changed: true},
		{name: "error after success", update: configUpdate{state: configLookupFailed, err: sameError}, changed: true},
		{name: "same lookup error", update: configUpdate{state: configLookupFailed, err: errors.New("region unavailable")}, changed: false},
		{name: "success after error", update: configUpdate{state: configAvailable, value: "config-b"}, changed: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if changed := observation.changed(test.update); changed != test.changed {
				t.Fatalf("expected changed=%t, got %t", test.changed, changed)
			}
		})
	}
}

func TestConfigObservationDeduplicatesEquivalentAWSErrors(t *testing.T) {
	var observation configObservation
	first := &smithy.GenericAPIError{Code: "AccessDeniedException", Message: "access denied"}
	second := &smithy.GenericAPIError{Code: "AccessDeniedException", Message: "access denied"}

	if !observation.changed(configUpdate{state: configLookupFailed, err: first, errorIdentity: awsErrorIdentity(first)}) {
		t.Fatal("first AWS error was not emitted")
	}
	if observation.changed(configUpdate{state: configLookupFailed, err: second, errorIdentity: awsErrorIdentity(second)}) {
		t.Fatal("equivalent AWS error was emitted again")
	}
}

func TestNewAWSConfigUpdateClassifiesResults(t *testing.T) {
	value := "config"
	tests := []struct {
		name  string
		value *secretsmanager.GetSecretValueOutput
		err   error
		state configUpdateState
	}{
		{name: "available", value: &secretsmanager.GetSecretValueOutput{SecretString: &value}, state: configAvailable},
		{name: "not found", err: &types.ResourceNotFoundException{}, state: configNotFound},
		{name: "unavailable", err: &types.InvalidRequestException{}, state: configUnavailable},
		{name: "lookup failure", err: errors.New("lookup failed"), state: configLookupFailed},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			update := newAWSConfigUpdate(test.value, test.err)
			if update.state != test.state {
				t.Fatalf("expected state %d, got %d", test.state, update.state)
			}
			if test.name == "available" && update.value != value {
				t.Fatalf("expected value %q, got %q", value, update.value)
			}
		})
	}
}

func TestConfigSourceResolverSelectsPlatformSource(t *testing.T) {
	resolver := newConfigSourceResolver(awsFactory{}, azureFactory{})
	tests := []struct {
		name         string
		platform     string
		installID    string
		wantName     string
		wantVaultURL string
	}{
		{name: "aws", platform: "aws", installID: "inst-test", wantName: "nuon/inst-test/runner-audit-export"},
		{name: "aws variant", platform: "aws-eks", installID: "inst-test", wantName: "nuon/inst-test/runner-audit-export"},
		{name: "azure", platform: "azure", installID: "instabcdefghijklmnopqrstuv", wantName: azureAuditExportSecretName, wantVaultURL: "https://instabcdefghijklmnopqrst.vault.azure.net"},
		{name: "azure variant", platform: "azure-aks", installID: "inst-test", wantName: azureAuditExportSecretName, wantVaultURL: "https://inst-test.vault.azure.net"},
		{name: "unsupported platform", platform: "gcp", installID: "inst-test"},
		{name: "missing install", platform: "aws"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := resolver.Resolve(test.platform, test.installID)
			if test.wantName == "" {
				if source != nil {
					t.Fatal("expected no configuration source")
				}
				return
			}
			if test.wantVaultURL != "" {
				azureSource, ok := source.(*azureConfigSource)
				if !ok {
					t.Fatalf("expected Azure configuration source, got %T", source)
				}
				if azureSource.name != test.wantName || azureSource.vaultURL != test.wantVaultURL {
					t.Fatalf("unexpected Azure source: name=%q vault=%q", azureSource.name, azureSource.vaultURL)
				}
				return
			}
			awsSource, ok := source.(*awsConfigSource)
			if !ok || awsSource.name != test.wantName {
				t.Fatalf("unexpected AWS source: %#v", source)
			}
		})
	}
}
