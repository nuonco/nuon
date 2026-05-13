package generics

import (
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
)

// TestNestedMapRoundTrip locks in the fix for the customer-input truncation
// bug. The phone-home payload contains nested maps (e.g. install_inputs) that
// land in hstore. Without EncodeNestedForHstore, ToStringMap stringifies them
// via fmt.Sprintf("%v", v) and the decode hook reads them back by splitting
// on whitespace — silently truncating any value with a space, colon, or
// bracket. With the JSON branch wired up, those values round-trip losslessly.
func TestNestedMapRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{"plain", "abc123etc"},
		{"with_space", "ssh abc123etc"},
		{"with_colons_and_brackets", "ssh abc:def] more"},
		{"empty", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Producer side: phone-home builds a map[string]any payload with
			// install_inputs as a nested map[string]string. ToStringMap then
			// flattens it for hstore storage.
			payload := map[string]any{
				"install_inputs": map[string]string{"my_key": tc.value},
				"vpc_id":         "vpc-123", // scalar, untouched
			}
			stored := ToStringMap(EncodeNestedForHstore(payload))

			// Consumer side: hstore is read back into a map[string]any with
			// the same decode hook used in InstallStackVersionRun.AfterQuery
			// and the v2 update_install_stack_outputs signal.
			var decoded map[string]any
			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				DecodeHook:       StringToMapDecodeHook(),
				WeaklyTypedInput: true,
				Result:           &decoded,
			})
			if err != nil {
				t.Fatalf("decoder: %v", err)
			}
			if err := decoder.Decode(stored); err != nil {
				t.Fatalf("decode: %v", err)
			}

			got, ok := decoded["install_inputs"].(map[string]any)
			if !ok {
				t.Fatalf("install_inputs missing or wrong type: %T", decoded["install_inputs"])
			}
			want := map[string]any{"my_key": tc.value}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("install_inputs mismatch:\n  got:  %#v\n  want: %#v", got, want)
			}
			if decoded["vpc_id"] != "vpc-123" {
				t.Fatalf("vpc_id mangled: %v", decoded["vpc_id"])
			}
		})
	}
}

// TestLegacyHstoreFormatStillDecodes guards back-compat. Rows written before
// this fix still carry fmt.Sprintf("%v", map) text. The legacy branch of the
// hook must keep parsing them — lossy as it always was, but at least it
// shouldn't crash or change behavior for pre-existing data.
func TestLegacyHstoreFormatStillDecodes(t *testing.T) {
	stored := map[string]string{
		"install_inputs": "map[my_key:legacy_value]",
	}
	var decoded map[string]any
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       StringToMapDecodeHook(),
		WeaklyTypedInput: true,
		Result:           &decoded,
	})
	if err != nil {
		t.Fatalf("decoder: %v", err)
	}
	if err := decoder.Decode(stored); err != nil {
		t.Fatalf("decode: %v", err)
	}
	got, ok := decoded["install_inputs"].(map[string]string)
	if !ok {
		t.Fatalf("legacy format must decode to map[string]string, got %T", decoded["install_inputs"])
	}
	if got["my_key"] != "legacy_value" {
		t.Fatalf("legacy decode mismatch: %v", got)
	}
}
