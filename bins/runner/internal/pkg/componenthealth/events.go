package componenthealth

import (
	"context"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var eventsGVR = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "events"}

// eventWarningWindow bounds how recent a Warning event must be to count. A
// persistent controller failure keeps re-emitting (fresh timestamp) and stays
// flagged; a one-off warning the resource recovered from ages out and clears.
const eventWarningWindow = 15 * time.Minute

type warningEvent struct {
	reason  string
	message string
}

// latestWarnings returns, keyed by resource identity, the resources whose latest
// recent event is a Warning — surfacing controller-side failures that never
// show in the object's own status. Listed on demand, no informer or cache.
func (e *Engine) latestWarnings(ctx context.Context, dynClient dynamic.Interface) map[string]warningEvent {
	list, err := dynClient.Resource(eventsGVR).List(ctx, metav1.ListOptions{FieldSelector: "type=Warning"})
	if err != nil {
		e.l.Warn("unable to list warning events for component health", zap.Error(err))
		return nil
	}

	cutoff := time.Now().Add(-eventWarningWindow)
	type latest struct {
		ts      time.Time
		reason  string
		message string
	}
	byObject := map[string]latest{}
	for i := range list.Items {
		u := &list.Items[i]
		key, ok := eventObjectKey(u)
		if !ok {
			continue
		}
		ts := eventTimestamp(u)
		if ts.Before(cutoff) {
			continue
		}
		if cur, seen := byObject[key]; seen && !ts.After(cur.ts) {
			continue
		}
		reason, _, _ := unstructured.NestedString(u.Object, "reason")
		message, _, _ := unstructured.NestedString(u.Object, "message")
		byObject[key] = latest{ts: ts, reason: reason, message: message}
	}

	out := make(map[string]warningEvent, len(byObject))
	for key, l := range byObject {
		out[key] = warningEvent{reason: l.reason, message: l.message}
	}
	return out
}

func eventObjectKey(u *unstructured.Unstructured) (string, bool) {
	involved, ok, _ := unstructured.NestedMap(u.Object, "involvedObject")
	if !ok {
		return "", false
	}
	kind, _ := involved["kind"].(string)
	namespace, _ := involved["namespace"].(string)
	name, _ := involved["name"].(string)
	if kind == "" || name == "" {
		return "", false
	}
	return resourceKey(kind, namespace, name), true
}

func resourceKey(kind, namespace, name string) string {
	return kind + "/" + namespace + "/" + name
}

func eventTimestamp(u *unstructured.Unstructured) time.Time {
	for _, field := range []string{"lastTimestamp", "eventTime", "firstTimestamp"} {
		if s, ok, _ := unstructured.NestedString(u.Object, field); ok && s != "" {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}
