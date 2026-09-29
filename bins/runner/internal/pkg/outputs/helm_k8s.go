package outputs

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/kubernetes"
)

func K8SGetHelmReleaseIngresses(ctx context.Context, chartName string, kubeCfg *rest.Config, l *zap.Logger) (map[string]interface{}, error) {
	ingressesOut := map[string]interface{}{}

	client, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, err
	}

	annotationSelectorKey := "meta.helm.sh/release-name"
	annotationSelectorValue := chartName
	labelSelector := "app.kubernetes.io/managed-by=Helm"

	ingresses, err := client.NetworkingV1().
		Ingresses("").
		List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	for _, ing := range ingresses.Items {
		l.Info(fmt.Sprintf("found ingress: %s", ing.Name), zap.Any("annotations", ing.GetAnnotations()))
		helmAnnotation, ok := ing.GetAnnotations()[annotationSelectorKey]
		if !ok || helmAnnotation != annotationSelectorValue {
			continue
		}
		l.Info(fmt.Sprintf("adding ingress to ouputs %s", ing.Name))
		var ingInterface map[string]interface{}
		ing.Spec = networkingv1.IngressSpec{}
		inrec, _ := json.Marshal(ing)
		json.Unmarshal(inrec, &ingInterface)
		nsIngresses, ok := ingressesOut[ing.Namespace]
		if ok {
			nsIngresses.(map[string]interface{})[ing.Name] = ingInterface
			ingressesOut[ing.Namespace] = nsIngresses
		} else {
			ingressesOut[ing.Namespace] = map[string]interface{}{ing.Name: ingInterface}
		}
	}
	return ingressesOut, nil
}

func K8SGetHelmReleaseServices(ctx context.Context, chartName string, kubeCfg *rest.Config, l *zap.Logger) (map[string]interface{}, error) {
	servicesOut := map[string]interface{}{}

	client, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, err
	}

	annotationSelectorKey := "meta.helm.sh/release-name"
	annotationSelectorValue := chartName
	labelSelector := "app.kubernetes.io/managed-by=Helm"

	services, err := client.CoreV1().
		Services("").
		List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	for _, svc := range services.Items {
		l.Info(fmt.Sprintf("found service: %s", svc.Name), zap.Any("annotations", svc.GetAnnotations()))
		helmAnnotation, ok := svc.GetAnnotations()[annotationSelectorKey]
		if !ok || helmAnnotation != annotationSelectorValue {
			continue
		}
		l.Info(fmt.Sprintf("adding service to ouputs %s", svc.Name))
		var svcInterface map[string]interface{}
		svc.Spec = corev1.ServiceSpec{}
		inrec, _ := json.Marshal(svc)
		json.Unmarshal(inrec, &svcInterface)
		nsServices, ok := servicesOut[svc.Namespace]
		if ok {
			nsServices.(map[string]interface{})[svc.Name] = svcInterface
			servicesOut[svc.Namespace] = nsServices
		} else {
			servicesOut[svc.Namespace] = map[string]interface{}{svc.Name: svcInterface}
		}
	}
	return servicesOut, nil
}

func K8SGetHelmReleaseDeployments(ctx context.Context, chartName string, kubeCfg *rest.Config, l *zap.Logger) (map[string]interface{}, error) {
	deploymentsOut := map[string]interface{}{}

	client, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, err
	}

	annotationSelectorKey := "meta.helm.sh/release-name"
	annotationSelectorValue := chartName
	labelSelector := "app.kubernetes.io/managed-by=Helm"

	deployments, err := client.AppsV1().
		Deployments("").
		List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	for _, dpl := range deployments.Items {
		l.Info(fmt.Sprintf("found deployment: %s", dpl.Name), zap.Any("annotations", dpl.GetAnnotations()))
		helmAnnotation, ok := dpl.GetAnnotations()[annotationSelectorKey]
		if !ok || helmAnnotation != annotationSelectorValue {
			continue
		}
		l.Info(fmt.Sprintf("adding deployment to ouputs %s", dpl.Name))
		var dplInterface map[string]interface{}
		dpl.Spec = appsv1.DeploymentSpec{}
		inrec, _ := json.Marshal(dpl)
		json.Unmarshal(inrec, &dplInterface)
		nsDeployments, ok := deploymentsOut[dpl.Namespace]
		if ok {
			nsDeployments.(map[string]interface{})[dpl.Name] = dplInterface
			deploymentsOut[dpl.Namespace] = nsDeployments
		} else {
			deploymentsOut[dpl.Namespace] = map[string]interface{}{dpl.Name: dplInterface}
		}
	}
	return deploymentsOut, nil
}
