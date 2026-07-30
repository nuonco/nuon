package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/events/envelope"
	"github.com/nuonco/nuon/pkg/metrics"
	serviceconfig "github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuecctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/signing"
)

func TestDecodeGenericJSONEvent(t *testing.T) {
	headers := http.Header{"X-Nuon-Event-Id": {"delivery-1"}, "X-Nuon-Event-Type": {"push"}, "Content-Type": {"application/json"}}
	trigger := &app.Trigger{Envelope: app.EventEnvelopeTypeNone, AuthType: app.TriggerAuthTypeHMAC, IDFrom: app.EventFieldSelector{Header: "X-Nuon-Event-ID"}, TypeFrom: app.EventFieldSelector{Header: "X-Nuon-Event-Type"}}
	event, err := decodeEvent(trigger, headers, []byte(`{"ref":"main"}`))
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "delivery-1" || event.Type != "push" || string(event.Payload) != `{"ref":"main"}` {
		t.Fatalf("unexpected normalized event: %#v", event)
	}
	if _, err := decodeEvent(trigger, nil, []byte(`not json`)); err == nil {
		t.Fatal("invalid JSON event accepted")
	}
}

func TestDecodeGenericJSONEventUsesBodyDigestFallback(t *testing.T) {
	body := []byte(`{"ref":"main"}`)
	event, err := decodeEvent(&app.Trigger{Envelope: app.EventEnvelopeTypeNone}, nil, body)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	if event.ID != hex.EncodeToString(sum[:]) {
		t.Fatalf("fallback ID = %q", event.ID)
	}
}

