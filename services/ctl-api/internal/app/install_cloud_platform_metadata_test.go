package app

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCloudPlatformMetadataRoundTrips(t *testing.T) {
	original := CloudPlatformMetadata{
		TargetAccountID:        "123456789012",
		ObservedAccountID:      "210987654321",
		TargetProjectID:        "my-project",
		ObservedProjectID:      "other-project",
		TargetSubscriptionID:   "sub-a",
		ObservedSubscriptionID: "sub-b",
		TargetSource:           CloudPlatformTargetSourceUser,
	}

	value, err := original.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	var scanned CloudPlatformMetadata
	if err := scanned.Scan(value); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if scanned != original {
		t.Fatalf("round trip mismatch:\n got %#v\nwant %#v", scanned, original)
	}
}

// The driver may hand back either []byte or string depending on the connection;
// ComponentHealthContext.Scan silently no-ops on string, which we must not repeat.
func TestCloudPlatformMetadataScanAcceptsStringAndBytes(t *testing.T) {
	const raw = `{"target_account_id":"123456789012"}`

	for name, input := range map[string]any{
		"bytes":  []byte(raw),
		"string": raw,
	} {
		var cpm CloudPlatformMetadata
		if err := cpm.Scan(input); err != nil {
			t.Fatalf("%s: Scan: %v", name, err)
		}
		if cpm.TargetAccountID != "123456789012" {
			t.Errorf("%s: target account not scanned, got %q", name, cpm.TargetAccountID)
		}
	}
}

func TestCloudPlatformMetadataScanEdgeCases(t *testing.T) {
	var cpm CloudPlatformMetadata
	if err := cpm.Scan(nil); err != nil {
		t.Errorf("nil should scan cleanly: %v", err)
	}
	if err := cpm.Scan([]byte{}); err != nil {
		t.Errorf("empty bytes should scan cleanly: %v", err)
	}
	if err := cpm.Scan(42); err == nil {
		t.Error("an unsupported type should error rather than silently no-op")
	}
}

func TestPhoneHomeAuthRoundTrips(t *testing.T) {
	verified := time.Now().UTC().Truncate(time.Second)
	original := PhoneHomeAuth{
		SecretARN:      "arn:aws:secretsmanager:us-west-2:123456789012:secret:nuon/phone-home/x-aB3xYz",
		SecretRegion:   "us-west-2",
		KMSKeyARN:      "arn:aws:kms:us-west-2:123456789012:key/abc",
		CreatedAt:      verified,
		LastVerifiedAt: &verified,
	}

	value, err := original.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	var scanned PhoneHomeAuth
	if err := scanned.Scan(value); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if scanned.SecretARN != original.SecretARN || scanned.SecretRegion != original.SecretRegion || scanned.KMSKeyARN != original.KMSKeyARN {
		t.Fatalf("round trip mismatch:\n got %#v\nwant %#v", scanned, original)
	}
	if scanned.LastVerifiedAt == nil || !scanned.LastVerifiedAt.Equal(verified) {
		t.Fatalf("LastVerifiedAt did not survive: %v", scanned.LastVerifiedAt)
	}
	if scanned.LastRejectedAt != nil {
		t.Errorf("LastRejectedAt should stay nil, got %v", scanned.LastRejectedAt)
	}
}

// The whole reason PhoneHomeAuth is its own column: it must persist to jsonb but never
// reach the wire. A nested json:"-" field would have failed both. Only the Status()
// projection is serialized, under the phone_home_auth name.
func TestPhoneHomeAuthPersistsButIsNotSerialized(t *testing.T) {
	provisioned := time.Now().UTC().Truncate(time.Second)
	rejected := provisioned.Add(time.Hour)
	auth := &PhoneHomeAuth{
		SecretARN:      "arn:aws:secretsmanager:us-west-2:123456789012:secret:nuon/phone-home/inst0-aB3xYz",
		SecretRegion:   "us-west-2",
		KMSKeyARN:      "arn:aws:kms:us-west-2:123456789012:key/abc",
		CreatedAt:      provisioned,
		LastRejectedAt: &rejected,
	}
	install := Install{
		ID: "inst00000000000000000000000",
		CloudPlatformMetadata: CloudPlatformMetadata{
			TargetAccountID: "123456789012",
			TargetSource:    CloudPlatformTargetSourceConnection,
		},
		PhoneHomeAuth:       auth,
		PhoneHomeAuthStatus: auth.Status(),
	}

	body, err := json.Marshal(install)
	if err != nil {
		t.Fatalf("marshal install: %v", err)
	}
	rendered := string(body)

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal install: %v", err)
	}
	for _, leaked := range []string{"aB3xYz", "secret_arn", "secret_region", "kms_key_arn", "secretsmanager", "arn:aws:kms"} {
		if strings.Contains(rendered, leaked) {
			t.Errorf("install response leaked %q", leaked)
		}
	}

	rawStatus, ok := payload["phone_home_auth"]
	if !ok {
		t.Fatal("phone_home_auth must serialize the wire-safe status projection")
	}
	var status PhoneHomeAuthStatus
	if err := json.Unmarshal(rawStatus, &status); err != nil {
		t.Fatalf("unmarshal phone_home_auth: %v", err)
	}
	if !status.ProvisionedAt.Equal(provisioned) {
		t.Errorf("provisioned_at did not serialize: got %v want %v", status.ProvisionedAt, provisioned)
	}
	if status.LastRejectedAt == nil || !status.LastRejectedAt.Equal(rejected) {
		t.Errorf("last_rejected_at did not serialize: %v", status.LastRejectedAt)
	}
	if status.LastVerifiedAt != nil {
		t.Errorf("last_verified_at should stay absent, got %v", status.LastVerifiedAt)
	}

	cpm, ok := payload["cloud_platform_metadata"]
	if !ok {
		t.Fatal("cloud_platform_metadata must be serialized on the install")
	}
	var decoded CloudPlatformMetadata
	if err := json.Unmarshal(cpm, &decoded); err != nil {
		t.Fatalf("unmarshal cloud_platform_metadata: %v", err)
	}
	if decoded.TargetAccountID != "123456789012" || decoded.TargetSource != CloudPlatformTargetSourceConnection {
		t.Errorf("cloud_platform_metadata did not serialize: %#v", decoded)
	}

	// ...and it still persists to the column.
	value, err := install.PhoneHomeAuth.Value()
	if err != nil {
		t.Fatalf("PhoneHomeAuth.Value: %v", err)
	}
	persisted, ok := value.([]byte)
	if !ok {
		t.Fatalf("PhoneHomeAuth.Value should marshal to bytes, got %T", value)
	}
	if !strings.Contains(string(persisted), "aB3xYz") {
		t.Errorf("PhoneHomeAuth must persist its secret ARN to jsonb, got %s", persisted)
	}
}

