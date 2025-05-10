package secret

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/kube"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=secret_mock.go -source=secret.go -package=secret
type Manager interface {
	Get(context.Context) ([]byte, error)
	Upsert(context.Context, []byte) error
}

var _ Manager = (*k8sSecretManager)(nil)

type k8sSecretManager struct {
	Namespace   string `validate:"required"`
	Name        string `validate:"required"`
	Key         string `validate:"required"`
	ClusterInfo *kube.ClusterInfo

	// internal state
	v *validator.Validate
}

type k8sSecretManagerOption func(*k8sSecretManager) error

func New(v *validator.Validate, opts ...k8sSecretManagerOption) (*k8sSecretManager, error) {
	k := &k8sSecretManager{v: v}

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

func WithNamespace(n string) k8sSecretManagerOption {
	return func(sg *k8sSecretManager) error {
		sg.Namespace = n
		return nil
	}
}

func WithName(n string) k8sSecretManagerOption {
	return func(sg *k8sSecretManager) error {
		sg.Name = n
		return nil
	}
}

func WithKey(n string) k8sSecretManagerOption {
	return func(sg *k8sSecretManager) error {
		sg.Key = n
		return nil
	}
}

func WithCluster(cfg *kube.ClusterInfo) k8sSecretManagerOption {
	return func(sg *k8sSecretManager) error {
		sg.ClusterInfo = cfg
		return nil
	}
}
