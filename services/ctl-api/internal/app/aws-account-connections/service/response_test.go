package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestResponseTrustMetadataRedaction(t *testing.T) {
	connection := &app.AWSAccountConnection{ID: "awc", AccountID: "123456789012", ExternalID: "secret-external-id"}
	listJSON, err := json.Marshal(response(connection, "", false))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(listJSON), connection.ExternalID) || strings.Contains(string(listJSON), "trust_policy") {
		t.Fatalf("list response leaked trust metadata: %s", listJSON)
	}
	detailJSON, err := json.Marshal(response(connection, "arn:aws:iam::999999999999:role/management", true))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(detailJSON), connection.ExternalID) || !strings.Contains(string(detailJSON), "sts:ExternalId") {
		t.Fatalf("detail response omitted trust metadata: %s", detailJSON)
	}
}
