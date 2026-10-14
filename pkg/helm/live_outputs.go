package helm

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// LiveOutputs are the deployments, services, and ingresses from a rendered
// chart, fetched by name after apply. The map shape matches the historical
// helm release outputs: namespace, then name, then the object with spec cleared.
type LiveOutputs struct {
	Ingresses   map[string]interface{}
	Services    map[string]interface{}
	Deployments map[string]interface{}
}

// CollectLiveOutputs gets only the deployments, services, and ingresses named
// in resources. Kinds the chart did not render are not requested.
func CollectLiveOutputs(ctx context.Context, kubeCfg *rest.Config, resources []Resource, releaseNamespace string, l *zap.Logger) (LiveOutputs, error) {
	client, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return LiveOutputs{}, fmt.Errorf("unable to create kubernetes client: %w", err)
	}
	return collectLiveOutputs(ctx, client, resources, releaseNamespace, l)
}

func collectLiveOutputs(ctx context.Context, client kubernetes.Interface, resources []Resource, releaseNamespace string, l *zap.Logger) (LiveOutputs, error) {
	out := LiveOutputs{
		Ingresses:   map[string]interface{}{},
		Services:    map[string]interface{}{},
		Deployments: map[string]interface{}{},
	}
	seen := map[string]struct{}{}

	for _, resource := range resources {
		namespace := resource.namespaceOr(releaseNamespace)
		key := resource.APIVersion + "/" + resource.Kind + "/" + namespace + "/" + resource.Name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		switch {
		case resource.Kind == "Deployment" && resource.APIVersion == "apps/v1":
			dep, err := client.AppsV1().Deployments(namespace).Get(ctx, resource.Name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				continue
			}
			if err != nil {
				return LiveOutputs{}, fmt.Errorf("unable to get deployment %s/%s: %w", namespace, resource.Name, err)
			}
			l.Info("adding deployment to outputs", zap.String("namespace", dep.Namespace), zap.String("name", dep.Name))
			dep.Spec = appsv1.DeploymentSpec{}
			if err := addOutput(out.Deployments, dep.Namespace, dep.Name, dep); err != nil {
				return LiveOutputs{}, err
			}
		case resource.Kind == "Service" && resource.APIVersion == "v1":
			svc, err := client.CoreV1().Services(namespace).Get(ctx, resource.Name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				continue
			}
			if err != nil {
				return LiveOutputs{}, fmt.Errorf("unable to get service %s/%s: %w", namespace, resource.Name, err)
			}
			l.Info("adding service to outputs", zap.String("namespace", svc.Namespace), zap.String("name", svc.Name))
			svc.Spec = corev1.ServiceSpec{}
			if err := addOutput(out.Services, svc.Namespace, svc.Name, svc); err != nil {
				return LiveOutputs{}, err
			}
		case resource.Kind == "Ingress" && resource.APIVersion == "networking.k8s.io/v1":
			ing, err := client.NetworkingV1().Ingresses(namespace).Get(ctx, resource.Name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				continue
			}
			if err != nil {
				return LiveOutputs{}, fmt.Errorf("unable to get ingress %s/%s: %w", namespace, resource.Name, err)
			}
			l.Info("adding ingress to outputs", zap.String("namespace", ing.Namespace), zap.String("name", ing.Name))
			ing.Spec = networkingv1.IngressSpec{}
			if err := addOutput(out.Ingresses, ing.Namespace, ing.Name, ing); err != nil {
				return LiveOutputs{}, err
			}
		}
	}

	return out, nil
}

func addOutput(out map[string]interface{}, namespace, name string, obj any) error {
	raw, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("unable to encode %s/%s: %w", namespace, name, err)
	}
	var encoded map[string]interface{}
	if err := json.Unmarshal(raw, &encoded); err != nil {
		return fmt.Errorf("unable to encode %s/%s: %w", namespace, name, err)
	}
	existing, ok := out[namespace]
	if !ok {
		out[namespace] = map[string]interface{}{name: encoded}
		return nil
	}
	existing.(map[string]interface{})[name] = encoded
	return nil
}
