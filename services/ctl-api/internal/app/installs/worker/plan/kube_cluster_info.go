package plan

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (p *Planner) getKubeClusterInfo(ctx workflow.Context, stack *app.InstallStack, state *state.State, cloudAuth *CloudAuth) (*kube.ClusterInfo, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	l.Info("checking sandbox outputs for kubernetes cluster info")
	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state data")
	}

	sandbox, sbOk := stateData["sandbox"]
	if sbOk {
		sb, ok := sandbox.(map[string]any)
		if ok {
			outputs, ok := sb["outputs"]
			if ok {
				outputsMap, ok := outputs.(map[string]any)
				if ok {
					res, clOk := outputsMap["cluster"]
					if !clOk || res == nil {
						l.Info("sandbox outputs do not include kubernetes cluster info, skipping")
						return nil, nil
					}
				}
			}
		}
	}

	l.Info("sandbox outputs contain kubernetes cluster info, parsing")
	obj := &kube.ClusterInfo{}
	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		if cloudAuth.AWS == nil {
			return nil, errors.New("aws auth information not provided")
		}
		obj = &kube.ClusterInfo{
			ID:       "{{.nuon.sandbox.outputs.cluster.name}}",
			Endpoint: "{{.nuon.sandbox.outputs.cluster.endpoint}}",
			CAData:   "{{.nuon.sandbox.outputs.cluster.certificate_authority_data}}",
			AWSAuth:  cloudAuth.AWS,
		}
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		if cloudAuth.Azure == nil {
			return nil, errors.New("azure auth information not provided")
		}
		obj = &kube.ClusterInfo{
			ID:        "{{.nuon.sandbox.outputs.cluster.name}}",
			Endpoint:  "{{.nuon.sandbox.outputs.cluster.host}}",
			CAData:    "{{.nuon.sandbox.outputs.cluster.cluster_ca_certificate}}",
			AzureAuth: cloudAuth.Azure,
		}
	case stack.InstallStackOutputs.GCPStackOutputs != nil:
		obj = &kube.ClusterInfo{
			ID:       "{{.nuon.sandbox.outputs.cluster.name}}",
			Endpoint: "{{.nuon.sandbox.outputs.cluster.endpoint}}",
			CAData:   "{{.nuon.sandbox.outputs.cluster.certificate_authority_data}}",
			GCPAuth:  cloudAuth.GCP,
		}
	}

	if err := render.RenderStruct(obj, stateData); err != nil {
		l.Error("error rendering cluster info",
			zap.Any("cluster-info", obj),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	l.Info("successfully parsed kubernetes cluster info, including in plan")
	return obj, nil
}