func TestDecodeCloudEventEnvelope(t *testing.T) {
	body := []byte(`{"specversion":"1.0","id":"evt-1","source":"urn:test","type":"test.created","data":{"ok":true}}`)
	event, err := decodeEvent(&app.Trigger{Envelope: app.EventEnvelopeTypeCloudEvents}, nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "evt-1" || event.DedupeID != "evt-1" || event.Source != "urn:test" || event.Type != "test.created" || string(event.Payload) != `{"ok":true}` {
		t.Fatalf("unexpected normalized event: %#v", event)
	}
}

func TestDecodeAzureEventGridEvent(t *testing.T) {
	body := []byte(`[{"id":"evt-1","eventType":"Nuon.Proof.Created","eventTime":"2026-07-28T12:00:00Z","subject":"proof","data":{"ok":true}}]`)
	event, err := decodeEvent(&app.Trigger{Preset: "azure-event-grid", Envelope: app.EventEnvelopeTypeNone}, nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "evt-1" || event.Type != "Nuon.Proof.Created" || event.OccurredAt == nil || !strings.Contains(string(event.Payload), `"subject":"proof"`) {
		t.Fatalf("unexpected normalized event: %#v", event)
	}
}

func TestDecodeSlackEvent(t *testing.T) {
	body := []byte(`{"type":"event_callback","event_id":"Ev123","event_time":1785254400,"team_id":"T123","event":{"type":"message","text":"proof"}}`)
	event, err := decodeEvent(&app.Trigger{Preset: "slack-events", Envelope: app.EventEnvelopeTypeNone}, nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "Ev123" || event.Type != "message" || event.OccurredAt == nil || !strings.Contains(string(event.Payload), `"team_id":"T123"`) {
		t.Fatalf("unexpected normalized event: %#v", event)
	}
}

func TestDecodeGenericJSONPayloadSelectors(t *testing.T) {
	trigger := &app.Trigger{
		Envelope: app.EventEnvelopeTypeNone,
		AuthType: app.TriggerAuthTypeAPIKey,
		IDFrom:   app.EventFieldSelector{Payload: "$.delivery.id"},
		TypeFrom: app.EventFieldSelector{Payload: `$['detail-type']`},
	}
	event, err := decodeEvent(trigger, nil, []byte(`{"delivery":{"id":"evt-1"},"detail-type":"image.push"}`))
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "evt-1" || event.Type != "image.push" {
		t.Fatalf("unexpected normalized event: %#v", event)
	}

	trigger.IDFrom.Payload = "$.delivery[*]"
	if err := validateEventFieldSelector(trigger.IDFrom); err == nil {
		t.Fatal("wildcard selector accepted in singular context")
	}
}

func TestDecodePubSubPushEnvelope(t *testing.T) {
	body := []byte(`{"message":{"data":"eyJhY3Rpb24iOiJwdXNoIn0=","messageId":"msg-1","publishTime":"2026-07-22T12:00:00Z"}}`)
	event, err := decodeEvent(&app.Trigger{Envelope: app.EventEnvelopeTypePubSubPush}, nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "msg-1" || string(event.Payload) != `{"action":"push"}` || event.OccurredAt == nil {
		t.Fatalf("unexpected normalized event: %#v", event)
	}
}

func TestDecodeWrappedEnvelopeSelectors(t *testing.T) {
	tests := map[string]struct {
		trigger  *app.Trigger
		headers  http.Header
		body     []byte
		dedupeID string
	}{
		"CloudEvents": {
			trigger:  &app.Trigger{Envelope: app.EventEnvelopeTypeCloudEvents, IDFrom: app.EventFieldSelector{Payload: "$.event_id"}, TypeFrom: app.EventFieldSelector{Header: "X-Event-Type"}},
			headers:  http.Header{"X-Event-Type": {"override.type"}},
			body:     []byte(`{"specversion":"1.0","id":"native-id","source":"urn:test","type":"native.type","data":{"event_id":"override-id"}}`),
			dedupeID: "native-id",
		},
		"Pub/Sub": {
			trigger:  &app.Trigger{Envelope: app.EventEnvelopeTypePubSubPush, IDFrom: app.EventFieldSelector{Payload: "$.event_id"}, TypeFrom: app.EventFieldSelector{Payload: "$.event_type"}},
			body:     []byte(`{"message":{"data":"eyJldmVudF9pZCI6Im92ZXJyaWRlLWlkIiwiZXZlbnRfdHlwZSI6Im92ZXJyaWRlLnR5cGUifQ==","messageId":"native-id"}}`),
			dedupeID: "override-id",
		},
		"SNS": {
			trigger:  &app.Trigger{Envelope: app.EventEnvelopeTypeSNS, IDFrom: app.EventFieldSelector{Payload: "$.event_id"}, TypeFrom: app.EventFieldSelector{Payload: "$.event_type"}},
			body:     []byte(`{"Type":"Notification","MessageId":"native-id","TopicArn":"arn:aws:sns:us-east-1:123456789012:events","Message":"{\"event_id\":\"override-id\",\"event_type\":\"override.type\"}","Timestamp":"2026-07-22T12:00:00Z","SignatureVersion":"1","Signature":"eA==","SigningCertURL":"https://sns.us-east-1.amazonaws.com/SimpleNotificationService-deadbeef.pem"}`),
			dedupeID: "override-id",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			event, err := decodeEvent(tt.trigger, tt.headers, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if event.ID != "override-id" || event.DedupeID != tt.dedupeID || event.Type != "override.type" {
				t.Fatalf("selectors were not applied: %#v", event)
			}
		})
	}
}

func TestMissingWrappedHeaderSelectorsPreserveNativeFields(t *testing.T) {
	tests := map[string]struct {
		envelope app.EventEnvelopeType
		body     []byte
		wantType string
	}{
		"CloudEvents": {
			envelope: app.EventEnvelopeTypeCloudEvents,
			body:     []byte(`{"specversion":"1.0","id":"native-id","source":"urn:test","type":"native.type","data":{"ok":true}}`),
			wantType: "native.type",
		},
		"Pub/Sub": {
			envelope: app.EventEnvelopeTypePubSubPush,
			body:     []byte(`{"message":{"data":"eyJvayI6dHJ1ZX0=","messageId":"native-id"}}`),
		},
		"SNS": {
			envelope: app.EventEnvelopeTypeSNS,
			body:     []byte(`{"Type":"Notification","MessageId":"native-id","TopicArn":"arn:aws:sns:us-east-1:123456789012:events","Message":"{\"ok\":true}","Timestamp":"2026-07-22T12:00:00Z","SignatureVersion":"1","Signature":"eA==","SigningCertURL":"https://sns.us-east-1.amazonaws.com/SimpleNotificationService-deadbeef.pem"}`),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			trigger := &app.Trigger{Envelope: tt.envelope, IDFrom: app.EventFieldSelector{Header: "X-Event-ID"}, TypeFrom: app.EventFieldSelector{Header: "X-Event-Type"}}
			event, err := decodeEvent(trigger, nil, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if event.ID != "native-id" || event.Type != tt.wantType {
				t.Fatalf("missing selectors erased native fields: %#v", event)
			}
		})
	}
}

func TestTriggerPresets(t *testing.T) {
	req := createTriggerRequest{Preset: "github"}
	if err := applyTriggerPreset(&req); err != nil {
		t.Fatal(err)
	}
	if err := defaultAndValidateAuthConfig(&req); err != nil {
		t.Fatal(err)
	}
	if req.AuthType != app.TriggerAuthTypeHMAC || req.AuthConfig.Header != "X-Hub-Signature-256" || req.IDFrom.Header != "X-GitHub-Delivery" {
		t.Fatalf("unexpected GitHub preset: %#v", req)
	}

	pubsub := createTriggerRequest{Preset: "google-pubsub", AuthConfig: app.TriggerAuthConfig{Audience: []string{"https://example.com/hook"}, ExpectedEmail: "push@example.iam.gserviceaccount.com"}}
	if err := applyTriggerPreset(&pubsub); err != nil {
		t.Fatal(err)
	}
	if err := defaultAndValidateAuthConfig(&pubsub); err != nil {
		t.Fatal(err)
	}
	if pubsub.Envelope != app.EventEnvelopeTypePubSubPush || pubsub.AuthType != app.TriggerAuthTypeBearerJWT {
		t.Fatalf("unexpected Pub/Sub preset: %#v", pubsub)
	}
	eventbridge := createTriggerRequest{Preset: "aws-eventbridge"}
	if err := applyTriggerPreset(&eventbridge); err != nil {
		t.Fatal(err)
	}
	if eventbridge.TypeFrom.Payload != `$['detail-type']` || eventbridge.IDFrom.Payload != "$.id" {
		t.Fatalf("unexpected EventBridge selectors: %#v", eventbridge)
	}
	eventGrid := createTriggerRequest{Preset: "azure-event-grid"}
	if err := applyTriggerPreset(&eventGrid); err != nil {
		t.Fatal(err)
	}
	if eventGrid.Envelope == "" {
		eventGrid.Envelope = app.EventEnvelopeTypeNone
	}
	if err := defaultAndValidateAuthConfig(&eventGrid); err != nil {
		t.Fatal(err)
	}
	if eventGrid.AuthType != app.TriggerAuthTypeAPIKey || eventGrid.AuthConfig.Header != "X-Nuon-API-Key" || eventGrid.Envelope != app.EventEnvelopeTypeNone {
		t.Fatalf("unexpected Azure Event Grid preset: %#v", eventGrid)
	}
	datadog := createTriggerRequest{Preset: "datadog"}
	if err := applyTriggerPreset(&datadog); err != nil {
		t.Fatal(err)
	}
	if datadog.AuthType != app.TriggerAuthTypeAPIKey || datadog.AuthConfig.Header != "X-Nuon-API-Key" || datadog.Envelope != app.EventEnvelopeTypeNone || datadog.IDFrom.Payload != "$.event_id" || datadog.TypeFrom.Payload != "$.event_type" {
		t.Fatalf("unexpected Datadog preset: %#v", datadog)
	}
	slackEvents := createTriggerRequest{Preset: "slack-events"}
	if err := applyTriggerPreset(&slackEvents); err != nil {
		t.Fatal(err)
	}
	if slackEvents.AuthType != app.TriggerAuthTypeHMAC || slackEvents.AuthConfig.Header != "X-Slack-Signature" || slackEvents.AuthConfig.Prefix != "v0=" || slackEvents.Envelope != app.EventEnvelopeTypeNone {
		t.Fatalf("unexpected Slack Events preset: %#v", slackEvents)
	}
	if slackEvents.TypeFrom != (app.EventFieldSelector{}) || slackEvents.IDFrom != (app.EventFieldSelector{}) {
		t.Fatalf("Slack preset advertised unused selectors: %#v", slackEvents)
	}
	for _, invalid := range []createTriggerRequest{
		{Preset: "azure-event-grid", AuthType: app.TriggerAuthTypeNone},
		{Preset: "azure-event-grid", Envelope: app.EventEnvelopeTypeCloudEvents},
		{Preset: "azure-event-grid", AuthConfig: app.TriggerAuthConfig{Header: "X-Other-Key"}},
		{Preset: "slack-events", AuthType: app.TriggerAuthTypeAPIKey},
		{Preset: "slack-events", Envelope: app.EventEnvelopeTypeCloudEvents},
		{Preset: "slack-events", AuthConfig: app.TriggerAuthConfig{Header: "X-Other-Signature"}},
	} {
		if err := applyTriggerPreset(&invalid); err == nil {
			t.Fatalf("accepted conflicting Azure Event Grid preset: %#v", invalid)
		}
	}
}

func TestTriggerPresetsRejectConflictingFixedFields(t *testing.T) {
	presets := []string{
		"github", "gitlab", "bitbucket", "gitea", "forgejo", "terraform-cloud",
		"google-pubsub", "azure-devops", "aws-eventbridge", "aws-sns", "azure-event-grid", "slack-events",
	}
	for _, preset := range presets {
		t.Run(preset, func(t *testing.T) {
			requests := []createTriggerRequest{
				{Preset: preset, AuthType: app.TriggerAuthTypeNone},
				{Preset: preset, Envelope: app.EventEnvelopeTypeCloudEvents},
				{Preset: preset, AuthConfig: app.TriggerAuthConfig{Header: "X-Conflicting-Auth"}},
				{Preset: preset, TypeFrom: app.EventFieldSelector{Header: "X-Conflicting-Type"}},
				{Preset: preset, IDFrom: app.EventFieldSelector{Header: "X-Conflicting-ID"}},
			}
			for _, req := range requests {
				if err := applyTriggerPreset(&req); err == nil {
					t.Fatalf("accepted conflicting preset request: %#v", req)
				}
			}
		})
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

func TestEventJWTClaimsRequireVerifiedExpectedEmail(t *testing.T) {
	claims := &eventJWTClaims{Email: "push@example.iam.gserviceaccount.com", EmailVerified: true, expectedEmail: "push@example.iam.gserviceaccount.com"}
	if err := claims.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	claims.EmailVerified = false
	if err := claims.Validate(context.Background()); err == nil {
		t.Fatal("accepted unverified email")
	}
}

func TestEventHeadersRedactCredentials(t *testing.T) {
	headers := http.Header{
		"Authorization":    {"Bearer token"},
		"Cookie":           {"X-Nuon-Auth=session-token"},
		"Set-Cookie":       {"session=response-token"},
		"X-Nuon-Signature": {"v1=signature"},
		"X-Webhook-Secret": {"secret"},
		"X-Event":          {"push"},
	}
	redacted := eventHeaders(&app.Trigger{AuthType: app.TriggerAuthTypeAPIKey, AuthConfig: app.TriggerAuthConfig{Header: "X-Webhook-Secret"}}, headers)
	if redacted.Get("Authorization") != "" || redacted.Get("Cookie") != "" || redacted.Get("Set-Cookie") != "" || redacted.Get("X-Webhook-Secret") != "" || redacted.Get("X-Event") != "push" {
		t.Fatalf("unexpected redacted headers: %#v", redacted)
	}
	if headers.Get("Authorization") == "" {
		t.Fatal("input headers were mutated")
	}
	redacted = eventHeaders(&app.Trigger{AuthType: app.TriggerAuthTypeHMAC}, headers)
	if redacted.Get("X-Nuon-Signature") != "" || redacted.Get("X-Event") != "push" {
		t.Fatalf("unexpected default HMAC redacted headers: %#v", redacted)
	}
}

func TestReadLimitedBody(t *testing.T) {
	if maxEventBody != 16*1024*1024 {
		t.Fatalf("maxEventBody = %d; want 16 MiB", maxEventBody)
	}
	if _, err := readLimitedBody(bytes.NewBuffer(make([]byte, maxEventBody))); err != nil {
		t.Fatal(err)
	}
	if _, err := readLimitedBody(strings.NewReader(strings.Repeat("x", maxEventBody+1))); err == nil {
		t.Fatal("oversized body accepted")
	}
}

func TestActiveTriggerSecret(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Minute)
	future := now.Add(time.Minute)
	tests := map[string]struct {
		secret app.TriggerSecret
		active bool
	}{
		"current": {secret: app.TriggerSecret{NotBefore: past}, active: true},
		"grace":   {secret: app.TriggerSecret{NotBefore: past, ExpiresAt: &future}, active: true},
		"future":  {secret: app.TriggerSecret{NotBefore: future}},
		"expired": {secret: app.TriggerSecret{NotBefore: past, ExpiresAt: &past}},
		"revoked": {secret: app.TriggerSecret{NotBefore: past, RevokedAt: &past}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := activeTriggerSecret(&tt.secret, now); got != tt.active {
				t.Fatalf("activeTriggerSecret() = %t, want %t", got, tt.active)
			}
		})
	}
}

func TestTriggerMetadataResponseCapsSecretMetadata(t *testing.T) {
	secrets := make([]app.TriggerSecret, maxEventSecretMetadata+5)
	for i := range secrets {
		secrets[i].ID = fmt.Sprintf("secret-%d", i)
	}
	response := triggerMetadataResponse(&app.Trigger{Secrets: secrets})
	if len(response.Secrets) != maxEventSecretMetadata {
		t.Fatalf("returned %d secrets, want %d", len(response.Secrets), maxEventSecretMetadata)
	}
	if response.Secrets[0].ID != "secret-0" || response.Secrets[maxEventSecretMetadata-1].ID != "secret-19" {
		t.Fatalf("unexpected metadata range: %q through %q", response.Secrets[0].ID, response.Secrets[maxEventSecretMetadata-1].ID)
	}
}

func TestTriggerMetadataResponseOmitsIngressURL(t *testing.T) {
	trigger := &app.Trigger{IngressKey: "opaque-key"}
	payload, err := json.Marshal(triggerMetadataResponse(trigger))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "opaque-key") || strings.Contains(string(payload), "ingress_url") {
		t.Fatalf("metadata response leaked ingress credentials: %s", payload)
	}
}

func TestBuildIngressURL(t *testing.T) {
	if url := buildIngressURL("https://api.example.com", "opaque-key"); url != "https://api.example.com/v1/event-ingress/opaque-key" {
		t.Fatalf("unexpected ingress URL %q", url)
	}
	if url := buildIngressURL("https://api.example.com", ""); url != "" {
		t.Fatalf("empty ingress key produced URL %q", url)
	}
}

func TestIngressDerivedOrgContextPopulatesQueueRecords(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Request = nil
	cctx.SetAccountIDGinContext(ctx, "acct-test")
	cctx.SetOrgIDGinContext(ctx, "org-test")

	signalContext := queuecctx.FromContext(ctx)
	if signalContext.AccountID != "acct-test" || signalContext.OrgID != "org-test" {
		t.Fatalf("unexpected signal context: %#v", signalContext)
	}
	if _, err := cctx.AccountFromGinContext(ctx); err == nil {
		t.Fatal("account ID attribution installed an authenticated account")
	}

	tx := &gorm.DB{Statement: &gorm.Statement{Context: context.Context(ctx)}}
	queue := app.Queue{}
	if err := queue.BeforeCreate(tx); err != nil || queue.OrgID == nil || *queue.OrgID != "org-test" {
		t.Fatalf("expected queue org context, got %#v: %v", queue.OrgID, err)
	}
	queueSignal := app.QueueSignal{}
	if err := queueSignal.BeforeCreate(tx); err != nil || queueSignal.CreatedByID != "acct-test" || queueSignal.OrgID == nil || *queueSignal.OrgID != "org-test" {
		t.Fatalf("unexpected queue signal attribution: account=%q org=%#v: %v", queueSignal.CreatedByID, queueSignal.OrgID, err)
	}
}

type ingressPersistenceDBConfig struct {
	Host     string `config:"db_host"`
	Port     string `config:"db_port"`
	User     string `config:"db_user"`
	Password string `config:"db_password"`
	SSLMode  string `config:"db_ssl_mode"`
	Name     string `config:"db_name"`
}

type ingressPersistenceTestSuite struct {
	suite.Suite

	db         *gorm.DB
	service    *service
	router     *gin.Engine
	trigger    *app.Trigger
	account    *app.Account
	ingressKey string
}

type legacyPayloadHashEvent struct {
	ID      string `gorm:"primaryKey"`
	Payload string
}

type migratedPayloadHashEvent struct {
	ID            string `gorm:"primaryKey"`
	Payload       string
	PayloadSHA256 string `gorm:"<-:create"`
}

func TestIngressPersistence(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
	}
	suite.Run(t, new(ingressPersistenceTestSuite))
}

func (s *ingressPersistenceTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	var cfg ingressPersistenceDBConfig
	require.NoError(s.T(), serviceconfig.LoadInto(nil, &cfg))
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "5432"
	}
	if cfg.User == "" {
		cfg.User = "ctl_api"
	}
	if cfg.Password == "" {
		cfg.Password = "password"
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.Name == "" {
		cfg.Name = "ctl_api_test"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn))
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.AutoMigrate(&app.TriggerEvent{}))
	s.db = db

	v := validator.New()
	mw, err := metrics.New(v, metrics.WithDisable(true), metrics.WithLogger(zap.NewNop()))
	require.NoError(s.T(), err)
	client := queueclient.New(queueclient.Params{DB: db, Cfg: &internal.Config{}, L: zap.NewNop(), MW: mw})
	s.service = &service{db: db, l: zap.NewNop(), queueClient: client, features: features.New(features.Params{DB: db, Cfg: &internal.Config{}, V: v})}
}

