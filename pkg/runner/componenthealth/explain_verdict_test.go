package componenthealth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// A verdict with no message is the least actionable thing health can show: the
// reason an ingress was progressing sat in the status the engine had already read.
func TestAssessResourceExplainsBlankProgressing(t *testing.T) {
	ingress := func(status map[string]any) *unstructured.Unstructured {
		return &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "networking.k8s.io/v1",
			"kind":       "Ingress",
			"metadata":   map[string]any{"name": "public", "namespace": "whoami"},
			"status":     status,
		}}
	}

	t.Run("no load balancer address is named", func(t *testing.T) {
		health, msg, native := assessResource(ingress(map[string]any{"loadBalancer": map[string]any{}}))
		assert.Equal(t, healthProgressing, health)
		assert.Equal(t, "Progressing", native)
		assert.Equal(t, "no load balancer address assigned yet", msg)
	})

	t.Run("an assigned address is healthy and needs no explanation", func(t *testing.T) {
		health, msg, _ := assessResource(ingress(map[string]any{
			"loadBalancer": map[string]any{
				"ingress": []any{map[string]any{"ip": "20.232.231.3"}},
			},
		}))
		assert.Equal(t, healthHealthy, health)
		assert.Empty(t, msg)
	})
}
