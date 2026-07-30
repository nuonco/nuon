package envelope

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestCloudEventsDecode(t *testing.T) {
	body := []byte(`{"specversion":"1.0","id":"evt-1","source":"urn:test","type":"test.created","data":{"ok":true}}`)
	event, err := (CloudEvents{}).Decode(nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "evt-1" || event.Source != "urn:test" || event.Type != "test.created" || string(event.Payload) != `{"ok":true}` {
		t.Fatalf("unexpected event: %#v", event)
	}
	for _, invalid := range []string{
		`{"specversion":"0.3","id":"evt-1","source":"urn:test","type":"test","data":{}}`,
		`{"specversion":"1.0","source":"urn:test","type":"test","data":{}}`,
		`{"specversion":"1.0","id":"evt-1","source":"urn:test","type":"test"}`,
	} {
		if _, err := (CloudEvents{}).Decode(nil, []byte(invalid)); err == nil {
			t.Fatalf("invalid CloudEvent accepted: %s", invalid)
		}
	}
}

func TestRawDecodeUsesBodyDigestFallback(t *testing.T) {
	body := []byte(`{"ref":"main"}`)
	event, err := (Raw{}).Decode(nil, body)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	if event.ID != hex.EncodeToString(sum[:]) {
		t.Fatalf("fallback ID = %q", event.ID)
	}
	if _, err := (Raw{}).Decode(nil, []byte(`not json`)); err == nil {
		t.Fatal("invalid JSON event accepted")
	}
}

func TestAzureEventGridDecode(t *testing.T) {
	body := []byte(`[{"id":"evt-1","eventType":"Nuon.Proof.Created","eventTime":"2026-07-28T12:00:00Z","subject":"proof","data":{"ok":true}}]`)
	event, err := (AzureEventGrid{}).Decode(nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "evt-1" || event.Type != "Nuon.Proof.Created" || event.OccurredAt == nil || !strings.Contains(string(event.Payload), `"subject":"proof"`) {
		t.Fatalf("unexpected event: %#v", event)
	}
	if _, err := (AzureEventGrid{}).Decode(nil, []byte(`[{},{}]`)); err == nil {
		t.Fatal("accepted an Event Grid batch with more than one event")
	}
	if _, err := (AzureEventGrid{}).Decode(nil, []byte(`[{"id":"evt-1","eventType":"Nuon.Proof.Created","data":{"ok":true}}]`)); err == nil {
		t.Fatal("accepted a regular Event Grid event without eventTime")
	}
}

func TestAzureEventGridValidationCode(t *testing.T) {
	event, err := (AzureEventGrid{}).Decode(nil, []byte(`[{"id":"validation-1","eventType":"Microsoft.EventGrid.SubscriptionValidationEvent","data":{"validationCode":"code-1"}}]`))
	if err != nil {
		t.Fatal(err)
	}
	code, err := AzureEventGridValidationCode(event)
	if err != nil || code != "code-1" {
		t.Fatalf("validation code = %q: %v", code, err)
	}
}

func TestSlackEventsDecode(t *testing.T) {
	body := []byte(`{"type":"event_callback","event_id":"Ev123","event_time":1785254400,"team_id":"T123","event":{"type":"message","text":"proof"}}`)
	event, err := (SlackEvents{}).Decode(nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "Ev123" || event.Type != "message" || event.OccurredAt == nil || !strings.Contains(string(event.Payload), `"team_id":"T123"`) {
		t.Fatalf("unexpected event: %#v", event)
	}
	challengeEvent, err := (SlackEvents{}).Decode(nil, []byte(`{"type":"url_verification","challenge":"proof-code"}`))
	if err != nil {
		t.Fatal(err)
	}
	challenge, err := SlackChallenge(challengeEvent)
	if err != nil || challenge != "proof-code" {
		t.Fatalf("challenge = %q: %v", challenge, err)
	}
	for _, invalid := range []string{
		`{"type":"url_verification"}`,
		`{"type":"event_callback","event_time":1785254400,"event":{"type":"message"}}`,
		`{"type":"event_callback","event_id":"Ev123","event_time":1785254400,"event":{}}`,
	} {
		if _, err := (SlackEvents{}).Decode(nil, []byte(invalid)); err == nil {
			t.Fatalf("invalid Slack event accepted: %s", invalid)
		}
	}
}
