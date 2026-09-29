package helm

import (
	"context"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	release "helm.sh/helm/v4/pkg/release/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func HelmInstallWithLogStreaming(
	ctx context.Context,
	client *action.Install, chart *chart.Chart, values map[string]interface{},
	kubeCfg *rest.Config,
	l *zap.Logger,
) (*release.Release, error) {
	annotationSelectorKey := "meta.helm.sh/release-name"
	annotationSelectorValue := chart.Metadata.Name
	labelSelector := "app.kubernetes.io/managed-by=Helm"

	k8sClient, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, err
	}

	streamCtx, cancelStreaming := context.WithCancel(ctx)
	defer cancelStreaming()

	streamer := NewLogStreamer(k8sClient, l)

	go streamLogs(streamCtx, cancelStreaming, streamer, k8sClient, labelSelector, annotationSelectorKey, annotationSelectorValue, l)

	rel, err := client.RunWithContext(ctx, chart, values)
	if err != nil {
		streamer.StopAllStreams()
		return nil, err
	}

	return rel, nil
}
