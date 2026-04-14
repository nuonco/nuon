package composite_errors

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	diagnosticsTimeout = 30 * time.Second
	maxLogLines        = 50
)

// CollectHelmK8sDiagnostics queries the Kubernetes cluster after a Helm failure
// to gather diagnostic information about failing resources: pod statuses, events,
// and container logs. It returns structured composite errors for each finding.
func CollectHelmK8sDiagnostics(ctx context.Context, client kubernetes.Interface, namespace, releaseName string) []CompositeError {
	ctx, cancel := context.WithTimeout(ctx, diagnosticsTimeout)
	defer cancel()

	var errs []CompositeError

	// Collect pod diagnostics for the release
	errs = append(errs, collectPodDiagnostics(ctx, client, namespace, releaseName)...)

	// Collect recent warning events for the namespace filtered by release
	errs = append(errs, collectEventDiagnostics(ctx, client, namespace, releaseName)...)

	return errs
}

func collectPodDiagnostics(ctx context.Context, client kubernetes.Interface, namespace, releaseName string) []CompositeError {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName),
	})
	if err != nil {
		return nil
	}

	var errs []CompositeError
	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Ready {
				continue
			}

			ce := diagnoseContainerStatus(ctx, client, namespace, pod.Name, cs)
			if ce != nil {
				errs = append(errs, *ce)
			}
		}

		for _, cs := range pod.Status.InitContainerStatuses {
			if cs.Ready {
				continue
			}

			ce := diagnoseContainerStatus(ctx, client, namespace, pod.Name, cs)
			if ce != nil {
				errs = append(errs, *ce)
			}
		}
	}

	return errs
}

func diagnoseContainerStatus(ctx context.Context, client kubernetes.Interface, namespace, podName string, cs corev1.ContainerStatus) *CompositeError {
	metadata := map[string]any{
		"pod":       podName,
		"container": cs.Name,
		"namespace": namespace,
	}

	var summary string

	if cs.State.Waiting != nil {
		summary = fmt.Sprintf("Container %s/%s: %s", podName, cs.Name, cs.State.Waiting.Reason)
		metadata["reason"] = cs.State.Waiting.Reason
		if cs.State.Waiting.Message != "" {
			metadata["message"] = cs.State.Waiting.Message
		}
	} else if cs.State.Terminated != nil {
		summary = fmt.Sprintf("Container %s/%s: terminated with exit code %d", podName, cs.Name, cs.State.Terminated.ExitCode)
		metadata["reason"] = cs.State.Terminated.Reason
		metadata["exit_code"] = cs.State.Terminated.ExitCode
	} else {
		return nil
	}

	// Try to get container logs for non-ready containers
	detail := getContainerLogs(ctx, client, namespace, podName, cs.Name)

	return &CompositeError{
		OwnerType: "k8s-diagnostics",
		Severity:     "critical",
		Summary:      summary,
		Detail:       detail,
		Metadata:     metadata,
	}
}

func getContainerLogs(ctx context.Context, client kubernetes.Interface, namespace, podName, containerName string) string {
	tailLines := int64(maxLogLines)
	logOpts := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &tailLines,
	}

	// Try previous container logs first (for crash loops)
	logOpts.Previous = true
	logs, err := client.CoreV1().Pods(namespace).GetLogs(podName, logOpts).Do(ctx).Raw()
	if err != nil || len(logs) == 0 {
		// Fall back to current container logs
		logOpts.Previous = false
		logs, err = client.CoreV1().Pods(namespace).GetLogs(podName, logOpts).Do(ctx).Raw()
		if err != nil {
			return ""
		}
	}

	return string(logs)
}

func collectEventDiagnostics(ctx context.Context, client kubernetes.Interface, namespace, releaseName string) []CompositeError {
	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil
	}

	var errs []CompositeError
	cutoff := time.Now().Add(-10 * time.Minute)

	for _, event := range events.Items {
		if event.Type != corev1.EventTypeWarning {
			continue
		}

		// Filter to recent events
		eventTime := event.LastTimestamp.Time
		if eventTime.IsZero() {
			eventTime = event.CreationTimestamp.Time
		}
		if eventTime.Before(cutoff) {
			continue
		}

		// Filter to events related to the release by checking involved object labels
		// Events don't carry labels directly, so we match on namespace and look for
		// common failure patterns
		if !isRelevantEvent(event, releaseName) {
			continue
		}

		metadata := map[string]any{
			"namespace": namespace,
			"reason":    event.Reason,
			"kind":      event.InvolvedObject.Kind,
			"name":      event.InvolvedObject.Name,
		}

		errs = append(errs, CompositeError{
			OwnerType: "k8s-diagnostics",
			Severity:     "warning",
			Summary:      fmt.Sprintf("%s %s/%s: %s", event.Reason, event.InvolvedObject.Kind, event.InvolvedObject.Name, event.Message),
			Metadata:     metadata,
		})
	}

	return errs
}

func isRelevantEvent(event corev1.Event, releaseName string) bool {
	// Match events whose involved object name contains the release name
	name := event.InvolvedObject.Name
	if strings.Contains(name, releaseName) {
		return true
	}
	return false
}