func (s *ingressPersistenceTestSuite) SetupTest() {
	account := &app.Account{ID: domains.NewAccountID(), Subject: domains.NewAccountID(), Email: domains.NewAccountID() + "@test.nuon.co", AccountType: app.AccountTypeAuth0}
	require.NoError(s.T(), s.db.Create(account).Error)
	ctx := cctx.SetAccountIDContext(context.Background(), account.ID)
	org := &app.Org{ID: domains.NewOrgID(), Name: "event-ingress-" + domains.NewOrgID(), OrgType: app.OrgTypeSandbox, Status: app.OrgStatusActive, SandboxMode: true, CreatedByID: account.ID, Features: map[string]bool{string(app.OrgFeatureTriggers): true}}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(org).Error)
	ctx = cctx.SetOrgIDContext(ctx, org.ID)
	s.account = account
	s.ingressKey = "test-ingress-key-" + org.ID
	s.trigger = &app.Trigger{
		OrgID:          org.ID,
		CreatedByID:    account.ID,
		IngressKeyHash: hashIngressKey(s.ingressKey),
		Name:           "registry-" + org.ID,
		AuthType:       app.TriggerAuthTypeNone,
		Envelope:       app.EventEnvelopeTypeNone,
		IDFrom:         app.EventFieldSelector{Header: "X-Event-ID"},
		Status:         app.TriggerStatusActive,
	}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(s.trigger).Error)

	s.router = gin.New()
	s.router.POST("/v1/event-ingress/:ingress_key", s.service.IngestEvent)
}

