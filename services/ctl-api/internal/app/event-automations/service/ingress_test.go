package service

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuecctx"
)

func TestParseCloudEvent(t *testing.T) {
	valid := []byte("{\"specversion\":\"1.0\",\"id\":\"evt-1\",\"source\":\"urn:test\",\"type\":\"test.created\",\"data\":{\"ok\":true}}")
	event, err := parseCloudEvent(valid)
	if err != nil || event.ID != "evt-1" {
		t.Fatalf("valid CloudEvent rejected: %v", err)
	}
	for _, body := range []string{
		"{\"specversion\":\"0.3\",\"id\":\"evt-1\",\"source\":\"urn:test\",\"type\":\"test\",\"data\":{}}",
		"{\"specversion\":\"1.0\",\"source\":\"urn:test\",\"type\":\"test\",\"data\":{}}",
		"{\"specversion\":\"1.0\",\"id\":\"evt-1\",\"source\":\"urn:test\",\"type\":\"test\"}",
	} {
		if _, err := parseCloudEvent([]byte(body)); err == nil {
			t.Fatalf("invalid CloudEvent accepted: %s", body)
		}
	}
}

func TestDecodeGenericJSONEvent(t *testing.T) {
	headers := http.Header{"X-Nuon-Event-Id": {"delivery-1"}, "X-Nuon-Event-Type": {"push"}, "Content-Type": {"application/json"}}
	source := &app.EventSource{Envelope: app.EventEnvelopeTypeNone, AuthType: app.EventSourceAuthTypeHMAC, IDFrom: app.EventFieldSelector{Header: "X-Nuon-Event-ID"}, TypeFrom: app.EventFieldSelector{Header: "X-Nuon-Event-Type"}}
	event, err := decodeEvent(source, headers, []byte(`{"ref":"main"}`))
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "delivery-1" || event.Type != "push" || string(event.Payload) != `{"ref":"main"}` {
		t.Fatalf("unexpected normalized event: %#v", event)
	}
	if _, err := decodeEvent(source, nil, []byte(`not json`)); err == nil {
		t.Fatal("invalid JSON event accepted")
	}
}

