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
	// these are the things we use for filtering the deployments and statefulsets
	annotationSelectorKey := "meta.helm.sh/release-name"
	annotationSelectorValue := chart.Metadata.Name
	labelSelector := "app.kubernetes.io/managed-by=Helm"

	// make k8s client
	k8sClient, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, err
	}

	// create log streamer
	streamCtx, cancelStreaming := context.WithCancel(ctx)
	defer cancelStreaming()

	// make a streamer
	streamer := NewLogStreamer(k8sClient, l)

	// the bulk of the work is here
	go streamLogs(streamCtx, cancelStreaming, streamer, k8sClient, labelSelector, annotationSelectorKey, annotationSelectorValue, l)

	// execute the upgrade
	rel, err := client.RunWithContext(ctx, chart, values)
	if err != nil {
		// NOTE(fd): i suspect if there is an error, the log streams may already be closed, but we're not taking any chances
		streamer.StopAllStreams()
		return nil, err
	}

	return rel, nil
}