func (s *ingressPersistenceTestSuite) request(body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/event-ingress/"+s.ingressKey, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-ID", "external-event-1")
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)
	return recorder
}

func (s *ingressPersistenceTestSuite) eventCount(triggerID string) int64 {
	var count int64
	require.NoError(s.T(), s.db.Model(&app.TriggerEvent{}).Where(app.TriggerEvent{TriggerID: triggerID}).Count(&count).Error)
	return count
}

func (s *ingressPersistenceTestSuite) createAuthenticatedTrigger(authType app.TriggerAuthType, authConfig app.TriggerAuthConfig, secret string) (*app.Trigger, string) {
	ingressKey := "authenticated-" + domains.NewOrgID()
	trigger := &app.Trigger{
		OrgID:          s.trigger.OrgID,
		CreatedByID:    s.account.ID,
		IngressKeyHash: hashIngressKey(ingressKey),
		Name:           "authenticated-" + domains.NewOrgID(),
		AuthType:       authType,
		AuthConfig:     authConfig,
		Envelope:       app.EventEnvelopeTypeNone,
		Status:         app.TriggerStatusActive,
	}
	require.NoError(s.T(), s.db.Create(trigger).Error)
	if secret != "" {
		require.NoError(s.T(), s.db.Create(&app.TriggerSecret{
			CreatedByID: s.account.ID,
			OrgID:       s.trigger.OrgID,
			TriggerID:   trigger.ID,
			Secret:      secret,
			NotBefore:   time.Now().Add(-time.Minute),
		}).Error)
	}
	return trigger, ingressKey
}

