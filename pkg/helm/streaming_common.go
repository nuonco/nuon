package helm

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func streamLogs(
	streamCtx context.Context, cancelStreaming func(),
	streamer *LogStreamer,
	k8sClient *kubernetes.Clientset,
	labelSelector string,
	annotationSelectorKey string,
	annotationSelectorValue string,
	l *zap.Logger,
) {
	now := time.Now().UTC()

	time.Sleep(3 * time.Second)

	for {
		select {
		case <-streamCtx.Done():
			return
		default:
			pods := []*corev1.Pod{}

			deployments, err := k8sClient.AppsV1().Deployments("").List(streamCtx, metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err != nil {
				l.Error("failed to fetch deployments", zap.String("label_selector", labelSelector))
			}
			for _, dpl := range deployments.Items {
				value, ok := dpl.Annotations[annotationSelectorKey]
				if !ok || value != annotationSelectorValue {
					continue
				}
				set := labels.Set(dpl.Spec.Selector.MatchLabels)
				dplPods, err := k8sClient.CoreV1().Pods(dpl.Namespace).List(streamCtx, metav1.ListOptions{LabelSelector: set.AsSelector().String()})
				if err != nil {
					l.Error(
						"failed to fetch pods for deployment",
						zap.String("label_selector", labelSelector),
						zap.String("deployment", fmt.Sprintf("%s.%s", dpl.Namespace, dpl.Name)),
					)
				}
				for _, pod := range dplPods.Items {
					if pod.CreationTimestamp.Time.Before(now) {
						continue
					}
					pods = append(pods, &pod)
				}
			}

			statefulsets, err := k8sClient.AppsV1().StatefulSets("").List(streamCtx, metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err != nil {
				l.Error("failed to fetch statefulsets", zap.String("label_selector", labelSelector))
			}
			for _, sfs := range statefulsets.Items {
				value, ok := sfs.Annotations[annotationSelectorKey]
				if !ok || value != annotationSelectorValue {
					continue
				}
				set := labels.Set(sfs.Spec.Selector.MatchLabels)
				sfsPods, err := k8sClient.CoreV1().Pods(sfs.Namespace).List(streamCtx, metav1.ListOptions{LabelSelector: set.AsSelector().String()})
				if err != nil {
					l.Error(
						"failed to fetch pods for statefulset",
						zap.String("label_selector", labelSelector),
						zap.String("statefulset", fmt.Sprintf("%s.%s", sfs.Namespace, sfs.Name)),
					)
				}
				for _, pod := range sfsPods.Items {
					if pod.CreationTimestamp.Time.Before(now) {
						continue
					}
					pods = append(pods, &pod)
				}
			}

			l.Info(fmt.Sprintf("streaming logs for %d pods", len(pods)),
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
