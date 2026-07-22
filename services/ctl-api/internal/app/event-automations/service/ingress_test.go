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
