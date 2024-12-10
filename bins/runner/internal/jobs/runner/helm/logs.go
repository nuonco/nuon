package helm

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/powertoolsdev/mono/pkg/zapwriter"
)

const (
	podCheckTimeout time.Duration = time.Second * 2
)

func (h *handler) getPodContainerLogs(ctx context.Context, l *zap.Logger, kubeClient kubernetes.Interface, podName, containerName, namespace string) error {
	req := kubeClient.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		Follow:    true,
		Container: containerName,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	writer := zapwriter.New(l, zapcore.ErrorLevel, fmt.Sprintf("%s: ", containerName))
	_, err = io.Copy(writer, stream)
	if err != nil {
		if err == io.EOF {
			return nil
		}

		return err
	}

	return nil
}

func (h *handler) getPods(ctx context.Context, l *zap.Logger, kubeClient kubernetes.Interface, namespace string) error {
	knownPods := make(map[string]struct{}, 0)
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// checking for pods
		pods, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "unable to get pods")
		}

		for _, podItem := range pods.Items {
			podName := podItem.Name

			_, ok := knownPods[podName]
			if ok {
				continue
			}

			pod, err := kubeClient.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, "unable to get pod")
			}

			for _, container := range pod.Spec.Containers {
				wg.Add(1)
				go func() {
					defer wg.Done()

					if err := h.getPodContainerLogs(ctx, l, kubeClient, namespace, pod.Name, container.Name); err != nil {
						if !errors.Is(err, context.Canceled) {
							l.Error("unable to get container logs", zap.Error(err))
						}
					}
				}()
			}
		}
	}
}