func (s *ingressPersistenceTestSuite) TestAuthenticationFailuresDoNotPersistEvents() {
	tests := map[string]struct {
		authType   app.TriggerAuthType
		authConfig app.TriggerAuthConfig
		secret     string
		body       string
		headers    http.Header
	}{
		"api key": {
			authType:   app.TriggerAuthTypeAPIKey,
			authConfig: app.TriggerAuthConfig{Header: "X-Nuon-API-Key"},
			secret:     "correct-api-key",
			body:       `{"ok":true}`,
			headers:    http.Header{"X-Nuon-Api-Key": {"wrong-api-key"}},
		},
		"hmac": {
			authType:   app.TriggerAuthTypeHMAC,
			authConfig: app.TriggerAuthConfig{Header: "X-Hub-Signature-256", Prefix: "sha256=", Algorithm: "sha256", Encoding: "hex"},
			secret:     "correct-hmac-secret",
			body:       `{"ok":true}`,
			headers:    http.Header{"X-Hub-Signature-256": {"sha256=0000000000000000000000000000000000000000000000000000000000000000"}},
		},
		"sns": {
			authType:   app.TriggerAuthTypeSNSSignature,
			authConfig: app.TriggerAuthConfig{TopicARN: "arn:aws:sns:us-east-1:123456789012:events"},
			body:       `{"Type":"Notification"}`,
		},
	}
	for name, tt := range tests {
		s.T().Run(name, func(t *testing.T) {
			trigger, ingressKey := s.createAuthenticatedTrigger(tt.authType, tt.authConfig, tt.secret)
			req := httptest.NewRequest(http.MethodPost, "/v1/event-ingress/"+ingressKey, strings.NewReader(tt.body))
			if tt.headers != nil {
				req.Header = tt.headers.Clone()
			}
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			s.router.ServeHTTP(recorder, req)
			require.Equal(t, http.StatusUnauthorized, recorder.Code)
			require.Zero(t, s.eventCount(trigger.ID))
		})
	}
}

func (s *ingressPersistenceTestSuite) TestOversizedRequestDoesNotPersistEvent() {
	req := httptest.NewRequest(http.MethodPost, "/v1/event-ingress/"+s.ingressKey, bytes.NewReader(make([]byte, maxEventBody+1)))
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)
	require.Equal(s.T(), http.StatusRequestEntityTooLarge, recorder.Code)
	require.Zero(s.T(), s.eventCount(s.trigger.ID))
}

func (s *ingressPersistenceTestSuite) TestRejectedEventRechecksTriggerState() {
	tests := map[string]func(){
		"replaced ingress key": func() {
			require.NoError(s.T(), s.db.Model(&app.Trigger{}).Where(app.Trigger{ID: s.trigger.ID}).Update("ingress_key_hash", hashIngressKey("replacement")).Error)
		},
		"suspended trigger": func() {
			require.NoError(s.T(), s.db.Model(&app.Trigger{}).Where(app.Trigger{ID: s.trigger.ID}).Update("status", app.TriggerStatusSuspended).Error)
		},
	}
	for name, changeTrigger := range tests {
		s.T().Run(name, func(t *testing.T) {
			staleTrigger := *s.trigger
			changeTrigger()
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-json"))
			err := s.service.persistRejectedEvent(ctx, &staleTrigger, []byte("not-json"), time.Now(), "invalid event")
			require.ErrorIs(t, err, errTriggerInactive)
			require.Zero(t, s.eventCount(s.trigger.ID))
		})
	}
}

