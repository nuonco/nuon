package composite_errors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCollectHelmK8sDiagnostics_CrashLoopBackOff(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-app-abc123",
				Namespace: "default",
				Labels: map[string]string{
					"app.kubernetes.io/instance": "my-app",
				},
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:  "web",
						Ready: false,
						State: corev1.ContainerState{
							Waiting: &corev1.ContainerStateWaiting{
								Reason:  "CrashLoopBackOff",
								Message: "back-off 5m0s restarting failed container",
							},
						},
					},
				},
			},
		},
	)

	errs := CollectHelmK8sDiagnostics(context.Background(), client, "default", "my-app")
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.OwnerType == "k8s-diagnostics" && e.Metadata["reason"] == "CrashLoopBackOff" {
			found = true
			assert.Contains(t, e.Summary, "CrashLoopBackOff")
			assert.Equal(t, "critical", e.Severity)
			assert.Equal(t, "web", e.Metadata["container"])
			assert.Equal(t, "my-app-abc123", e.Metadata["pod"])
		}
	}
	assert.True(t, found, "expected CrashLoopBackOff diagnostic")
}

func TestCollectHelmK8sDiagnostics_TerminatedContainer(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-app-xyz789",
				Namespace: "default",
				Labels: map[string]string{
					"app.kubernetes.io/instance": "my-app",
				},
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:  "worker",
						Ready: false,
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{
								ExitCode: 1,
								Reason:   "Error",
							},
						},
					},
				},
			},
		},
	)

	errs := CollectHelmK8sDiagnostics(context.Background(), client, "default", "my-app")
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.Metadata["container"] == "worker" {
			found = true
			assert.Contains(t, e.Summary, "exit code 1")
			assert.Equal(t, int32(1), e.Metadata["exit_code"])
		}
	}
	assert.True(t, found, "expected terminated container diagnostic")
}

func TestCollectHelmK8sDiagnostics_WarningEvents(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-app-abc123.event1",
				Namespace: "default",
			},
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "my-app-abc123",
			},
			Type:          corev1.EventTypeWarning,
			Reason:        "FailedPullImage",
			Message:       "Failed to pull image \"nginx:invalid\": rpc error: code = NotFound",
			LastTimestamp:  metav1.Now(),
		},
	)

	errs := CollectHelmK8sDiagnostics(context.Background(), client, "default", "my-app")
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.Severity == "warning" && e.Metadata["reason"] == "FailedPullImage" {
			found = true
			assert.Contains(t, e.Summary, "Failed to pull image")
		}
	}
	assert.True(t, found, "expected warning event diagnostic")
}

func TestCollectHelmK8sDiagnostics_NoPods(t *testing.T) {
	client := fake.NewSimpleClientset()
	errs := CollectHelmK8sDiagnostics(context.Background(), client, "default", "my-app")
	assert.Empty(t, errs)
}

func TestCollectHelmK8sDiagnostics_AllReady(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-app-healthy",
				Namespace: "default",
				Labels: map[string]string{
					"app.kubernetes.io/instance": "my-app",
				},
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:  "web",
						Ready: true,
					},
				},
			},
		},
	)

	errs := CollectHelmK8sDiagnostics(context.Background(), client, "default", "my-app")
	assert.Empty(t, errs)
}

func TestCollectHelmK8sDiagnostics_InitContainerFailure(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-app-init",
				Namespace: "default",
				Labels: map[string]string{
					"app.kubernetes.io/instance": "my-app",
				},
			},
			Status: corev1.PodStatus{
				InitContainerStatuses: []corev1.ContainerStatus{
					{
						Name:  "migrate",
						Ready: false,
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{
								ExitCode: 2,
								Reason:   "Error",
							},
						},
					},
				},
			},
		},
	)

	errs := CollectHelmK8sDiagnostics(context.Background(), client, "default", "my-app")
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.Metadata["container"] == "migrate" {
			found = true
			assert.Contains(t, e.Summary, "exit code 2")
		}
	}
	assert.True(t, found, "expected init container failure diagnostic")
}
