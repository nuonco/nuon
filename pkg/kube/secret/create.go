package secret

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
)

func (k *k8sSecretManager) Upsert(ctx context.Context, value []byte) error {
	kubeClient, err := k.getClient(ctx)
	if err != nil {
		return err
	}

	return k.upsert(ctx, kubeClient, value)
}

// upsert applies the secret using server-side apply with a per-key field manager. The field manager is scoped to
// k.Key so two Nuon secrets writing different keys into the same Kubernetes secret each own only their own key and
// never clobber one another's data: with SSA's granular map merge, an apply that declares only its key leaves keys
// owned by other managers intact. Force is set so the first SSA apply can take ownership of a key previously written
// by the legacy (non-SSA) create/update path on migration.
//
// It is split out of Upsert so tests can pass a fake clientset (fake.NewClientset models SSA field management, so the
// multi-key coexistence guarantee is unit-testable; see create_test.go).
func (k *k8sSecretManager) upsert(ctx context.Context, kubeClient kubernetes.Interface, value []byte) error {
	// Create the namespace if it doesn't exist.
	nsClient := kubeClient.CoreV1().Namespaces()
	if _, err := nsClient.Get(ctx, k.Namespace, metav1.GetOptions{}); err != nil {
		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: k.Namespace,
			},
		}
		if _, err := nsClient.Create(ctx, namespace, metav1.CreateOptions{}); err != nil {
			return errors.Wrap(err, "unable to create a namespace")
		}
	}

	secretApply := applycorev1.Secret(k.Name, k.Namespace).
		WithType(v1.SecretTypeOpaque).
		WithData(map[string][]byte{k.Key: value})

	_, err := kubeClient.CoreV1().Secrets(k.Namespace).Apply(ctx, secretApply, metav1.ApplyOptions{
		FieldManager: "nuon-secret-sync-" + k.Key,
		Force:        true,
	})

	return errors.Wrap(err, "unable to apply secret")
}
