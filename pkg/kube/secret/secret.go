package secret

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=secret_mock.go -source=secret.go -package=secret
type Getter interface {
	Get(context.Context) ([]byte, error)
}

var _ Getter = (*k8sSecretGetter)(nil)

type k8sSecretGetter struct {
	Namespace   string `validate:"required"`
	Name        string `validate:"required"`
	Key         string `validate:"required"`
	ClusterInfo *kube.ClusterInfo

	// internal state
	v *validator.Validate
	// NOTE: this is only used during testing to stub out the actual call
	client kubeClientSecretGetter
}

type k8sSecretGetterOption func(*k8sSecretGetter) error

func New(v *validator.Validate, opts ...k8sSecretGetterOption) (*k8sSecretGetter, error) {
	k := &k8sSecretGetter{v: v}

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

func WithNamespace(n string) k8sSecretGetterOption {
	return func(sg *k8sSecretGetter) error {
		sg.Namespace = n
		return nil
	}
}

func WithName(n string) k8sSecretGetterOption {
	return func(sg *k8sSecretGetter) error {
		sg.Name = n
		return nil
	}
}

func WithKey(n string) k8sSecretGetterOption {
	return func(sg *k8sSecretGetter) error {
		sg.Key = n
		return nil
	}
}

func WithCluster(cfg *kube.ClusterInfo) k8sSecretGetterOption {
	return func(sg *k8sSecretGetter) error {
		sg.ClusterInfo = cfg
		return nil
	}
}
