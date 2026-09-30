package installgrouperrors

import "testing"

func TestInstallUpdateFailedErrorMessage(t *testing.T) {
	named := &InstallUpdateFailedError{InstallID: "ins_1", InstallName: "jm-test-001"}
	if named.Error() != "jm-test-001 failed during deploy" {
		t.Fatalf("message = %q", named.Error())
	}
	if named.Type() != InstallUpdateFailedType {
		t.Fatalf("type = %q", named.Type())
	}

	unnamed := &InstallUpdateFailedError{InstallID: "ins_1"}
	if unnamed.Error() != "ins_1 failed during deploy" {
		t.Fatalf("message = %q", unnamed.Error())
	}
}
