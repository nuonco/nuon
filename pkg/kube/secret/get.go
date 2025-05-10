package secret

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *k8sSecretManager) Get(ctx context.Context) ([]byte, error) {
	kubeClient, err := k.getClient(ctx)
	if err != nil {
		return nil, err
	}
	client := kubeClient.CoreV1().Secrets(k.Namespace)

	secret, err := k.getSecret(ctx, client, k.Name)
	if err != nil {
		return nil, fmt.Errorf("unable to get secret: %w", err)
	}

	encVal, ok := secret.Data[k.Key]
	if !ok {
		return nil, fmt.Errorf("key not found on secret: %v", k.Name)
	}
	if len(encVal) < 1 {
		return nil, fmt.Errorf("key was empty on secret: %s", k.Name)
	}

	return encVal, nil
}

type kubeClientSecretGetter interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Secret, error)
}

func (k *k8sSecretManager) getSecret(ctx context.Context, client kubeClientSecretGetter, secretName string) (*corev1.Secret, error) {
	return client.Get(ctx, secretName, metav1.GetOptions{})
}