func (s *ingressPersistenceTestSuite) TestDeleteTriggerRejectsActiveWaiter() {
	require.NoError(s.T(), s.db.Create(&app.EventRunbookWaiter{
		OrgID:          s.trigger.OrgID,
		AppID:          domains.NewAppID(),
		InstallID:      domains.NewInstallID(),
		WorkflowID:     domains.NewWorkflowID(),
		WorkflowStepID: domains.NewWorkflowStepID(),
		QueueSignalID:  domains.NewQueueSignalID(),
		TriggerID:      s.trigger.ID,
		Status:         app.EventRunbookWaiterStatusActive,
		ActivatedAt:    time.Now(),
	}).Error)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/v1/triggers/"+s.trigger.ID+"?force=true", nil)
	ctx.Params = gin.Params{{Key: "trigger_id", Value: s.trigger.ID}}
	cctx.SetOrgGinContext(ctx, &app.Org{ID: s.trigger.OrgID})
	s.service.DeleteTrigger(ctx)
	require.Equal(s.T(), http.StatusConflict, ctx.Writer.Status())
	var trigger app.Trigger
	require.NoError(s.T(), s.db.Where(app.Trigger{ID: s.trigger.ID}).First(&trigger).Error)
}

func (s *ingressPersistenceTestSuite) TestIngressKeyReplacementRejectsInFlightEvent() {
	staleTrigger := *s.trigger
	require.NoError(s.T(), s.db.Model(&app.Trigger{}).
		Where(app.Trigger{ID: s.trigger.ID}).
		Update("ingress_key_hash", hashIngressKey("replaced-"+s.ingressKey)).Error)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/event-ingress/"+s.ingressKey, strings.NewReader(`{"tag":"one"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	cctx.SetAccountIDGinContext(ctx, s.account.ID)
	cctx.SetOrgIDGinContext(ctx, s.trigger.OrgID)

	queue, err := s.service.orgTriggerQueue(ctx, s.trigger.OrgID)
	require.NoError(s.T(), err)

	event := &envelope.Event{ID: "in-flight-1", Payload: json.RawMessage(`{"tag":"one"}`), ContentType: "application/json"}
	_, _, _, err = s.service.persistEvent(ctx, &staleTrigger, nil, queue.ID, event, []byte(`{"tag":"one"}`), time.Now())
	require.ErrorIs(s.T(), err, errTriggerInactive)

	var events int64
	require.NoError(s.T(), s.db.Model(&app.TriggerEvent{}).Where(app.TriggerEvent{TriggerID: s.trigger.ID}).Count(&events).Error)
	require.Zero(s.T(), events)
}

func (s *ingressPersistenceTestSuite) TestMalformedEventPersistsRejectedLedgerEntry() {
	recorder := s.request("not-json")
	require.Equal(s.T(), http.StatusAccepted, recorder.Code)

	var event app.TriggerEvent
	require.NoError(s.T(), s.db.Where(app.TriggerEvent{TriggerID: s.trigger.ID}).First(&event).Error)
	require.Equal(s.T(), app.EventRoutingStatusRejected, event.RoutingStatus)
	require.JSONEq(s.T(), `{}`, string(event.Payload))
	require.Equal(s.T(), "envelope decoding failed: invalid JSON event", event.RoutingError)
}

func (s *ingressPersistenceTestSuite) TestAzureEventGridValidationHandshake() {
	ingressKey := "azure-event-grid-" + s.trigger.OrgID
	trigger := &app.Trigger{
		OrgID:          s.trigger.OrgID,
		CreatedByID:    s.account.ID,
		IngressKeyHash: hashIngressKey(ingressKey),
		Name:           "azure-event-grid-" + s.trigger.OrgID,
		Preset:         "azure-event-grid",
		AuthType:       app.TriggerAuthTypeAPIKey,
		AuthConfig:     app.TriggerAuthConfig{Header: "X-Nuon-API-Key"},
		Envelope:       app.EventEnvelopeTypeNone,
		Status:         app.TriggerStatusActive,
	}
	require.NoError(s.T(), s.db.Create(trigger).Error)
	require.NoError(s.T(), s.db.Create(&app.TriggerSecret{
		CreatedByID: s.account.ID,
		OrgID:       s.trigger.OrgID,
		TriggerID:   trigger.ID,
		Secret:      "azure-secret",
		NotBefore:   time.Now().Add(-time.Minute),
	}).Error)

	body := `[{"id":"validation-1","eventType":"Microsoft.EventGrid.SubscriptionValidationEvent","data":{"validationCode":"validation-code"}}]`
	req := httptest.NewRequest(http.MethodPost, "/v1/event-ingress/"+ingressKey, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Nuon-API-Key", "azure-secret")
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)

	require.Equal(s.T(), http.StatusOK, recorder.Code)
	require.JSONEq(s.T(), `{"validationResponse":"validation-code"}`, recorder.Body.String())
	var events int64
	require.NoError(s.T(), s.db.Model(&app.TriggerEvent{}).Where(app.TriggerEvent{TriggerID: trigger.ID}).Count(&events).Error)
	require.Zero(s.T(), events)
}