func TestDecodeCloudEventEnvelope(t *testing.T) {
	body := []byte(`{"specversion":"1.0","id":"evt-1","source":"urn:test","type":"test.created","data":{"ok":true}}`)
	event, err := decodeEvent(&app.EventSource{Envelope: app.EventEnvelopeTypeCloudEvents}, nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "evt-1" || event.Type != "test.created" || string(event.Payload) != `{"ok":true}` {
		t.Fatalf("unexpected normalized event: %#v", event)
	}
}

func TestDecodeGenericJSONPayloadSelectors(t *testing.T) {
	source := &app.EventSource{
		Envelope: app.EventEnvelopeTypeNone,
		AuthType: app.EventSourceAuthTypeAPIKey,
		IDFrom:   app.EventFieldSelector{Payload: "$.delivery.id"},
		TypeFrom: app.EventFieldSelector{Payload: `$['detail-type']`},
	}
	event, err := decodeEvent(source, nil, []byte(`{"delivery":{"id":"evt-1"},"detail-type":"image.push"}`))
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "evt-1" || event.Type != "image.push" {
		t.Fatalf("unexpected normalized event: %#v", event)
	}

	source.IDFrom.Payload = "$.delivery[*]"
	if err := validateEventFieldSelector(source.IDFrom); err == nil {
		t.Fatal("wildcard selector accepted in singular context")
	}
}

func TestDecodePubSubPushEnvelope(t *testing.T) {
	body := []byte(`{"message":{"data":"eyJhY3Rpb24iOiJwdXNoIn0=","messageId":"msg-1","publishTime":"2026-07-22T12:00:00Z"}}`)
	event, err := decodeEvent(&app.EventSource{Envelope: app.EventEnvelopeTypePubSubPush}, nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "msg-1" || string(event.Payload) != `{"action":"push"}` || event.OccurredAt == nil {
		t.Fatalf("unexpected normalized event: %#v", event)
	}
}

func TestEventSourcePresets(t *testing.T) {
	req := createEventSourceRequest{Preset: "github"}
	if err := applyEventSourcePreset(&req); err != nil {
		t.Fatal(err)
	}
	if err := defaultAndValidateAuthConfig(&req); err != nil {
		t.Fatal(err)
	}
	if req.AuthType != app.EventSourceAuthTypeHMAC || req.AuthConfig.Header != "X-Hub-Signature-256" || req.IDFrom.Header != "X-GitHub-Delivery" {
		t.Fatalf("unexpected GitHub preset: %#v", req)
	}

	pubsub := createEventSourceRequest{Preset: "google-pubsub", AuthConfig: app.EventSourceAuthConfig{Audience: []string{"https://example.com/hook"}}}
	if err := applyEventSourcePreset(&pubsub); err != nil {
		t.Fatal(err)
	}
	if err := defaultAndValidateAuthConfig(&pubsub); err != nil {
		t.Fatal(err)
	}
	if pubsub.Envelope != app.EventEnvelopeTypePubSubPush || pubsub.AuthType != app.EventSourceAuthTypeBearerJWT {
		t.Fatalf("unexpected Pub/Sub preset: %#v", pubsub)
	}
}

func TestValidateOIDCIssuerRejectsLocalNetworks(t *testing.T) {
	for _, issuer := range []string{"http://accounts.example.com", "https://localhost", "https://127.0.0.1", "https://10.0.0.1"} {
		if err := validateOIDCIssuer(issuer); err == nil {
			t.Fatalf("accepted unsafe issuer %q", issuer)
		}
	}
	if err := validateOIDCIssuer("https://accounts.google.com"); err != nil {
		t.Fatal(err)
	}
}

func TestEventHeadersRedactCredentials(t *testing.T) {
	headers := http.Header{
		"Authorization":    {"Bearer token"},
		"X-Webhook-Secret": {"secret"},
		"X-Event":          {"push"},
	}
	redacted := eventHeaders(&app.EventSource{AuthType: app.EventSourceAuthTypeAPIKey, AuthConfig: app.EventSourceAuthConfig{Header: "X-Webhook-Secret"}}, headers)
	if redacted.Get("Authorization") != "" || redacted.Get("X-Webhook-Secret") != "" || redacted.Get("X-Event") != "push" {
		t.Fatalf("unexpected redacted headers: %#v", redacted)
	}
	if headers.Get("Authorization") == "" {
		t.Fatal("input headers were mutated")
	}
}

func TestReadLimitedBody(t *testing.T) {
	if _, err := readLimitedBody(bytes.NewBuffer(make([]byte, maxEventBody))); err != nil {
		t.Fatal(err)
	}
	if _, err := readLimitedBody(strings.NewReader(strings.Repeat("x", maxEventBody+1))); err == nil {
		t.Fatal("oversized body accepted")
	}
}

func TestIngressDerivedOrgContextPopulatesQueueRecords(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Request = nil
	cctx.SetOrgIDGinContext(ctx, "org-test")

	signalContext := queuecctx.FromContext(ctx)
	if signalContext.OrgID != "org-test" {
		t.Fatalf("expected signal org context, got %q", signalContext.OrgID)
	}

	tx := &gorm.DB{Statement: &gorm.Statement{Context: context.Context(ctx)}}
	queue := app.Queue{}
	if err := queue.BeforeCreate(tx); err != nil || queue.OrgID == nil || *queue.OrgID != "org-test" {
		t.Fatalf("expected queue org context, got %#v: %v", queue.OrgID, err)
	}
	queueSignal := app.QueueSignal{}
	if err := queueSignal.BeforeCreate(tx); err != nil || queueSignal.OrgID == nil || *queueSignal.OrgID != "org-test" {
		t.Fatalf("expected queue signal org context, got %#v: %v", queueSignal.OrgID, err)
	}
}
