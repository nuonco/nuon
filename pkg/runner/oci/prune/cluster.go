package prune

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/kube/clientcfg"
)

const listPageSize = 500

type InUse struct {
	Tags    []string
	Digests []string
}

func ScanCluster(ctx context.Context, ci *kube.ClusterInfo, repositoryURL string) (*InUse, error) {
	restCfg, err := clientcfg.ConfigForCluster(ctx, ci)
	if err != nil {
		return nil, fmt.Errorf("unable to build kube config: %w", err)
	}
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create kube client: %w", err)
	}

	c := &collector{repo: repositoryURL, tags: map[string]bool{}, digests: map[string]bool{}}

	if err := paginate(ctx, func(opts metav1.ListOptions) (string, error) {
		list, err := client.CoreV1().Pods("").List(ctx, opts)
		if err != nil {
			return "", fmt.Errorf("list pods: %w", err)
		}
		for _, p := range list.Items {
			c.addPodSpec(p.Spec)
			for _, statuses := range [][]corev1.ContainerStatus{p.Status.ContainerStatuses, p.Status.InitContainerStatuses, p.Status.EphemeralContainerStatuses} {
				for _, s := range statuses {
					c.add(s.Image)
					c.add(s.ImageID)
				}
			}
		}
		return list.Continue, nil
	}); err != nil {
		return nil, err
	}

	if err := paginate(ctx, func(opts metav1.ListOptions) (string, error) {
		list, err := client.AppsV1().Deployments("").List(ctx, opts)
		if err != nil {
			return "", fmt.Errorf("list deployments: %w", err)
		}
		for _, o := range list.Items {
			c.addPodSpec(o.Spec.Template.Spec)
		}
		return list.Continue, nil
	}); err != nil {
		return nil, err
	}

	if err := paginate(ctx, func(opts metav1.ListOptions) (string, error) {
		list, err := client.AppsV1().ReplicaSets("").List(ctx, opts)
		if err != nil {
			return "", fmt.Errorf("list replicasets: %w", err)
		}
		for _, o := range list.Items {
			c.addPodSpec(o.Spec.Template.Spec)
		}
		return list.Continue, nil
	}); err != nil {
		return nil, err
	}

	if err := paginate(ctx, func(opts metav1.ListOptions) (string, error) {
		list, err := client.AppsV1().StatefulSets("").List(ctx, opts)
		if err != nil {
			return "", fmt.Errorf("list statefulsets: %w", err)
		}
		for _, o := range list.Items {
			c.addPodSpec(o.Spec.Template.Spec)
		}
		return list.Continue, nil
	}); err != nil {
		return nil, err
	}

	if err := paginate(ctx, func(opts metav1.ListOptions) (string, error) {
		list, err := client.AppsV1().DaemonSets("").List(ctx, opts)
		if err != nil {
			return "", fmt.Errorf("list daemonsets: %w", err)
		}
		for _, o := range list.Items {
			c.addPodSpec(o.Spec.Template.Spec)
		}
		return list.Continue, nil
	}); err != nil {
		return nil, err
	}

	if err := paginate(ctx, func(opts metav1.ListOptions) (string, error) {
		list, err := client.BatchV1().Jobs("").List(ctx, opts)
		if err != nil {
			return "", fmt.Errorf("list jobs: %w", err)
		}
		for _, o := range list.Items {
			c.addPodSpec(o.Spec.Template.Spec)
		}
		return list.Continue, nil
	}); err != nil {
		return nil, err
	}

	if err := paginate(ctx, func(opts metav1.ListOptions) (string, error) {
		list, err := client.BatchV1().CronJobs("").List(ctx, opts)
		if err != nil {
			return "", fmt.Errorf("list cronjobs: %w", err)
		}
		for _, o := range list.Items {
			c.addPodSpec(o.Spec.JobTemplate.Spec.Template.Spec)
		}
		return list.Continue, nil
	}); err != nil {
		return nil, err
	}

	return c.result(), nil
}

func paginate(ctx context.Context, fn func(metav1.ListOptions) (string, error)) error {
	opts := metav1.ListOptions{Limit: listPageSize}
	for {
		next, err := fn(opts)
		if err != nil {
			return err
		}
		if next == "" {
			return nil
		}
		opts.Continue = next
	}
}

type collector struct {
	repo    string
	tags    map[string]bool
	digests map[string]bool
}

func (c *collector) addPodSpec(spec corev1.PodSpec) {
	for _, ctr := range spec.Containers {
		c.add(ctr.Image)
	}
	for _, ctr := range spec.InitContainers {
		c.add(ctr.Image)
	}
	for _, ctr := range spec.EphemeralContainers {
		c.add(ctr.Image)
	}
}

func (c *collector) add(ref string) {
	tag, digest, ok := parseImageRef(c.repo, ref)
	if !ok {
		return
	}
	if tag != "" {
		c.tags[tag] = true
	}
	if digest != "" {
		c.digests[digest] = true
	}
}

func (c *collector) result() *InUse {
	out := &InUse{}
	for t := range c.tags {
		out.Tags = append(out.Tags, t)
	}
	for d := range c.digests {
		out.Digests = append(out.Digests, d)
	}
	return out
}

func parseImageRef(repo, ref string) (string, string, bool) {
	ref = strings.TrimPrefix(ref, "docker-pullable://")
	ref = strings.TrimPrefix(ref, "docker://")
	rest, ok := strings.CutPrefix(ref, repo)
	if !ok || rest == "" {
		return "", "", false
	}

	var tag, digest string
	if before, after, found := strings.Cut(rest, "@"); found {
		digest = after
		rest = before
	}
	if rest != "" {
		if !strings.HasPrefix(rest, ":") {
			return "", "", false
		}
		tag = rest[1:]
	}
	if tag == "" && digest == "" {
		return "", "", false
	}
	return tag, digest, true
}
