package helm

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
)

func TestPodLogTargetsUsesChartNamespaces(t *testing.T) {
	targets := podLogTargets([]Resource{
		{Kind: "Deployment", Name: "api"},
		{Kind: "Deployment", Name: "worker", Namespace: "other"},
		{Kind: "Deployment", Name: "api"},
		{Kind: "Service", Name: "api", Namespace: "kube-system"},
		{Kind: "StatefulSet", Name: "db", Namespace: "data"},
		{Kind: "ReplicaSet", Name: "batch", Namespace: "app"},
		{Kind: "ConfigMap", Name: "cfg", Namespace: "app"},
		{Kind: "DaemonSet", Name: "agent", Namespace: "app"},
	}, "app")

	if len(targets) != 4 {
		t.Fatalf("got %d targets, want 4: %+v", len(targets), targets)
	}
	assertTarget := func(kind, namespace string, names ...string) {
		t.Helper()
		for _, target := range targets {
			if target.Kind != kind || target.Namespace != namespace {
				continue
			}
			if len(target.names) != len(names) {
				t.Fatalf("%s/%s names = %v, want %v", kind, namespace, target.names, names)
			}
			for _, name := range names {
				if !target.owns(name) {
					t.Fatalf("%s/%s missing %s", kind, namespace, name)
				}
			}
			return
		}
		t.Fatalf("missing %s in %s", kind, namespace)
	}
	assertTarget("Deployment", "app", "api")
	assertTarget("Deployment", "other", "worker")
	assertTarget("StatefulSet", "data", "db")
	assertTarget("ReplicaSet", "app", "batch")
}

func TestPodsForTargetsStaysInChartNamespaces(t *testing.T) {
	helmLabels := map[string]string{"app.kubernetes.io/managed-by": "Helm"}
	helmAnnotations := map[string]string{"meta.helm.sh/release-name": "demo"}
	recent := metav1.NewTime(time.Now().UTC())
	old := metav1.NewTime(time.Now().UTC().Add(-time.Hour))

	client := fake.NewSimpleClientset(
		workloadDeployment("api", "app", helmLabels, helmAnnotations, map[string]string{"app": "api"}),
		workloadDeployment("other", "kube-system", helmLabels, helmAnnotations, map[string]string{"app": "other"}),
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "app", Labels: helmLabels, Annotations: helmAnnotations},
			Spec:       appsv1.StatefulSetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "db"}}},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Name: "batch", Namespace: "extra", Labels: helmLabels, Annotations: helmAnnotations},
			Spec:       appsv1.ReplicaSetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "batch"}}},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "skipped", Namespace: "app", Labels: helmLabels, Annotations: helmAnnotations},
			Spec:       appsv1.DeploymentSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "skipped"}}},
		},
		pod("api-new", "app", map[string]string{"app": "api"}, recent),
		pod("api-old", "app", map[string]string{"app": "api"}, old),
		pod("batch-0", "extra", map[string]string{"app": "batch"}, recent),
		pod("stray", "kube-system", map[string]string{"app": "other"}, recent),
		pod("skipped-0", "app", map[string]string{"app": "skipped"}, recent),
	)

	var lists []string
	client.PrependReactor("list", "*", func(action kubetesting.Action) (bool, runtime.Object, error) {
		lists = append(lists, action.GetResource().Resource+"/"+action.GetNamespace())
		return false, nil, nil
	})

	targets := podLogTargets([]Resource{
		{Kind: "Deployment", Name: "api", Namespace: "app"},
		{Kind: "ReplicaSet", Name: "batch", Namespace: "extra"},
		{Kind: "Service", Name: "api", Namespace: "kube-system"},
	}, "app")
	pods := podsForTargets(context.Background(), client, targets, "app.kubernetes.io/managed-by=Helm", "meta.helm.sh/release-name", "demo", metav1.NewTime(old.Add(time.Minute)), zap.NewNop())

	got := map[string]bool{}
	for _, pod := range pods {
		got[pod.Namespace+"/"+pod.Name] = true
	}
	if !got["app/api-new"] || !got["extra/batch-0"] || len(got) != 2 {
		t.Fatalf("pods = %#v", got)
	}
	for _, list := range lists {
		switch list {
		case "deployments/app", "replicasets/extra", "pods/app", "pods/extra":
		default:
			t.Fatalf("unexpected list %s in %v", list, lists)
		}
	}
}

func workloadDeployment(name, namespace string, labels, annotations, selector map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels, Annotations: annotations},
		Spec:       appsv1.DeploymentSpec{Selector: &metav1.LabelSelector{MatchLabels: selector}},
	}
}

func pod(name, namespace string, labels map[string]string, created metav1.Time) *corev1.Pod {
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels, CreationTimestamp: created}}
}
