package healthcheck

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type WorkloadResource struct {
	Kind      string
	Name      string
	Namespace string
}

type Result struct {
	Resource WorkloadResource
	Ready    bool
	Message  string
}

const (
	defaultTimeout      = 30 * time.Second
	defaultPollInterval = 5 * time.Second
)

func CheckWorkloadHealth(ctx context.Context, l *zap.Logger, kubeCfg *rest.Config, resources []WorkloadResource, timeout time.Duration) error {
	if len(resources) == 0 {
		return nil
	}

	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return fmt.Errorf("unable to create kubernetes client for health check: %w", err)
	}

	if timeout == 0 {
		timeout = defaultTimeout
	}

	deadline := time.Now().Add(timeout)

	for {
		var unhealthy []Result

		for _, res := range resources {
			result, err := checkResource(ctx, clientset, res)
			if err != nil {
				return fmt.Errorf("error checking health of %s %s/%s: %w", res.Kind, res.Namespace, res.Name, err)
			}
			if !result.Ready {
				unhealthy = append(unhealthy, *result)
			}
		}

		if len(unhealthy) == 0 {
			l.Info("all workload resources are healthy")
			return nil
		}

		if time.Now().After(deadline) {
			return buildUnhealthyError(unhealthy)
		}

		for _, u := range unhealthy {
			l.Debug("workload not ready, will retry",
				zap.String("kind", u.Resource.Kind),
				zap.String("name", u.Resource.Name),
				zap.String("namespace", u.Resource.Namespace),
				zap.String("reason", u.Message))
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(defaultPollInterval):
		}
	}
}

func checkResource(ctx context.Context, clientset *kubernetes.Clientset, res WorkloadResource) (*Result, error) {
	result := &Result{Resource: res}

	switch res.Kind {
	case "Deployment":
		return checkDeployment(ctx, clientset, res)
	case "StatefulSet":
		return checkStatefulSet(ctx, clientset, res)
	case "DaemonSet":
		return checkDaemonSet(ctx, clientset, res)
	default:
		result.Ready = true
		return result, nil
	}
}

func checkDeployment(ctx context.Context, clientset *kubernetes.Clientset, res WorkloadResource) (*Result, error) {
	result := &Result{Resource: res}

	deploy, err := clientset.AppsV1().Deployments(res.Namespace).Get(ctx, res.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	result.Ready, result.Message = isDeploymentReady(deploy)
	return result, nil
}

func isDeploymentReady(deploy *appsv1.Deployment) (bool, string) {
	desired := int32(1)
	if deploy.Spec.Replicas != nil {
		desired = *deploy.Spec.Replicas
	}

	if deploy.Status.ReadyReplicas < desired {
		return false, fmt.Sprintf("ready replicas %d/%d", deploy.Status.ReadyReplicas, desired)
	}

	if deploy.Status.UnavailableReplicas > 0 {
		return false, fmt.Sprintf("%d unavailable replicas", deploy.Status.UnavailableReplicas)
	}

	return true, ""
}

func checkStatefulSet(ctx context.Context, clientset *kubernetes.Clientset, res WorkloadResource) (*Result, error) {
	result := &Result{Resource: res}

	sts, err := clientset.AppsV1().StatefulSets(res.Namespace).Get(ctx, res.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	desired := int32(1)
	if sts.Spec.Replicas != nil {
		desired = *sts.Spec.Replicas
	}

	if sts.Status.ReadyReplicas < desired {
		result.Message = fmt.Sprintf("ready replicas %d/%d", sts.Status.ReadyReplicas, desired)
		return result, nil
	}

	result.Ready = true
	return result, nil
}

func checkDaemonSet(ctx context.Context, clientset *kubernetes.Clientset, res WorkloadResource) (*Result, error) {
	result := &Result{Resource: res}

	ds, err := clientset.AppsV1().DaemonSets(res.Namespace).Get(ctx, res.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if ds.Status.DesiredNumberScheduled == 0 {
		result.Ready = true
		return result, nil
	}

	if ds.Status.NumberReady < ds.Status.DesiredNumberScheduled {
		result.Message = fmt.Sprintf("ready %d/%d", ds.Status.NumberReady, ds.Status.DesiredNumberScheduled)
		return result, nil
	}

	result.Ready = true
	return result, nil
}

func buildUnhealthyError(unhealthy []Result) error {
	msg := "workload health check failed: the following resources are not ready:\n"
	for _, u := range unhealthy {
		msg += fmt.Sprintf("  - %s %s/%s: %s\n", u.Resource.Kind, u.Resource.Namespace, u.Resource.Name, u.Message)
	}
	return fmt.Errorf("%s", msg)
}
