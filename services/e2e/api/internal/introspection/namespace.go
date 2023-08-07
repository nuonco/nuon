package introspection

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const KubeNamespaceDescription = "Returns details about a namespace"

func (s *svc) GetNamespaceHandler(ctx *gin.Context) {
	namespace := ctx.Param("namespace")

	resp, err := s.getNamespaceHandler(ctx, namespace)
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: KubeNamespaceDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: KubeNamespaceDescription,
		Response:    resp,
	})
}

func (s *svc) getNamespaceHandler(ctx context.Context, namespace string) (*kubeNamespaceResponse, error) {
	kubeCfg, err := s.getKubeConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to get kube config: %w", err)
	}

	resp := &kubeNamespaceResponse{
		Name: namespace,
	}

	secrets, err := clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get secrets: %w", err)
	}
	resp.SecretsCount = len(secrets.Items)
	resp.Secrets = secrets.Items

	services, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get services: %w", err)
	}
	resp.ServicesCount = len(services.Items)
	resp.Services = services.Items

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get pods: %w", err)
	}
	resp.PodsCount = len(pods.Items)
	resp.Pods = pods.Items

	return resp, nil
}

type kubeNamespaceResponse struct {
	Name string `json:"name"`

	SecretsCount int             `json:"secrets_count"`
	Secrets      []corev1.Secret `json:"secrets"`

	PodsCount int          `json:"pods_count"`
	Pods      []corev1.Pod `json:"pods"`

	ServicesCount int              `json:"services_count"`
	Services      []corev1.Service `json:"services"`
}
