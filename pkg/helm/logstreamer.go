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

/*

A package for streaming logs from pods managed by resources in a helm release.

Supports Deployments and Statefulsets.
Does not support initContainers or standalone Pods.

*/

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

// this is the entrypoint
func (ls *LogStreamer) StreamPodLogs(ctx context.Context, pods []*corev1.Pod) error {
	// for each pod, and for each container in each pod, start a gorouting to stream the logs
	for _, pod := range pods {
		for _, con := range pod.Spec.Containers {
			if err := ls.streamPodContainerLog(ctx, pod, con.Name); err != nil {
				return err
			}
		}
	}
	return nil

}

// this actually does the work
func (ls *LogStreamer) streamPodContainerLog(ctx context.Context, pod *corev1.Pod, containerName string) error {
	// NOTE(fd): we use the "{pod.Namespace}.{pod.Name}.{containerName}" as the identifier
	podIdentifier := fmt.Sprintf("%s.%s.%s", pod.Namespace, pod.Name, containerName)

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
		// the namespace is generic at the top level but we fetch the pods by its namespace
		req := ls.clientset.CoreV1().Pods(pod.Namespace).GetLogs(
			pod.Name,
			&corev1.PodLogOptions{Container: containerName, Follow: true})

		logStream, err := req.Stream(podCtx)

		if err != nil {
			ls.l.Warn(
				fmt.Sprintf("Error creating k8s log stream for pod %s: %v", podIdentifier, err),
				zap.String("pod.metadata.name", pod.GetName()),
				zap.String("pod.metadata.namespace", pod.GetNamespace()),
				zap.String("pod.spec.container", containerName),
				zap.String("pod.name", pod.GetName()),
				zap.Any("pod.metadata.labels", pod.GetLabels()),
				zap.Any("pod.metadata.annotations", pod.GetAnnotations()),
			)
			return
		}
		defer logStream.Close()

		// NOTE(fd): ruthlessly stolen from:
		// > https://github.com/powertoolsdev/mono/blob/main/bins/runner/internal/jobs/deploy/job/monitor.go#L67-L79
		reader := bufio.NewReader(logStream)
		for {
			line, err := reader.ReadString('\n')
			// NOTE(jm): in the case of an EOF, we want to write any bytes that were copied into the buffer, to
			// ensure we do not leak any logs
			if err != nil {
				if errors.Is(err, io.EOF) && podCtx.Err() == nil {
					// we are done
					ls.l.Warn(
						fmt.Sprintf("Error reading k8s log stream for pod %s: %v", podIdentifier, err),
						zap.String("pod.metadata.name", pod.GetName()),
						zap.String("pod.metadata.namespace", pod.GetNamespace()),
						zap.String("pod.spec.container", containerName),
						zap.String("pod.name", pod.GetName()),
						zap.Any("pod.metadata.labels", pod.GetLabels()),
						zap.Any("pod.metadata.annotations", pod.GetAnnotations()),
					)
				}
				return
			}

			select {
			case <-podCtx.Done():
				return
			default:
				// write the log to the logger
				ls.l.Debug(line,
					zap.String("pod.metadata.name", pod.GetName()),
					zap.String("pod.metadata.namespace", pod.GetNamespace()),
					zap.String("pod.spec.container", containerName),
					zap.String("pod.name", pod.GetName()),
					zap.Any("pod.metadata.labels", pod.GetLabels()),
					zap.Any("pod.metadata.annotations", pod.GetAnnotations()),
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
