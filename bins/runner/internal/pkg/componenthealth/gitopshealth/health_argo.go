package gitopshealth

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type nodePhase string

const (
	nodePending   nodePhase = "Pending"
	nodeRunning   nodePhase = "Running"
	nodeSucceeded nodePhase = "Succeeded"
	nodeFailed    nodePhase = "Failed"
	nodeError     nodePhase = "Error"
)

type argoWorkflow struct {
	Status struct {
		Phase   nodePhase
		Message string
	}
}

func getArgoWorkflowHealth(obj *unstructured.Unstructured) (*HealthStatus, error) {
	var wf argoWorkflow
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &wf)
	if err != nil {
		return nil, err
	}
	switch wf.Status.Phase {
	case "", nodePending, nodeRunning:
		return &HealthStatus{Status: HealthStatusProgressing, Message: wf.Status.Message}, nil
	case nodeSucceeded:
		return &HealthStatus{Status: HealthStatusHealthy, Message: wf.Status.Message}, nil
	case nodeFailed, nodeError:
		return &HealthStatus{Status: HealthStatusDegraded, Message: wf.Status.Message}, nil
	}
	return &HealthStatus{Status: HealthStatusUnknown, Message: wf.Status.Message}, nil
}
