package kube

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewNamespaceClient(ns string, client kubernetes.Interface) *namespaceClient {
	return &namespaceClient{
		namespace: ns,
		client:    client,
	}
}

type namespaceClient struct {
	namespace string
	client    kubernetes.Interface
}

// secretClient implements a corev1.SecretsInterface
type secretClient struct{ *namespaceClient }

var _ corev1.SecretInterface = (*secretClient)(nil)

func newSecretClient(lc *namespaceClient) *secretClient {
	return &secretClient{namespaceClient: lc}
}

func (s *secretClient) Create(ctx context.Context, secret *v1.Secret, opts metav1.CreateOptions) (result *v1.Secret, err error) {
	return s.client.CoreV1().Secrets(s.namespace).Create(ctx, secret, opts)
}

func (s *secretClient) Update(ctx context.Context, secret *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error) {
	return s.client.CoreV1().Secrets(s.namespace).Update(ctx, secret, opts)
}

func (s *secretClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return s.client.CoreV1().Secrets(s.namespace).Delete(ctx, name, opts)
}

func (s *secretClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.client.CoreV1().Secrets(s.namespace).DeleteCollection(ctx, opts, listOpts)
}

func (s *secretClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Secret, error) {
	return s.client.CoreV1().Secrets(s.namespace).Get(ctx, name, opts)
}

func (s *secretClient) List(ctx context.Context, opts metav1.ListOptions) (*v1.SecretList, error) {
	return s.client.CoreV1().Secrets(s.namespace).List(ctx, opts)
}

func (s *secretClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return s.client.CoreV1().Secrets(s.namespace).Watch(ctx, opts)
}

func (s *secretClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*v1.Secret, error) {
	return s.client.CoreV1().Secrets(s.namespace).Patch(ctx, name, pt, data, opts, subresources...)
}

func (s *secretClient) Apply(ctx context.Context, secretConfiguration *applycorev1.SecretApplyConfiguration, opts metav1.ApplyOptions) (*v1.Secret, error) {
	return s.client.CoreV1().Secrets(s.namespace).Apply(ctx, secretConfiguration, opts)
}

// configMapClient implements a corev1.ConfigMapInterface
type configMapClient struct{ *namespaceClient }

var _ corev1.ConfigMapInterface = (*configMapClient)(nil)

func newConfigMapClient(lc *namespaceClient) *configMapClient {
	return &configMapClient{namespaceClient: lc}
}

func (c *configMapClient) Create(ctx context.Context, configMap *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(c.namespace).Create(ctx, configMap, opts)
}

func (c *configMapClient) Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(c.namespace).Update(ctx, configMap, opts)
}

func (c *configMapClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.CoreV1().ConfigMaps(c.namespace).Delete(ctx, name, opts)
}

func (c *configMapClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return c.client.CoreV1().ConfigMaps(c.namespace).DeleteCollection(ctx, opts, listOpts)
}

func (c *configMapClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(c.namespace).Get(ctx, name, opts)
}

func (c *configMapClient) List(ctx context.Context, opts metav1.ListOptions) (*v1.ConfigMapList, error) {
	return c.client.CoreV1().ConfigMaps(c.namespace).List(ctx, opts)
}

func (c *configMapClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.CoreV1().ConfigMaps(c.namespace).Watch(ctx, opts)
}

func (c *configMapClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*v1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(c.namespace).Patch(ctx, name, pt, data, opts, subresources...)
}

func (c *configMapClient) Apply(ctx context.Context, configMap *applycorev1.ConfigMapApplyConfiguration, opts metav1.ApplyOptions) (*v1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(c.namespace).Apply(ctx, configMap, opts)
}
