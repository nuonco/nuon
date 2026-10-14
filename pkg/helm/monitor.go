package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// monitorTarget is one workload kind to watch in one namespace.
// names are the objects of that kind the rendered chart contains there.
type monitorTarget struct {
	Namespace string
	Kind      string
	names     map[string]struct{}
}

// podLogTargets is the Deployment, StatefulSet, and ReplicaSet namespaces the
// chart will use. Pod logs are collected only for those objects.
func podLogTargets(resources []Resource, releaseNamespace string) []monitorTarget {
	index := map[string]int{}
	var targets []monitorTarget
	for _, resource := range resources {
		switch resource.Kind {
		case "Deployment", "StatefulSet", "ReplicaSet":
		default:
			continue
		}
		namespace := resource.namespaceOr(releaseNamespace)
		key := resource.Kind + "\x00" + namespace
		if i, ok := index[key]; ok {
			targets[i].names[resource.Name] = struct{}{}
			continue
		}
		index[key] = len(targets)
		targets = append(targets, monitorTarget{
			Namespace: namespace,
			Kind:      resource.Kind,
			names:     map[string]struct{}{resource.Name: {}},
		})
	}
	return targets
}

func (t monitorTarget) owns(name string) bool {
	_, ok := t.names[name]
	return ok
}

func podsForTargets(
	ctx context.Context,
	client kubernetes.Interface,
	targets []monitorTarget,
	labelSelector string,
	annotationKey string,
	annotationValue string,
	createdAfter metav1.Time,
	l *zap.Logger,
) []*corev1.Pod {
	var pods []*corev1.Pod
	for _, target := range targets {
		owners, err := listWorkloadOwners(ctx, client, target, labelSelector)
		if err != nil {
			l.Error("failed to fetch workloads",
				zap.String("kind", target.Kind),
				zap.String("namespace", target.Namespace),
				zap.String("label_selector", labelSelector),
				zap.Error(err),
			)
			continue
		}
		for _, owner := range owners {
			if !target.owns(owner.name) {
				continue
			}
			value, ok := owner.annotations[annotationKey]
			if !ok || value != annotationValue {
				continue
			}
			if len(owner.matchLabels) == 0 {
				continue
			}
			set := labels.Set(owner.matchLabels)
			owned, err := client.CoreV1().Pods(target.Namespace).List(ctx, metav1.ListOptions{LabelSelector: set.AsSelector().String()})
			if err != nil {
				l.Error("failed to fetch pods",
					zap.String("kind", target.Kind),
					zap.String("namespace", target.Namespace),
					zap.String("name", owner.name),
					zap.Error(err),
				)
				continue
			}
			for _, pod := range owned.Items {
				if pod.CreationTimestamp.Before(&createdAfter) {
					continue
				}
				pod := pod
				pods = append(pods, &pod)
			}
		}
	}
	return pods
}

type workloadOwner struct {
	name        string
	annotations map[string]string
	matchLabels map[string]string
}

func listWorkloadOwners(ctx context.Context, client kubernetes.Interface, target monitorTarget, labelSelector string) ([]workloadOwner, error) {
	opts := metav1.ListOptions{LabelSelector: labelSelector}
	switch target.Kind {
	case "Deployment":
		list, err := client.AppsV1().Deployments(target.Namespace).List(ctx, opts)
		if err != nil {
			return nil, err
		}
		owners := make([]workloadOwner, 0, len(list.Items))
		for _, item := range list.Items {
			owners = append(owners, workloadOwner{name: item.Name, annotations: item.Annotations, matchLabels: selectorLabels(item.Spec.Selector)})
		}
		return owners, nil
	case "StatefulSet":
		list, err := client.AppsV1().StatefulSets(target.Namespace).List(ctx, opts)
		if err != nil {
			return nil, err
		}
		owners := make([]workloadOwner, 0, len(list.Items))
		for _, item := range list.Items {
			owners = append(owners, workloadOwner{name: item.Name, annotations: item.Annotations, matchLabels: selectorLabels(item.Spec.Selector)})
		}
		return owners, nil
	case "ReplicaSet":
		list, err := client.AppsV1().ReplicaSets(target.Namespace).List(ctx, opts)
		if err != nil {
			return nil, err
		}
		owners := make([]workloadOwner, 0, len(list.Items))
		for _, item := range list.Items {
			owners = append(owners, workloadOwner{name: item.Name, annotations: item.Annotations, matchLabels: selectorLabels(item.Spec.Selector)})
		}
		return owners, nil
	default:
		return nil, fmt.Errorf("unsupported workload kind %q", target.Kind)
	}
}

func selectorLabels(sel *metav1.LabelSelector) map[string]string {
	if sel == nil {
		return nil
	}
	return sel.MatchLabels
}

func monitorNamespaces(targets []monitorTarget) []string {
	seen := map[string]struct{}{}
	var namespaces []string
	for _, target := range targets {
		if _, ok := seen[target.Namespace]; ok {
			continue
		}
		seen[target.Namespace] = struct{}{}
		namespaces = append(namespaces, target.Namespace)
	}
	return namespaces
}
