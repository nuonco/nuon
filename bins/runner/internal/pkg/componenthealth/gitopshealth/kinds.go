// Package gitopshealth is vendored from
// github.com/argoproj/gitops-engine@v0.7.3/pkg/health (Apache-2.0).
package gitopshealth

const (
	DeploymentKind              = "Deployment"
	ReplicaSetKind              = "ReplicaSet"
	StatefulSetKind             = "StatefulSet"
	DaemonSetKind               = "DaemonSet"
	IngressKind                 = "Ingress"
	JobKind                     = "Job"
	PersistentVolumeClaimKind   = "PersistentVolumeClaim"
	PodKind                     = "Pod"
	ServiceKind                 = "Service"
	HorizontalPodAutoscalerKind = "HorizontalPodAutoscaler"
)
