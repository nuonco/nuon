package waypoint

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/go-kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const tokenSecretNameTmpl string = "waypoint-bootstrap-token-%s"

type k8sTokenGetter struct {
	// NOTE: this is only used during testing to stub out the actual call
	client kubeClientSecretGetter
}

func (k *k8sTokenGetter) getOrgToken(ctx context.Context, namespace, orgID string) (string, error) {
	secretName := fmt.Sprintf(tokenSecretNameTmpl, orgID)

	client, err := k.getClient(namespace)
	if err != nil {
		return "", err
	}

	secret, err := k.getSecret(ctx, client, secretName)
	if err != nil {
		return "", fmt.Errorf("unable to get secret: %w", err)
	}

	encVal, ok := secret.Data["token"]
	if !ok {
		return "", fmt.Errorf("token not found on secret %v", secretName)
	}
	if len(encVal) < 1 {
		return "", fmt.Errorf("token was empty on secret %s", secretName)
	}

	return string(encVal), nil
}

func (k *k8sTokenGetter) getClient(namespace string) (kubeClientSecretGetter, error) {
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

	return clientset.CoreV1().Secrets(namespace), nil
}

type kubeClientSecretGetter interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Secret, error)
}

func (k *k8sTokenGetter) getSecret(ctx context.Context, client kubeClientSecretGetter, secretName string) (*corev1.Secret, error) {
	return client.Get(ctx, secretName, metav1.GetOptions{})
}
