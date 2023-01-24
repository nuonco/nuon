package token

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type k8sTokenGetter struct {
	Namespace string `validate:"required"`
	Name      string `validate:"required"`

	// internal state
	v *validator.Validate
	// NOTE: this is only used during testing to stub out the actual call
	client kubeClientSecretGetter
}

type k8sTokenGetterOption func(*k8sTokenGetter) error

func New(v *validator.Validate, opts ...k8sTokenGetterOption) (*k8sTokenGetter, error) {
	k := &k8sTokenGetter{v: v}

	if v == nil {
		return nil, fmt.Errorf("error instantiating token getter: validator is nil")
	}

	for _, opt := range opts {
		if err := opt(k); err != nil {
			return nil, err
		}
	}

	if err := k.v.Struct(k); err != nil {
		return nil, err
	}

	return k, nil
}

func WithNamespace(n string) k8sTokenGetterOption {
	return func(ktg *k8sTokenGetter) error {
		ktg.Namespace = n
		return nil
	}
}

func WithName(n string) k8sTokenGetterOption {
	return func(ktg *k8sTokenGetter) error {
		ktg.Name = n
		return nil
	}
}

func (k *k8sTokenGetter) GetOrgToken(ctx context.Context) (string, error) {
	client, err := k.getClient()
	if err != nil {
		return "", err
	}

	secret, err := k.getSecret(ctx, client, k.Name)
	if err != nil {
		return "", fmt.Errorf("unable to get secret: %w", err)
	}

	encVal, ok := secret.Data["token"]
	if !ok {
		return "", fmt.Errorf("token not found on secret: %v", k.Name)
	}
	if len(encVal) < 1 {
		return "", fmt.Errorf("token was empty on secret: %s", k.Name)
	}

	return string(encVal), nil
}

func (k *k8sTokenGetter) getClient() (kubeClientSecretGetter, error) {
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

func (k *k8sTokenGetter) getSecret(ctx context.Context, client kubeClientSecretGetter, secretName string) (*corev1.Secret, error) {
	return client.Get(ctx, secretName, metav1.GetOptions{})
}
