package secret

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/go-kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (k *k8sSecretGetter) Get(ctx context.Context) ([]byte, error) {
	client, err := k.getClient()
	if err != nil {
		return nil, err
	}

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

func (k *k8sSecretGetter) getClient() (kubeClientSecretGetter, error) {
	if k.client != nil {
		return k.client, nil
	}

	cfg, err := kube.GetKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %w", err)
	}

	return clientset.CoreV1().Secrets(k.Namespace), nil
}

type kubeClientSecretGetter interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Secret, error)
}

func (k *k8sSecretGetter) getSecret(ctx context.Context, client kubeClientSecretGetter, secretName string) (*corev1.Secret, error) {
	return client.Get(ctx, secretName, metav1.GetOptions{})
}
