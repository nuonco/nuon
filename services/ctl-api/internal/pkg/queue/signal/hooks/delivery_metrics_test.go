package hooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

func TestNotificationDeliveryMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	metrics := newDeliveryMetrics(provider)

	webhook := &WebhookSignalLifecycleHook{httpClient: http.DefaultClient, deliveryMetrics: metrics}
	successWebhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(successWebhook.Close)
	require.NoError(t, webhook.sendWebhook(context.Background(), webhookTarget{URL: successWebhook.URL}, []byte(`{}`)))

	failureWebhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "rejected", http.StatusBadGateway)
	}))
	require.Error(t, webhook.sendWebhook(context.Background(), webhookTarget{URL: failureWebhook.URL}, []byte(`{}`)))
	failureWebhook.Close()
	require.Error(t, webhook.sendWebhook(context.Background(), webhookTarget{URL: failureWebhook.URL}, []byte(`{}`)))

	slackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/chat.update" {
			_, _ = w.Write([]byte(`{"ok":false,"error":"message_not_found"}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"channel":"channel-sensitive","ts":"1"}`))
	}))
	t.Cleanup(slackServer.Close)
	slackHook := &SlackSignalLifecycleHook{
		slackClient:     slackclient.New(slackclient.WithBaseURL(slackServer.URL)),
		deliveryMetrics: metrics,
	}
	_, err := slackHook.postMessage(context.Background(), "token-sensitive", slackclient.PostMessageRequest{Channel: "channel-sensitive", Text: "message-sensitive"})
	require.NoError(t, err)
	_, err = slackHook.updateMessage(context.Background(), "token-sensitive", slackclient.UpdateMessageRequest{Channel: "channel-sensitive", TS: "1", Text: "message-sensitive"})
	require.Error(t, err)

	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	require.Len(t, data.ScopeMetrics, 1)
	require.Len(t, data.ScopeMetrics[0].Metrics, 2)

	wantCounts := map[string]int64{
		metricAttributeKey(metricAttributes(deliveryChannelWebhook, deliveryOperationPost, "success")): 1,
		metricAttributeKey(metricAttributes(deliveryChannelWebhook, deliveryOperationPost, "failure")): 2,
		metricAttributeKey(metricAttributes(deliveryChannelSlack, deliveryOperationPost, "success")):   1,
		metricAttributeKey(metricAttributes(deliveryChannelSlack, deliveryOperationUpdate, "failure")): 1,
	}
	for _, collected := range data.ScopeMetrics[0].Metrics {
		switch collected.Name {
		case "nuon.notification.delivery.attempts":
			points := collected.Data.(metricdata.Sum[int64]).DataPoints
			require.Len(t, points, len(wantCounts))
			for _, point := range points {
				require.Equal(t, wantCounts[point.Attributes.Encoded(attribute.DefaultEncoder())], point.Value)
				require.Equal(t, 3, point.Attributes.Len())
			}
		case "nuon.notification.delivery.duration":
			points := collected.Data.(metricdata.Histogram[float64]).DataPoints
			require.Len(t, points, len(wantCounts))
			for _, point := range points {
				require.EqualValues(t, wantCounts[point.Attributes.Encoded(attribute.DefaultEncoder())], point.Count)
				require.Equal(t, 3, point.Attributes.Len())
			}
		default:
			t.Fatalf("unexpected metric %q", collected.Name)
		}
	}
}

func metricAttributes(channel, operation, outcome string) attribute.Set {
	return attribute.NewSet(
		attribute.String("nuon.notification.channel", channel),
		attribute.String("nuon.notification.operation", operation),
		attribute.String("nuon.notification.outcome", outcome),
	)
}

func metricAttributeKey(attrs attribute.Set) string {
	return attrs.Encoded(attribute.DefaultEncoder())
}
