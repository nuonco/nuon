package platform

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (p *Platform) startJob(ctx context.Context, clientset *kubernetes.Clientset) (*batchv1.Job, error) {
	// create the job config
	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Cfg.JobName,
			Namespace: p.Cfg.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: p.Cfg.ServiceAccount,
					Containers: []corev1.Container{
						{
							Name:    p.Cfg.ContainerName,
							Image:   fmt.Sprintf("%s:%s", p.Cfg.ImageURL, p.Cfg.Tag),
							Command: p.Cfg.Cmd,
							Args:    p.Cfg.Args,
							Env:     toEnv(p.Cfg.EnvVars),
						},
					},
					RestartPolicy: corev1.RestartPolicy(p.Cfg.RestartPolicy),
				},
			},
		},
	}

	// start the job
	jobsClient := clientset.BatchV1().Jobs(p.Cfg.Namespace)
	return jobsClient.Create(ctx, jobSpec, metav1.CreateOptions{})
}

func toEnv(envVars map[string]string) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	for key, val := range envVars {
		env = append(env, corev1.EnvVar{
			Name:  key,
			Value: val,
		})
	}

	return env
}