func (s *ingressPersistenceTestSuite) TestAzureEventGridRejectsBatches() {
	ingressKey := "azure-event-grid-batch-" + s.trigger.OrgID
	trigger := &app.Trigger{
		OrgID:          s.trigger.OrgID,
		CreatedByID:    s.account.ID,
		IngressKeyHash: hashIngressKey(ingressKey),
		Name:           "azure-event-grid-batch-" + s.trigger.OrgID,
		Preset:         "azure-event-grid",
		AuthType:       app.TriggerAuthTypeAPIKey,
		AuthConfig:     app.TriggerAuthConfig{Header: "X-Nuon-API-Key"},
		Envelope:       app.EventEnvelopeTypeNone,
		Status:         app.TriggerStatusActive,
	}
	require.NoError(s.T(), s.db.Create(trigger).Error)
	require.NoError(s.T(), s.db.Create(&app.TriggerSecret{CreatedByID: s.account.ID, OrgID: s.trigger.OrgID, TriggerID: trigger.ID, Secret: "azure-secret", NotBefore: time.Now().Add(-time.Minute)}).Error)
	body := `[{"id":"evt-1","eventType":"Nuon.Proof.Created","eventTime":"2026-07-28T12:00:00Z","data":{}},{"id":"evt-2","eventType":"Nuon.Proof.Created","eventTime":"2026-07-28T12:00:00Z","data":{}}]`
	req := httptest.NewRequest(http.MethodPost, "/v1/event-ingress/"+ingressKey, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Nuon-API-Key", "azure-secret")
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)
	require.Equal(s.T(), http.StatusBadRequest, recorder.Code)

	var event app.TriggerEvent
	require.NoError(s.T(), s.db.Where(app.TriggerEvent{TriggerID: trigger.ID}).First(&event).Error)
	require.Equal(s.T(), app.EventRoutingStatusRejected, event.RoutingStatus)
}

func slackTestSignature(secret, timestamp, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = fmt.Fprintf(mac, "v0:%s:%s", timestamp, body)
	return "v0=" + hex.EncodeToString(mac.Sum(nil))
}

