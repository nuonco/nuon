package helm

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResourcesFromManifest(t *testing.T) {
	manifest := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
---
apiVersion: v1
kind: Service
metadata:
  name: api
  namespace: other
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: api
---
# empty document
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ""
`
	resources, err := ResourcesFromManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 3 {
		t.Fatalf("got %d resources, want 3", len(resources))
	}
	if resources[0].Kind != "Deployment" || resources[0].Namespace != "" || resources[0].Name != "api" {
		t.Fatalf("deployment = %+v", resources[0])
	}
	if resources[1].Kind != "Service" || resources[1].Namespace != "other" {
		t.Fatalf("service = %+v", resources[1])
	}
	if resources[2].Kind != "ClusterRole" || resources[2].Namespace != "" {
		t.Fatalf("clusterrole = %+v", resources[2])
	}
}

func TestRenderManifestUsesReleaseNamespace(t *testing.T) {
	ch := &chart.Chart{
		Metadata: &chart.Metadata{Name: "app", Version: "0.1.0", APIVersion: "v2"},
		Templates: []*common.File{{
			Name: "templates/deploy.yaml",
			Data: []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: {{ .Release.Namespace }}
---
apiVersion: v1
kind: Service
metadata:
  name: api
  namespace: extra
`),
		}},
	}

	manifest, err := RenderManifest(ch, nil, "app", "install-ns")
	if err != nil {
		t.Fatal(err)
	}
	resources, err := ResourcesFromManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	byKind := map[string]Resource{}
	for _, resource := range resources {
		byKind[resource.Kind] = resource
	}
	if byKind["Deployment"].Namespace != "install-ns" {
		t.Fatalf("deployment namespace = %q\n%s", byKind["Deployment"].Namespace, manifest)
	}
	if byKind["Service"].Namespace != "extra" {
		t.Fatalf("service namespace = %q\n%s", byKind["Service"].Namespace, manifest)
	}
}

func TestCheckResourceAccess(t *testing.T) {
	gvApps := schema.GroupVersion{Group: "apps", Version: "v1"}
	gvCore := schema.GroupVersion{Version: "v1"}
	gvRBAC := schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1"}
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{gvApps, gvCore, gvRBAC})
	mapper.Add(gvApps.WithKind("Deployment"), meta.RESTScopeNamespace)
	mapper.Add(gvCore.WithKind("Service"), meta.RESTScopeNamespace)
	mapper.Add(gvRBAC.WithKind("ClusterRole"), meta.RESTScopeRoot)

	resources := []Resource{
		{APIVersion: "apps/v1", Kind: "Deployment", Name: "api"},
		{APIVersion: "apps/v1", Kind: "Deployment", Name: "worker"},
		{APIVersion: "v1", Kind: "Service", Name: "api", Namespace: "other"},
		{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRole", Name: "api"},
		{APIVersion: "custom.example.com/v1", Kind: "Widget", Name: "w"},
	}

	var listed []string
	err := checkResourceAccess(context.Background(), mapper, func(_ context.Context, gvr schema.GroupVersionResource, namespace string) error {
		listed = append(listed, gvr.Resource+"/"+namespace)
		if gvr.Resource == "services" {
			return apierrors.NewNotFound(schema.GroupResource{Resource: "services"}, "")
		}
		return nil
	}, resources, "app", zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"deployments/app", "services/other", "clusterroles/"}
	if len(listed) != len(want) {
		t.Fatalf("listed %v, want %v", listed, want)
	}
	for i := range want {
		if listed[i] != want[i] {
			t.Fatalf("listed %v, want %v", listed, want)
		}
	}

	err = checkResourceAccess(context.Background(), mapper, func(_ context.Context, gvr schema.GroupVersionResource, namespace string) error {
		return apierrors.NewForbidden(schema.GroupResource{Resource: gvr.Resource}, "", errors.New("no"))
	}, []Resource{{APIVersion: "apps/v1", Kind: "Deployment", Name: "api"}}, "app", zap.NewNop())
	if err == nil || err.Error() != `unable to list deployments in namespace "app": deployments is forbidden: no` {
		t.Fatalf("forbidden error = %v", err)
	}
}

func TestCollectLiveOutputsUsesRenderedObjects(t *testing.T) {
	replicas := int32(2)
	client := fake.NewSimpleClientset(
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "app"}, Spec: appsv1.DeploymentSpec{Replicas: &replicas}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: "app"}, Spec: appsv1.DeploymentSpec{Replicas: &replicas}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "extra"}, Spec: corev1.ServiceSpec{ClusterIP: "10.0.0.1"}},
		&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "app"}},
	)

	out, err := collectLiveOutputs(context.Background(), client, []Resource{
		{APIVersion: "apps/v1", Kind: "Deployment", Name: "api"},
		{APIVersion: "v1", Kind: "Service", Name: "api", Namespace: "extra"},
		{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "missing", Namespace: "app"},
		{APIVersion: "v1", Kind: "ConfigMap", Name: "cfg", Namespace: "app"},
	}, "app", zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}

	deps := out.Deployments["app"].(map[string]interface{})
	if _, ok := deps["other"]; ok {
		t.Fatal("fetched a deployment that is not in the chart")
	}
	api := deps["api"].(map[string]interface{})
	spec := api["spec"].(map[string]interface{})
	if _, ok := spec["replicas"]; ok {
		t.Fatalf("deployment spec still contains replicas: %#v", spec)
	}
	if _, ok := out.Services["extra"].(map[string]interface{})["api"]; !ok {
		t.Fatal("missing service from its namespace")
	}
	if len(out.Ingresses) != 0 {
		t.Fatalf("ingresses = %#v", out.Ingresses)
	}
}
