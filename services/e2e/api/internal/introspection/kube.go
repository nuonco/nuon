package introspection

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const kubeDescription = "Returns details about the kubernetes cluster and what is running in it."

func (s *svc) GetKubeHandler(ctx *gin.Context) {
	resp, err := s.getKubeHandler(ctx)
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: kubeDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: kubeDescription,
		Response:    resp,
	})
}

func (s *svc) getKubeHandler(ctx context.Context) (*kubeResponse, error) {
	kubeCfg, err := s.getKubeConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to get kube config: %w", err)
	}

	resp := &kubeResponse{
		Namespaces: make([]kubeNamespaceResponseShort, 0),
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get namespaces: %w", err)
	}
	for _, ns := range namespaces.Items {
		resp.Namespaces = append(resp.Namespaces, kubeNamespaceResponseShort{
			Name:   ns.Name,
			Status: ns.Status,
		})

	}
	return resp, nil
}

type kubeNamespaceResponseShort struct {
	Name   string                 `json:"name"`
	Status corev1.NamespaceStatus `json:"status"`
}

type kubeResponse struct {
	Namespaces []kubeNamespaceResponseShort `json:"namespaces"`
}