func (s *ingressPersistenceTestSuite) TestSlackEventsHandshakeAndDelivery() {
	ingressKey := "slack-events-" + s.trigger.OrgID
	trigger := &app.Trigger{
		OrgID:          s.trigger.OrgID,
		CreatedByID:    s.account.ID,
		IngressKeyHash: hashIngressKey(ingressKey),
		Name:           "slack-events-" + s.trigger.OrgID,
		Preset:         "slack-events",
		AuthType:       app.TriggerAuthTypeHMAC,
		AuthConfig:     app.TriggerAuthConfig{Header: signing.SignatureHeader, Prefix: "v0=", Algorithm: "sha256", Encoding: "hex"},
		Envelope:       app.EventEnvelopeTypeNone,
		Status:         app.TriggerStatusActive,
	}
	require.NoError(s.T(), s.db.Create(trigger).Error)
	const secret = "slack-signing-secret"
	require.NoError(s.T(), s.db.Create(&app.TriggerSecret{CreatedByID: s.account.ID, OrgID: s.trigger.OrgID, TriggerID: trigger.ID, Secret: secret, NotBefore: time.Now().Add(-time.Minute)}).Error)

	send := func(body string) *httptest.ResponseRecorder {
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		req := httptest.NewRequest(http.MethodPost, "/v1/event-ingress/"+ingressKey, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(signing.TimestampHeader, timestamp)
		req.Header.Set(signing.SignatureHeader, slackTestSignature(secret, timestamp, body))
		recorder := httptest.NewRecorder()
		s.router.ServeHTTP(recorder, req)
		return recorder
	}

	validation := send(`{"type":"url_verification","challenge":"slack-proof"}`)
	require.Equal(s.T(), http.StatusOK, validation.Code)
	require.JSONEq(s.T(), `{"challenge":"slack-proof"}`, validation.Body.String())
	var eventCount int64
	require.NoError(s.T(), s.db.Model(&app.TriggerEvent{}).Where(app.TriggerEvent{TriggerID: trigger.ID}).Count(&eventCount).Error)
	require.Zero(s.T(), eventCount)

	invalidBody := `{"type":"event_callback","event_id":"Ev-forged","event_time":1785254400,"event":{"type":"message"}}`
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	invalidRequest := httptest.NewRequest(http.MethodPost, "/v1/event-ingress/"+ingressKey, strings.NewReader(invalidBody))
	invalidRequest.Header.Set(signing.TimestampHeader, timestamp)
	invalidRequest.Header.Set(signing.SignatureHeader, slackTestSignature("wrong-secret", timestamp, invalidBody))
	invalidRecorder := httptest.NewRecorder()
	s.router.ServeHTTP(invalidRecorder, invalidRequest)
	require.Equal(s.T(), http.StatusUnauthorized, invalidRecorder.Code)
	require.NoError(s.T(), s.db.Model(&app.TriggerEvent{}).Where(app.TriggerEvent{TriggerID: trigger.ID}).Count(&eventCount).Error)
	require.Zero(s.T(), eventCount)

	deliveryBody := `{"type":"event_callback","team_id":"T123","api_app_id":"A123","event_id":"Ev-proof-1","event_time":1785254400,"event":{"type":"message","text":"Nuon proof"}}`
	delivery := send(deliveryBody)
	require.Equal(s.T(), http.StatusAccepted, delivery.Code)

	var event app.TriggerEvent
	require.NoError(s.T(), s.db.Where(app.TriggerEvent{TriggerID: trigger.ID, ExternalID: "Ev-proof-1"}).First(&event).Error)
	require.Equal(s.T(), "message", event.EventType)
	require.JSONEq(s.T(), deliveryBody, string(event.Payload))

	duplicate := send(deliveryBody)
	require.Equal(s.T(), http.StatusAccepted, duplicate.Code)
	require.Contains(s.T(), duplicate.Body.String(), `"duplicate":true`)
	require.NoError(s.T(), s.db.Model(&app.TriggerEvent{}).Where(app.TriggerEvent{TriggerID: trigger.ID, ExternalID: "Ev-proof-1"}).Count(&eventCount).Error)
	require.Equal(s.T(), int64(1), eventCount)
}

func (s *ingressPersistenceTestSuite) TestPayloadHashColumnRolloutIsNullable() {
	table := "event_payload_hash_migration_" + strings.ToLower(domains.NewOrgID())
	require.NoError(s.T(), s.db.Table(table).AutoMigrate(&legacyPayloadHashEvent{}))
	s.T().Cleanup(func() { _ = s.db.Migrator().DropTable(table) })
	require.NoError(s.T(), s.db.Table(table).Create(map[string]any{"id": "legacy", "payload": `{"tag":"one"}`}).Error)

	require.NoError(s.T(), s.db.Table(table).AutoMigrate(&migratedPayloadHashEvent{}))
	var nullable string
	require.NoError(s.T(), s.db.Raw(`SELECT is_nullable FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ? AND column_name = 'payload_sha256'`, table).Scan(&nullable).Error)
	require.Equal(s.T(), "YES", nullable)

	var payloadHash *string
	require.NoError(s.T(), s.db.Table(table).Select("payload_sha256").Where(map[string]any{"id": "legacy"}).Scan(&payloadHash).Error)
	require.Nil(s.T(), payloadHash)
}

func (s *ingressPersistenceTestSuite) TestDuplicateDeliveryAndCollision() {
	first := s.request(`{"tag":"one"}`)
	require.Equal(s.T(), http.StatusAccepted, first.Code)
	require.NoError(s.T(), s.db.Table(plugins.TableName(s.db, &app.TriggerEvent{})).
		Where(map[string]any{"trigger_id": s.trigger.ID, "external_id": "external-event-1"}).
		Update("payload_sha256", nil).Error)
	var missingHash int64
	require.NoError(s.T(), s.db.Table(plugins.TableName(s.db, &app.TriggerEvent{})).
		Where(map[string]any{"trigger_id": s.trigger.ID, "external_id": "external-event-1", "payload_sha256": nil}).
		Count(&missingHash).Error)
	require.Equal(s.T(), int64(1), missingHash)

	duplicate := s.request(`{"tag":"one"}`)
	require.Equal(s.T(), http.StatusAccepted, duplicate.Code)
	require.Contains(s.T(), duplicate.Body.String(), `"duplicate":true`)
	var persisted app.TriggerEvent
	require.NoError(s.T(), s.db.Where(app.TriggerEvent{TriggerID: s.trigger.ID, ExternalID: "external-event-1"}).First(&persisted).Error)
	require.NotEmpty(s.T(), persisted.PayloadSHA256)
	require.Equal(s.T(), "external-event-1", persisted.DedupeID)

	collision := s.request(`{"tag":"two"}`)
	require.Equal(s.T(), http.StatusConflict, collision.Code)

	var events int64
	require.NoError(s.T(), s.db.Model(&app.TriggerEvent{}).Where(app.TriggerEvent{TriggerID: s.trigger.ID, ExternalID: "external-event-1"}).Count(&events).Error)
	require.Equal(s.T(), int64(1), events)

	var signal app.QueueSignal
	require.NoError(s.T(), s.db.Where(app.QueueSignal{CreatedByID: s.account.ID}).First(&signal).Error)
	require.Equal(s.T(), s.trigger.OrgID, *signal.OrgID)
}

func (s *ingressPersistenceTestSuite) TestCloudEventDedupeUsesNativeSourceAndID() {
	ingressKey := "cloud-events-" + domains.NewOrgID()
	trigger := &app.Trigger{
		OrgID:          s.trigger.OrgID,
		CreatedByID:    s.account.ID,
		IngressKeyHash: hashIngressKey(ingressKey),
		Name:           "cloud-events-" + domains.NewOrgID(),
		AuthType:       app.TriggerAuthTypeNone,
		Envelope:       app.EventEnvelopeTypeCloudEvents,
		IDFrom:         app.EventFieldSelector{Payload: "$.event_id"},
		Status:         app.TriggerStatusActive,
	}
	require.NoError(s.T(), s.db.Create(trigger).Error)

	send := func(source string) *httptest.ResponseRecorder {
		body := fmt.Sprintf(`{"specversion":"1.0","id":"native-id","source":%q,"type":"proof.created","data":{"event_id":"display-id"}}`, source)
		req := httptest.NewRequest(http.MethodPost, "/v1/event-ingress/"+ingressKey, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/cloudevents+json")
		recorder := httptest.NewRecorder()
		s.router.ServeHTTP(recorder, req)
		return recorder
	}

	first := send("urn:provider:first")
	require.Equal(s.T(), http.StatusAccepted, first.Code)
	require.NotContains(s.T(), first.Body.String(), `"duplicate":true`)

	duplicate := send("urn:provider:first")
	require.Equal(s.T(), http.StatusAccepted, duplicate.Code)
	require.Contains(s.T(), duplicate.Body.String(), `"duplicate":true`)

	secondSource := send("urn:provider:second")
	require.Equal(s.T(), http.StatusAccepted, secondSource.Code)
	require.NotContains(s.T(), secondSource.Body.String(), `"duplicate":true`)

	var events []app.TriggerEvent
	require.NoError(s.T(), s.db.Where(app.TriggerEvent{TriggerID: trigger.ID}).Order("source ASC").Find(&events).Error)
	require.Len(s.T(), events, 2)
	for _, event := range events {
		require.Equal(s.T(), "display-id", event.ExternalID)
		require.Equal(s.T(), "native-id", event.DedupeID)
	}
	require.Equal(s.T(), "urn:provider:first", events[0].Source)
	require.Equal(s.T(), "urn:provider:second", events[1].Source)
}
