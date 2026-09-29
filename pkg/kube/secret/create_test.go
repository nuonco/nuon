package secret

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpsert_CreatesNamespaceAndSecret(t *testing.T) {
	t.Parallel()

	k, err := New(validator.New(),
		WithNamespace("my-ns"),
		WithName("my-secret"),
		WithKey("api-key"),
	)
	require.NoError(t, err)

	client := fake.NewClientset()

	err = k.upsert(context.Background(), client, []byte("s3cr3t"))
	require.NoError(t, err)

	ns, err := client.CoreV1().Namespaces().Get(context.Background(), "my-ns", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "my-ns", ns.Name)

	got, err := client.CoreV1().Secrets("my-ns").Get(context.Background(), "my-secret", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, []byte("s3cr3t"), got.Data["api-key"])
}

func TestUpsert_ExistingNamespace(t *testing.T) {
	t.Parallel()

	k, err := New(validator.New(),
		WithNamespace("default"),
		WithName("my-secret"),
		WithKey("api-key"),
	)
	require.NoError(t, err)

	client := fake.NewClientset(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
	})

	err = k.upsert(context.Background(), client, []byte("v"))
	require.NoError(t, err)
}

func TestUpsert_PerKeyFieldManagerDoesNotClobber(t *testing.T) {
	t.Parallel()

	client := fake.NewClientset()

	apiKeyMgr, err := New(validator.New(),
		WithNamespace("datadog"),
		WithName("datadog"),
		WithKey("api-key"),
	)
	require.NoError(t, err)
	require.NoError(t, apiKeyMgr.upsert(context.Background(), client, []byte("api-value")))

	appKeyMgr, err := New(validator.New(),
		WithNamespace("datadog"),
		WithName("datadog"),
		WithKey("app-key"),
	)
	require.NoError(t, err)
	require.NoError(t, appKeyMgr.upsert(context.Background(), client, []byte("app-value")))

	got, err := client.CoreV1().Secrets("datadog").Get(context.Background(), "datadog", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, []byte("api-value"), got.Data["api-key"], "first manager's key should be preserved")
	assert.Equal(t, []byte("app-value"), got.Data["app-key"], "second manager's key should be present")
}
