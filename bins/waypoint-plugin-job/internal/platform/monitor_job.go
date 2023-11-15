package platform

import (
	"context"
	"fmt"
	"io"

	batchv1 "k8s.io/api/batch/v1"
	// "k8s.io/apimachinery/pkg/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// monitorJob monitors the job status and tails the logs. It assumes there is only one pod.
func (p *Platform) monitorJob(ctx context.Context, clientSet *kubernetes.Clientset, job *batchv1.Job) error {
	// get pod (once it's been created)
	p.logger.Debug("getting pod")
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
			p.logger.Debug("pod created",
				"name", pod.GetName(),
			)
			break
			// case watch.Error:
			// 	// item := event.Object.(*api.Status)
			// 	p.logger.Debug("error event")
		}
	}
	if pod == nil {
		return fmt.Errorf("failed to get pod for job")
	}
	p.logger.Debug("got pod")

	// get the log stream
	p.logger.Debug("getting log stream")
	logStream, err := clientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{}).Stream(ctx)
	if err != nil {
		return err
	}
	defer logStream.Close()
	p.logger.Debug("got log stream")

	// read from log stream
	p.logger.Debug("reading logs from pod")
	for {
		buf := make([]byte, 2000)
		_, err := logStream.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		p.logger.Info(string(buf))
	}

	return nil
}
