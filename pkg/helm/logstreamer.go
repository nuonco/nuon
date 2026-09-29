package helm

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
)

type LogStreamer struct {
	clientset *kubernetes.Clientset
	wg        sync.WaitGroup
	mu        sync.Mutex
	streams   map[string]context.CancelFunc
	l         *zap.Logger
}

func NewLogStreamer(clientset *kubernetes.Clientset, l *zap.Logger) *LogStreamer {
	return &LogStreamer{
		clientset: clientset,
		streams:   make(map[string]context.CancelFunc),
		l:         l,
	}
}

func getPodState(pod *corev1.Pod) string {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil {
			if cs.State.Waiting.Reason != "" {
				return cs.State.Waiting.Reason
			}
			return "Waiting"
		}
		if cs.State.Terminated != nil {
			if cs.State.Terminated.Reason != "" {
				return cs.State.Terminated.Reason
			}
			return "Terminated"
		}
	}
	if len(pod.Status.ContainerStatuses) == 0 {
		return string(pod.Status.Phase)
	}
	return "Running"
}

func podStateSummary(pods []*corev1.Pod) (string, map[string][]string) {
	byState := map[string][]string{}
	for _, pod := range pods {
		state := getPodState(pod)
		byState[state] = append(byState[state], pod.Name)
	}
	var summary string
	for state, names := range byState {
		if summary != "" {
			summary += " "
		}
		summary += fmt.Sprintf("%d %s", len(names), state)
	}
	return summary, byState
}

func (ls *LogStreamer) StreamPodLogs(ctx context.Context, pods []*corev1.Pod) error {
	summary, counts := podStateSummary(pods)

	var ready []*corev1.Pod
	for _, pod := range pods {
		if getPodState(pod) == "Running" {
			ready = append(ready, pod)
		}
	}

	if len(ready) < len(pods) {
		ls.l.Info(summary, zap.Any("pod_states", counts))
	}

	for _, pod := range ready {
		for _, con := range pod.Spec.Containers {
			if err := ls.streamPodContainerLog(ctx, pod, con.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ls *LogStreamer) streamPodContainerLog(ctx context.Context, pod *corev1.Pod, containerName string) error {
	podIdentifier := fmt.Sprintf("%s.%s.%s", pod.Namespace, pod.Name, containerName)

	ls.l.Info(
		fmt.Sprintf("starting log stream for pod %s", podIdentifier),
		zap.String("pod.metadata.name", pod.GetName()),
		zap.String("pod.metadata.namespace", pod.GetNamespace()),
		zap.String("pod.spec.container", containerName),
		zap.Any("pod.metadata.labels", pod.GetLabels()),
		zap.Any("pod.metadata.annotations", pod.GetAnnotations()),
	)

	podCtx, cancel := context.WithCancel(ctx)

	ls.mu.Lock()
	ls.streams[podIdentifier] = cancel
	ls.mu.Unlock()

	ls.wg.Add(1)

	go func() {
		defer ls.wg.Done()
		defer func() {
			ls.mu.Lock()
			delete(ls.streams, podIdentifier)
			ls.mu.Unlock()
		}()
		req := ls.clientset.CoreV1().Pods(pod.Namespace).GetLogs(
			pod.Name,
			&corev1.PodLogOptions{Container: containerName, Follow: true})

		logStream, err := req.Stream(podCtx)

		if err != nil {
			return
		}
		defer logStream.Close()

		reader := bufio.NewReader(logStream)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) && podCtx.Err() == nil {
					ls.l.Warn(
						fmt.Sprintf("Error reading k8s log stream for pod %s: %v", podIdentifier, err),
						zap.String("pod.metadata.name", pod.GetName()),
						zap.String("pod.metadata.namespace", pod.GetNamespace()),
						zap.String("pod.spec.container", containerName),
					)
				}
				return
			}

			select {
			case <-podCtx.Done():
				return
			default:
				ls.l.Info(line,
					zap.String("pod.metadata.name", pod.GetName()),
					zap.String("pod.metadata.namespace", pod.GetNamespace()),
					zap.String("pod.spec.container", containerName),
				)
			}
		}

	}()

	return nil
}

func (ls *LogStreamer) StopStream(podIdentifier string) {
	ls.mu.Lock()
	if cancel, exists := ls.streams[podIdentifier]; exists {
		cancel()
	}
	ls.mu.Unlock()
}

func (ls *LogStreamer) StopAllStreams() {
	ls.mu.Lock()
	for _, cancel := range ls.streams {
		cancel()
	}
	ls.mu.Unlock()
}

func (ls *LogStreamer) Wait() {
	ls.wg.Wait()
}
