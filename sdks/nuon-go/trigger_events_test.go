package nuon

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestTriggerEventRequests(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" || r.Header.Get("X-Nuon-Org-ID") != "org-id" {
			t.Errorf("missing auth headers: %#v", r.Header)
		}
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		switch r.URL.Path {
		case "/v1/triggers/trigger-1/events":
			fmt.Fprint(w, `{"items":[{"id":"evt-1","dispatches":[{"id":"dispatch-1","status":"triggered"}]}],"next_cursor":"next"}`)
		case "/v1/triggers/events/evt-1":
			fmt.Fprint(w, `{"id":"evt-1","payload":{"ok":true},"dispatches":[{"id":"dispatch-1"}]}`)
		case "/v1/triggers/events/evt-1/raw":
			fmt.Fprint(w, `{"raw_body_base64":"eyJvayI6dHJ1ZX0=","raw_body_sha256":"hash","raw_body_size":11}`)
		case "/v1/triggers/events/evt-1/replay":
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"event_id":"evt-1","replay_id":"replay-1"}`)
		case "/v1/triggers/dispatches":
			fmt.Fprint(w, `{"items":[{"id":"dispatch-1","trigger_event_id":"evt-1","status":"triggered"}],"next_cursor":"dispatch-next"}`)
		case "/v1/triggers/dispatches/dispatch-1":
			fmt.Fprint(w, `{"id":"dispatch-1","status":"retryable_failed"}`)
		case "/v1/triggers/dispatches/dispatch-1/retry":
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"dispatch_id":"dispatch-1","retry_id":"retry-1"}`)
		case "/v1/triggers":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				fmt.Fprint(w, `{"trigger":{"id":"trigger-1"},"ingress_url":"https://ingress","secret":"secret"}`)
			} else {
				fmt.Fprint(w, `[{"id":"source-1","name":"github"}]`)
			}
		case "/v1/triggers/trigger-1":
			if r.Method == http.MethodDelete {
				if r.URL.Query().Get("force") != "true" {
					t.Error("force missing")
				}
				w.WriteHeader(http.StatusNoContent)
			} else {
				fmt.Fprint(w, `{"id":"source-1"}`)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := New(WithURL(server.URL), WithAuthToken("token"), WithOrgID("org-id"))
	if err != nil {
		t.Fatal(err)
	}
	if events, err := client.ListTriggerEvents(context.Background(), 23, "trigger-1"); err != nil || len(events) != 1 || len(events[0].Dispatches) != 1 {
		t.Fatalf("list: events=%#v err=%v", events, err)
	}
	if page, err := client.ListTriggerEventsPage(context.Background(), 23, "trigger-1", "cursor/value"); err != nil || page.NextCursor != "next" {
		t.Fatalf("list page: page=%#v err=%v", page, err)
	}
	if _, err := client.SearchTriggerEvents(context.Background(), models.TriggerEventListQuery{Limit: 1}); err == nil || !strings.Contains(err.Error(), "trigger is required") {
		t.Fatalf("expected trigger required error, got %v", err)
	}
	if event, err := client.GetTriggerEvent(context.Background(), "evt-1"); err != nil || len(event.Dispatches) != 1 {
		t.Fatalf("get: event=%#v err=%v", event, err)
	}
	if raw, err := client.GetTriggerEventRaw(context.Background(), "evt-1"); err != nil || raw.RawBodyBase64 != "eyJvayI6dHJ1ZX0=" || raw.RawBodySHA256 != "hash" {
		t.Fatalf("get raw: event=%#v err=%v", raw, err)
	}
	if replay, err := client.ReplayTriggerEvent(context.Background(), "evt-1"); err != nil || replay.ReplayID != "replay-1" {
		t.Fatalf("replay: response=%#v err=%v", replay, err)
	}
	if dispatches, err := client.ListTriggerEventDispatches(context.Background(), 17); err != nil || len(dispatches) != 1 || dispatches[0].Status != "triggered" {
		t.Fatalf("list dispatches: dispatches=%#v err=%v", dispatches, err)
	}
	if page, err := client.ListTriggerEventDispatchesPage(context.Background(), 17, "evt/1", "dispatch/cursor"); err != nil || page.NextCursor != "dispatch-next" {
		t.Fatalf("list dispatch page: page=%#v err=%v", page, err)
	}
	if dispatch, err := client.GetTriggerEventDispatch(context.Background(), "dispatch-1"); err != nil || dispatch.Status != "retryable_failed" {
		t.Fatalf("get dispatch: %#v %v", dispatch, err)
	}
	if retry, err := client.RetryTriggerEventDispatch(context.Background(), "dispatch-1"); err != nil || retry.RetryID != "retry-1" {
		t.Fatalf("retry: %#v %v", retry, err)
	}
	if created, err := client.CreateTrigger(context.Background(), &models.TriggerCreateRequest{Name: "github"}); err != nil || created.Secret != "secret" {
		t.Fatalf("create source: %#v %v", created, err)
	}
	if triggers, err := client.ListTriggers(context.Background()); err != nil || len(triggers) != 1 {
		t.Fatalf("list triggers: %#v %v", triggers, err)
	}
	if err := client.DeleteTrigger(context.Background(), "trigger-1", true); err != nil {
		t.Fatalf("delete source: %v", err)
	}
	want := []string{"GET /v1/triggers/trigger-1/events?limit=23", "GET /v1/triggers/trigger-1/events?cursor=cursor%2Fvalue&limit=23", "GET /v1/triggers/events/evt-1", "GET /v1/triggers/events/evt-1/raw", "POST /v1/triggers/events/evt-1/replay", "GET /v1/triggers/dispatches?limit=17", "GET /v1/triggers/dispatches?cursor=dispatch%2Fcursor&event_id=evt%2F1&limit=17", "GET /v1/triggers/dispatches/dispatch-1", "POST /v1/triggers/dispatches/dispatch-1/retry", "POST /v1/triggers", "GET /v1/triggers", "DELETE /v1/triggers/trigger-1?force=true"}
	if strings.Join(requests, "\n") != strings.Join(want, "\n") {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
}

func TestTriggerEventRequestIncludesErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"useful detail"}`)
	}))
	defer server.Close()
	client, err := New(WithURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.GetTriggerEvent(context.Background(), "evt-1")
	if err == nil || !strings.Contains(err.Error(), "useful detail") {
		t.Fatalf("error = %v", err)
	}
}
