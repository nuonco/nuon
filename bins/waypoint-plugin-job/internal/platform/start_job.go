package platform

import (
	"context"
	"strings"

	"github.com/hashicorp/waypoint-plugin-sdk/component"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (p *Platform) startJob(ctx context.Context, clientset *kubernetes.Clientset, jobInfo *component.JobInfo) (*batchv1.Job, error) {
	// create the job config
	imageName := strings.Split(p.Cfg.ImageURL, "/")
	name := imageName[len(imageName)-1] + "-"
	namespace := jobInfo.Workspace
	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
			Namespace:    namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "job",
							Image: p.Cfg.ImageURL,
							Args:  []string{p.Cfg.Cmd},
							Env:   toEnv(p.Cfg.StaticEnvVars),
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

	// start the job
	jobsClient := clientset.BatchV1().Jobs(namespace)
	return jobsClient.Create(ctx, jobSpec, metav1.CreateOptions{})
}

func toEnv(statisEnvVars map[string]string) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	for key, val := range statisEnvVars {
		env = append(env, corev1.EnvVar{
			Name:  key,
			Value: val,
		})
	}

	return env
}
