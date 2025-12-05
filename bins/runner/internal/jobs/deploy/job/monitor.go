package job

import (
	"context"
	"errors"
	"fmt"
	"io"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"

	// "k8s.io/apimachinery/pkg/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// pollJob monitors the job status and tails the logs. It assumes there is only one pod.
func (p *handler) pollJob(ctx context.Context, clientSet *kubernetes.Clientset, job *batchv1.Job) error {
	// get pod (once it's been created)
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Debug("getting pod")
	timeout := int64(60)
	podWatcher, err := clientSet.CoreV1().Pods(p.Cfg.Namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector:  fmt.Sprintf("job-name=%s", job.Name),
		TimeoutSeconds: &timeout,
	})
	if err != nil {
		return err
	}
	var pod *corev1.Pod
	for event := range podWatcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			pod = event.Object.(*corev1.Pod)
			l.Debug("pod created",
				zap.String("name", pod.GetName()),
			)
			break
			// case watch.Error:
			//	// item := event.Object.(*api.Status)
			//	p.logger.Debug("error event")
		}
	}
	if pod == nil {
		return fmt.Errorf("failed to get pod for job")
	}
	l.Debug("got pod")

	// get the log stream
	l.Debug("getting log stream")
	logStream, err := clientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{}).Stream(ctx)
	if err != nil {
		return err
	}
	defer logStream.Close()
	l.Debug("got log stream")

	// read from log stream
	l.Debug("reading logs from pod")
	for {
		buf := make([]byte, 2000)
		written, err := logStream.Read(buf)
		// NOTE(jm): in the case of an EOF, we want to write any bytes that were copied into the buffer, to
		// ensure we do not leak any logs
		l.Info(string(buf[:written]))
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}
