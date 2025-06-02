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

/*
Functions to ingresses, services, and deployments managed by the given chart
from the kubernetes cluster and returns a map like this:

	{
		"namespace": {"name": JsonMarshalled(networkingv1.Ingress)}
	}

	{
		"namespace": {"name": JsonMarshalled(corev1.Service)},
	}

	{
		"namespace": {"name": JsonMarshalled(corev1.Deployment)},
	}
*/

func K8SGetHelmReleaseIngresses(ctx context.Context, chartName string, kubeCfg *rest.Config, l *zap.Logger) (map[string]interface{}, error) {
	// return values
	ingressesOut := map[string]interface{}{}

	// initialize a kube client - up to so we can exit early in case of error
	client, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, err
	}

	// fetch the resources from k8s
	// filter by annotation: meta.helm.sh/release-name
	annotationSelectorKey := "meta.helm.sh/release-name"
	annotationSelectorValue := chartName
	labelSelector := "app.kubernetes.io/managed-by=Helm"

	ingresses, err := client.NetworkingV1().
		Ingresses("").
		List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	// serialize into something GORM can write
	for _, ing := range ingresses.Items {
		l.Info(fmt.Sprintf("found ingress: %s", ing.Name), zap.Any("annotations", ing.GetAnnotations()))
		helmAnnotation, ok := ing.GetAnnotations()[annotationSelectorKey]
		if !ok || helmAnnotation != annotationSelectorValue {
			continue
		}
		l.Info(fmt.Sprintf("adding ingress to ouputs %s", ing.Name))
		var ingInterface map[string]interface{}
		// set ingress spec to empty to reduce output size.
		ing.Spec = networkingv1.IngressSpec{}
		inrec, _ := json.Marshal(ing)
		json.Unmarshal(inrec, &ingInterface)
		nsIngresses, ok := ingressesOut[ing.Namespace]
		if ok { // the namespace exists; it's holding a map. add a key to the map.
			nsIngresses.(map[string]interface{})[ing.Name] = ingInterface
			ingressesOut[ing.Namespace] = nsIngresses
		} else { // the namespace does not exist - set to map[string]ingIterfaces
			ingressesOut[ing.Namespace] = map[string]interface{}{ing.Name: ingInterface}
		}
	}
	return ingressesOut, nil
}

func K8SGetHelmReleaseServices(ctx context.Context, chartName string, kubeCfg *rest.Config, l *zap.Logger) (map[string]interface{}, error) {
	// return values
	servicesOut := map[string]interface{}{}

	// initialize a kube client - up to so we can exit early in case of error
	client, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, err
	}

	// fetch the resources from k8s
	// filter by annotation: meta.helm.sh/release-name
	annotationSelectorKey := "meta.helm.sh/release-name"
	annotationSelectorValue := chartName
	labelSelector := "app.kubernetes.io/managed-by=Helm"

	services, err := client.CoreV1().
		Services("").
		List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	// serialize into something GORM can write
	for _, svc := range services.Items {
		l.Info(fmt.Sprintf("found service: %s", svc.Name), zap.Any("annotations", svc.GetAnnotations()))
		helmAnnotation, ok := svc.GetAnnotations()[annotationSelectorKey]
		if !ok || helmAnnotation != annotationSelectorValue {
			continue
		}
		l.Info(fmt.Sprintf("adding service to ouputs %s", svc.Name))
		var svcInterface map[string]interface{}
		// set service spec to empty to reduce output size.
		svc.Spec = corev1.ServiceSpec{}
		inrec, _ := json.Marshal(svc)
		json.Unmarshal(inrec, &svcInterface)
		nsServices, ok := servicesOut[svc.Namespace]
		if ok { // the namespace exists; it's holding a map. add a key to the map.
			nsServices.(map[string]interface{})[svc.Name] = svcInterface
			servicesOut[svc.Namespace] = nsServices
		} else { // the namespace does not exist - set to map[string]ingIterfaces
			servicesOut[svc.Namespace] = map[string]interface{}{svc.Name: svcInterface}
		}
	}
	return servicesOut, nil
}

func K8SGetHelmReleaseDeployments(ctx context.Context, chartName string, kubeCfg *rest.Config, l *zap.Logger) (map[string]interface{}, error) {
	// return values
	deploymentsOut := map[string]interface{}{}

	// initialize a kube client - up to so we can exit early in case of error
	client, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, err
	}

	// fetch the resources from k8s
	// filter by annotation: meta.helm.sh/release-name
	annotationSelectorKey := "meta.helm.sh/release-name"
	annotationSelectorValue := chartName
	labelSelector := "app.kubernetes.io/managed-by=Helm"

	deployments, err := client.AppsV1().
		Deployments("").
		List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	// serialize into something GORM can write
	for _, dpl := range deployments.Items {
		l.Info(fmt.Sprintf("found deployment: %s", dpl.Name), zap.Any("annotations", dpl.GetAnnotations()))
		helmAnnotation, ok := dpl.GetAnnotations()[annotationSelectorKey]
		if !ok || helmAnnotation != annotationSelectorValue {
			continue
		}
		l.Info(fmt.Sprintf("adding deployment to ouputs %s", dpl.Name))
		var dplInterface map[string]interface{}
		// set deployment spec to empty to reduce output size.
		dpl.Spec = appsv1.DeploymentSpec{}
		inrec, _ := json.Marshal(dpl)
		json.Unmarshal(inrec, &dplInterface)
		nsDeployments, ok := deploymentsOut[dpl.Namespace]
		if ok { // the namespace exists; it's holding a map. add a key to the map.
			nsDeployments.(map[string]interface{})[dpl.Name] = dplInterface
			deploymentsOut[dpl.Namespace] = nsDeployments
		} else { // the namespace does not exist - set to map[string]dplIterfaces
			deploymentsOut[dpl.Namespace] = map[string]interface{}{dpl.Name: dplInterface}
		}
	}
	return deploymentsOut, nil
}
