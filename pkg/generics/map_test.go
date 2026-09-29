package generics

import (
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
)

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
			payload := map[string]any{
				"install_inputs": map[string]string{"my_key": tc.value},
				"vpc_id":         "vpc-123",
			}
			stored := ToStringMap(EncodeNestedForHstore(payload))

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
