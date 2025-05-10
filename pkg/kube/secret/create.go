package secret

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *k8sSecretManager) Upsert(ctx context.Context, value []byte) error {
	kubeClient, err := k.getClient(ctx)
	if err != nil {
		return err
	}
	client := kubeClient.CoreV1().Secrets(k.Namespace)

	kubeSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name,
			Namespace: k.Namespace,
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			k.Key: value,
		},
	}

	_, err = client.Create(ctx, kubeSecret, metav1.CreateOptions{})
	if err != nil {
		_, err = client.Update(ctx, kubeSecret, metav1.UpdateOptions{})
	}

	return err
}
