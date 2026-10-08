package helm

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func streamLogs(
	streamCtx context.Context, cancelStreaming func(),
	streamer *LogStreamer,
	k8sClient *kubernetes.Clientset,
	targets []monitorTarget,
	labelSelector string,
	annotationSelectorKey string,
	annotationSelectorValue string,
	l *zap.Logger,
) {
	// The helm release annotation is on the pod owner, not the pod. List only
	// the workload kinds the chart rendered, and only in those namespaces.
	if len(targets) == 0 {
		l.Debug("chart has no deployments, statefulsets, or replicasets; not streaming pod logs")
		return
	}

	now := time.Now().UTC()
	namespaces := monitorNamespaces(targets)

	time.Sleep(3 * time.Second)

	for {
		select {
		case <-streamCtx.Done():
			return
		default:
			pods := podsForTargets(streamCtx, k8sClient, targets, labelSelector, annotationSelectorKey, annotationSelectorValue, metav1.NewTime(now), l)

			l.Info(fmt.Sprintf("streaming logs for %d pods", len(pods)),
				zap.Strings("namespaces", namespaces),
				zap.String("label_selector", labelSelector),
				zap.String("annotation", fmt.Sprintf("%s=%s", annotationSelectorKey, annotationSelectorValue)),
				zap.String("created_on.gte", now.String()),
			)

			if err := streamer.StreamPodLogs(streamCtx, pods); err != nil {
				// TODO(fd): use error wrap
				l.Error(fmt.Sprintf("Error streaming logs: %v", err))
			}

			time.Sleep(5 * time.Second)
		}
	}
}