func TestPhoneHomeAuthStatusOmittedWhenUnprovisioned(t *testing.T) {
	if status := (*PhoneHomeAuth)(nil).Status(); status != nil {
		t.Errorf("Status on a nil column should stay nil, got %#v", status)
	}

	body, err := json.Marshal(Install{ID: "inst00000000000000000000000"})
	if err != nil {
		t.Fatalf("marshal install: %v", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal install: %v", err)
	}
	if _, ok := payload["phone_home_auth"]; ok {
		t.Error("phone_home_auth must be absent until credentials are provisioned")
	}

	// A rejection recorded before provisioning leaves CreatedAt zero.
	rejected := time.Now().UTC()
	body, err = json.Marshal((&PhoneHomeAuth{LastRejectedAt: &rejected}).Status())
	if err != nil {
		t.Fatalf("marshal status: %v", err)
	}
	if strings.Contains(string(body), "provisioned_at") {
		t.Errorf("a zero provisioned_at must be omitted, got %s", body)
	}
}

func TestCloudPlatformMetadataHasTarget(t *testing.T) {
	for name, tc := range map[string]struct {
		metadata CloudPlatformMetadata
		want     bool
	}{
		"empty":               {want: false},
		"aws target":          {metadata: CloudPlatformMetadata{TargetAccountID: "123456789012"}, want: true},
		"gcp target":          {metadata: CloudPlatformMetadata{TargetProjectID: "p"}, want: true},
		"azure target":        {metadata: CloudPlatformMetadata{TargetSubscriptionID: "s"}, want: true},
		"observed only":       {metadata: CloudPlatformMetadata{ObservedAccountID: "123456789012"}, want: false},
		"target source alone": {metadata: CloudPlatformMetadata{TargetSource: "user"}, want: false},
	} {
		if got := tc.metadata.HasTarget(); got != tc.want {
			t.Errorf("%s: HasTarget() = %v, want %v", name, got, tc.want)
		}
	}
}

func TestSetExpectedCloudIdentifiers(t *testing.T) {
	for name, tc := range map[string]struct {
		metadata                    CloudPlatformMetadata
		account, project, subscript string
	}{
		"target wins over observed": {
			metadata: CloudPlatformMetadata{
				TargetAccountID:        "111111111111",
				ObservedAccountID:      "222222222222",
				TargetProjectID:        "target-project",
				ObservedProjectID:      "observed-project",
				TargetSubscriptionID:   "target-sub",
				ObservedSubscriptionID: "observed-sub",
			},
			account: "111111111111", project: "target-project", subscript: "target-sub",
		},
		"falls back to observed": {
			metadata: CloudPlatformMetadata{
				ObservedAccountID:      "222222222222",
				ObservedProjectID:      "observed-project",
				ObservedSubscriptionID: "observed-sub",
			},
			account: "222222222222", project: "observed-project", subscript: "observed-sub",
		},
		"empty when neither set": {},
	} {
		install := Install{CloudPlatformMetadata: tc.metadata}
		install.setExpectedCloudIdentifiers()

		if install.ExpectedAccountID != tc.account {
			t.Errorf("%s: ExpectedAccountID = %q, want %q", name, install.ExpectedAccountID, tc.account)
		}
		if install.ExpectedProjectID != tc.project {
			t.Errorf("%s: ExpectedProjectID = %q, want %q", name, install.ExpectedProjectID, tc.project)
		}
		if install.ExpectedSubscriptionID != tc.subscript {
			t.Errorf("%s: ExpectedSubscriptionID = %q, want %q", name, install.ExpectedSubscriptionID, tc.subscript)
		}
	}
}
