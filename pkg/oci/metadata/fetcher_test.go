package metadata

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestDecodeInTotoStatement(t *testing.T) {
	statement := `{
		"_type": "https://in-toto.io/Statement/v0.1",
		"predicateType": "https://slsa.dev/provenance/v0.2",
		"subject": [{"name": "example/image"}],
		"predicate": {"materials": []}
	}`

	tests := []struct {
		name string
		data string
	}{
		{name: "direct", data: statement},
		{
			name: "DSSE envelope",
			data: fmt.Sprintf(`{"payloadType":"application/vnd.in-toto+json","payload":%q}`, base64.StdEncoding.EncodeToString([]byte(statement))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := decodeInTotoStatement([]byte(tt.data))
			if err != nil {
				t.Fatalf("decodeInTotoStatement() error = %v", err)
			}
			if decoded.PredicateType != "https://slsa.dev/provenance/v0.2" {
				t.Fatalf("PredicateType = %q", decoded.PredicateType)
			}
			if len(decoded.Subject) != 1 || decoded.Subject[0].Name != "example/image" {
				t.Fatalf("Subject = %#v", decoded.Subject)
			}
		})
	}
}
