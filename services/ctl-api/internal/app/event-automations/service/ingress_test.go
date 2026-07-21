package service

import (
	"bytes"
	"context"
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
