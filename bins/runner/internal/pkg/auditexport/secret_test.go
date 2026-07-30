package auditexport

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
)

func TestSecretObservationEmitsOnlyStateChanges(t *testing.T) {
	var observation secretObservation
	sameError := errors.New("region unavailable")
	differentError := errors.New("credentials unavailable")
	tests := []struct {
		name    string
		update  secretUpdate
		changed bool
	}{
		{name: "first error", update: secretUpdate{state: secretInitializationFailed, err: sameError}, changed: true},
		{name: "same error", update: secretUpdate{state: secretInitializationFailed, err: errors.New("region unavailable")}, changed: false},
		{name: "different error", update: secretUpdate{state: secretInitializationFailed, err: differentError}, changed: true},
		{name: "first success", update: secretUpdate{state: secretAvailable, value: "config-a"}, changed: true},
		{name: "unchanged success", update: secretUpdate{state: secretAvailable, value: "config-a"}, changed: false},
		{name: "changed contents", update: secretUpdate{state: secretAvailable, value: "config-b"}, changed: true},
		{name: "error after success", update: secretUpdate{state: secretLookupFailed, err: sameError}, changed: true},
		{name: "same lookup error", update: secretUpdate{state: secretLookupFailed, err: errors.New("region unavailable")}, changed: false},
		{name: "success after error", update: secretUpdate{state: secretAvailable, value: "config-b"}, changed: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if changed := observation.changed(test.update); changed != test.changed {
				t.Fatalf("expected changed=%t, got %t", test.changed, changed)
			}
		})
	}
}

func TestSecretObservationDeduplicatesEquivalentAWSErrors(t *testing.T) {
	var observation secretObservation
	first := &smithy.GenericAPIError{Code: "AccessDeniedException", Message: "access denied"}
	second := &smithy.GenericAPIError{Code: "AccessDeniedException", Message: "access denied"}

	if !observation.changed(secretUpdate{state: secretLookupFailed, err: first}) {
		t.Fatal("first AWS error was not emitted")
	}
	if observation.changed(secretUpdate{state: secretLookupFailed, err: second}) {
		t.Fatal("equivalent AWS error was emitted again")
	}
}

func TestNewSecretUpdateClassifiesAWSResults(t *testing.T) {
	value := "config"
	tests := []struct {
		name  string
		value *secretsmanager.GetSecretValueOutput
		err   error
		state secretUpdateState
	}{
		{name: "available", value: &secretsmanager.GetSecretValueOutput{SecretString: &value}, state: secretAvailable},
		{name: "not found", err: &types.ResourceNotFoundException{}, state: secretNotFound},
		{name: "unavailable", err: &types.InvalidRequestException{}, state: secretUnavailable},
		{name: "lookup failure", err: errors.New("lookup failed"), state: secretLookupFailed},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			update := newSecretUpdate(test.value, test.err)
			if update.state != test.state {
				t.Fatalf("expected state %d, got %d", test.state, update.state)
			}
			if test.name == "available" && update.value != value {
				t.Fatalf("expected value %q, got %q", value, update.value)
			}
		})
	}
}
