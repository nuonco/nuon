// Package kubernetes runs action containers as Kubernetes pods, one pod per
// step. It is the launcher.Launcher backend for control-plane workers that
// drive runnerless installs on EKS or GKE (any cloud-agnostic Kubernetes
// cluster) — unlike launcher.DockerLauncher, it does not need a local Docker
// daemon on the ctl-api host.
package kubernetes

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/nuonco/nuon/pkg/kube/clientcfg"
	"github.com/nuonco/nuon/pkg/runner/launcher"
)

const (
	// workspaceVolumeMount is the shared emptyDir mount both the init
	// container and the step container see. Every RunSpec.Mount.ContainerPath
	// the action handler asks for lives under this same directory
	// (containerWorkspaceMount in exec_image.go), so one emptyDir covers them.
	workspaceVolumeMount = "/nuon/work"
	readyVolumeMount     = "/nuon/ready"
	readyMarkerPath      = readyVolumeMount + "/.ready"

	initContainerName = "workspace-init"
	stepContainerName = "step"

	// initContainerImage only needs a shell and tar to wait for the ready
	// marker and receive the tar stream copyToPod sends; it never runs any of
	// the action's own code.
	defaultInitContainerImage = "busybox:1.36"

	podPollInterval = 500 * time.Millisecond
)

// Launcher implements launcher.Launcher by running each step as its own pod in
// a configured namespace, using in-cluster config (or KUBECONFIG for local
// development against a kind cluster). See pkg/runner/launcher/launcher.go for
// the interface this satisfies.
type Launcher struct {
	clientset *kubernetes.Clientset
	restCfg   *rest.Config
	namespace string
	initImage string
	l         *zap.Logger
}

var _ launcher.Launcher = (*Launcher)(nil)

// Params configures the Kubernetes launcher.
type Params struct {
	// Namespace is the namespace every action pod and secret is created in.
	// Required: ctl-api's RBAC Role only grants pods/secrets access scoped to
	// this one namespace (see manifests under deploy/k8s).
	Namespace string

	// InitContainerImage overrides the image used for the workspace-init
	// container. Defaults to busybox.
	InitContainerImage string

	L *zap.Logger
}

// New builds a Kubernetes launcher. It resolves cluster credentials via
// pkg/kube/clientcfg.GetKubeConfig, which prefers in-cluster config and falls
// back to KUBECONFIG / ~/.kube/config for local development (e.g. against a
// kind cluster).
func New(params Params) (*Launcher, error) {
	if params.Namespace == "" {
		return nil, errors.New("kubernetes launcher requires a namespace")
	}

	restCfg, err := clientcfg.GetKubeConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get kube config")
	}
	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build kube clientset")
	}

	initImage := params.InitContainerImage
	if initImage == "" {
		initImage = defaultInitContainerImage
	}

	l := params.L
	if l == nil {
		l = zap.NewNop()
	}

	return &Launcher{
		clientset: clientset,
		restCfg:   restCfg,
		namespace: params.Namespace,
		initImage: initImage,
		l:         l,
	}, nil
}

// Prepare is a no-op for the Kubernetes backend: every pod pulls its own
// image from the node it lands on, so there is no shared host-level pull to
// lease the way DockerLauncher leases a host image cache.
func (k *Launcher) Prepare(ctx context.Context, spec launcher.PrepareSpec) error {
	return nil
}

// Release is a no-op: Run already deletes the pod and its secret before
// returning, so there is nothing left to release by lease ID.
func (k *Launcher) Release(leaseID string) {}

// Run creates one pod for the step, waits for its init container to come up,
// copies every mount's contents in over an exec'd tar, signals the init
// container to let the step container start, streams the step container's
// logs to spec.Stdout, and waits for the pod to finish. The pod and its
// per-step secret are always deleted before Run returns, including when ctx
// is canceled or times out.
func (k *Launcher) Run(ctx context.Context, spec launcher.RunSpec) error {
	name := podName(spec.ContainerName)
	secretName := name + "-env"
	l := k.l.With(zap.String("action.pod", name), zap.String("action.namespace", k.namespace))

	if err := k.createSecret(ctx, secretName, spec.Env, spec.Labels); err != nil {
		return errors.Wrap(err, "unable to create action pod secret")
	}
	defer k.deleteSecret(secretName)

	pod := k.buildPod(name, secretName, spec)
	if _, err := k.clientset.CoreV1().Pods(k.namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return errors.Wrap(err, "unable to create action pod")
	}
	defer k.deletePod(name)

	if err := k.waitForInitContainerRunning(ctx, name); err != nil {
		return errors.Wrap(err, "unable to wait for action pod init container")
	}

	for _, m := range spec.Mounts {
		if err := k.copyToPod(ctx, name, m.HostPath, m.ContainerPath); err != nil {
			return errors.Wrapf(err, "unable to copy mount into action pod at %s", m.ContainerPath)
		}
	}

	if err := k.signalWorkspaceReady(ctx, name); err != nil {
		return errors.Wrap(err, "unable to signal action pod workspace ready")
	}

	// Logs stream best-effort: a pod that fails before the step container ever
	// starts (e.g. ImagePullBackOff) has nothing to stream, and the exit-code
	// check below is what actually surfaces the failure.
	if err := k.streamLogs(ctx, name, spec.Stdout); err != nil {
		l.Warn("unable to stream action pod logs", zap.Error(err))
	}

	return k.waitForCompletion(ctx, name)
}

func (k *Launcher) buildPod(name, secretName string, spec launcher.RunSpec) *corev1.Pod {
	labels := map[string]string{"app.kubernetes.io/managed-by": "nuon-actions"}
	for key, val := range spec.Labels {
		labels[sanitizeLabelKey(key)] = sanitizeLabelValue(val)
	}

	automount := false
	runAsNonRoot := true
	runAsUser := int64(65532)
	allowPrivEsc := false

	securityContext := &corev1.SecurityContext{
		RunAsNonRoot:             &runAsNonRoot,
		RunAsUser:                &runAsUser,
		AllowPrivilegeEscalation: &allowPrivEsc,
	}

	workspaceVolumeMounts := []corev1.VolumeMount{{Name: "workspace", MountPath: workspaceVolumeMount}}

	initContainer := corev1.Container{
		Name:  initContainerName,
		Image: k.initImage,
		// Waits for copyToPod's tar exec to finish and signalWorkspaceReady to
		// write the marker, then exits so the step container starts.
		Command: []string{"sh", "-c", fmt.Sprintf("mkdir -p %s; until [ -f %s ]; do sleep 0.2; done", readyVolumeMount, readyMarkerPath)},
		VolumeMounts: append(workspaceVolumeMounts,
			corev1.VolumeMount{Name: "ready", MountPath: readyVolumeMount},
		),
		SecurityContext: securityContext,
	}

	container := corev1.Container{
		Name:            stepContainerName,
		Image:           spec.Image,
		WorkingDir:      spec.Workdir,
		VolumeMounts:    workspaceVolumeMounts,
		SecurityContext: securityContext,
		EnvFrom: []corev1.EnvFromSource{{
			SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: secretName}},
		}},
	}
	if len(spec.Command) > 0 {
		container.Command = []string{spec.Command[0]}
		container.Args = spec.Command[1:]
	}
	if limits := resourceLimits(spec); len(limits) > 0 {
		container.Resources = corev1.ResourceRequirements{Limits: limits}
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: k.namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy:                corev1.RestartPolicyNever,
			AutomountServiceAccountToken: &automount,
			InitContainers:               []corev1.Container{initContainer},
			Containers:                   []corev1.Container{container},
			Volumes: []corev1.Volume{
				{Name: "workspace", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "ready", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
		},
	}
}

// resourceLimits maps RunSpec's docker-flavored limits onto the nearest
// Kubernetes equivalents. CPUShares (a cgroup v1 relative weight) and
// PidsLimit have no direct Kubernetes resource-request analog, so they are
// deliberately dropped here rather than mapped to a misleading CPU quota —
// see the gaps note in the task report.
func resourceLimits(spec launcher.RunSpec) corev1.ResourceList {
	limits := corev1.ResourceList{}
	if spec.Memory != "" {
		if q, err := resource.ParseQuantity(spec.Memory); err == nil {
			limits[corev1.ResourceMemory] = q
		}
	}
	return limits
}

func (k *Launcher) createSecret(ctx context.Context, name string, env map[string]string, labels map[string]string) error {
	data := map[string][]byte{}
	for key, val := range env {
		data[key] = []byte(val)
	}
	podLabels := map[string]string{"app.kubernetes.io/managed-by": "nuon-actions"}
	for key, val := range labels {
		podLabels[sanitizeLabelKey(key)] = sanitizeLabelValue(val)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: k.namespace, Labels: podLabels},
		Type:       corev1.SecretTypeOpaque,
		Data:       data,
	}
	_, err := k.clientset.CoreV1().Secrets(k.namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

func (k *Launcher) deleteSecret(name string) {
	// Use a background context: this always runs from a defer, including when
	// the caller's ctx is already canceled or timed out, and cleanup must
	// still happen.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := k.clientset.CoreV1().Secrets(k.namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		k.l.Warn("unable to delete action pod secret", zap.String("secret", name), zap.Error(err))
	}
}

func (k *Launcher) deletePod(name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	grace := int64(0)
	if err := k.clientset.CoreV1().Pods(k.namespace).Delete(ctx, name, metav1.DeleteOptions{GracePeriodSeconds: &grace}); err != nil {
		k.l.Warn("unable to delete action pod", zap.String("pod", name), zap.Error(err))
	}
}

func (k *Launcher) waitForInitContainerRunning(ctx context.Context, name string) error {
	ticker := time.NewTicker(podPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for _, cs := range pod.Status.InitContainerStatuses {
			if cs.Name != initContainerName {
				continue
			}
			if cs.State.Running != nil {
				return nil
			}
			if cs.State.Terminated != nil {
				return fmt.Errorf("action pod init container exited early: %s", cs.State.Terminated.Reason)
			}
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "ImagePullBackOff" {
				return fmt.Errorf("action pod init container image pull failed: %s", cs.State.Waiting.Message)
			}
		}
	}
}

func (k *Launcher) waitForCompletion(ctx context.Context, name string) error {
	ticker := time.NewTicker(podPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		switch pod.Status.Phase {
		case corev1.PodSucceeded:
			return nil
		case corev1.PodFailed:
			return fmt.Errorf("action pod failed: %s", stepExitMessage(pod))
		}
	}
}

func stepExitMessage(pod *corev1.Pod) string {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name != stepContainerName {
			continue
		}
		if cs.State.Terminated != nil {
			return fmt.Sprintf("exit code %d: %s", cs.State.Terminated.ExitCode, cs.State.Terminated.Reason)
		}
	}
	return pod.Status.Message
}

// copyToPod streams hostPath as a tar archive into containerPath inside the
// init container, the same mechanism `kubectl cp` uses (exec a `tar -x` and
// stream the archive over its stdin).
func (k *Launcher) copyToPod(ctx context.Context, podName, hostPath, containerPath string) error {
	reader, writer := io.Pipe()
	tarErrCh := make(chan error, 1)
	go func() {
		tarErrCh <- tarDirectory(hostPath, writer)
		writer.Close()
	}()

	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(k.namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: initContainerName,
			Command:   []string{"tar", "-xf", "-", "-C", containerPath},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.restCfg, "POST", req.URL())
	if err != nil {
		return errors.Wrap(err, "unable to build exec stream")
	}

	var stdout, stderr bytes.Buffer
	streamErr := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  reader,
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if tarErr := <-tarErrCh; tarErr != nil {
		return errors.Wrap(tarErr, "unable to build tar archive")
	}
	if streamErr != nil {
		return errors.Wrapf(streamErr, "tar extract failed (stderr: %s)", strings.TrimSpace(stderr.String()))
	}
	return nil
}

func (k *Launcher) signalWorkspaceReady(ctx context.Context, podName string) error {
	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(k.namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: initContainerName,
			Command:   []string{"sh", "-c", fmt.Sprintf("mkdir -p %s && touch %s", readyVolumeMount, readyMarkerPath)},
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.restCfg, "POST", req.URL())
	if err != nil {
		return errors.Wrap(err, "unable to build exec stream")
	}
	var stderr bytes.Buffer
	if err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{Stderr: &stderr}); err != nil {
		return errors.Wrapf(err, "ready-marker exec failed (stderr: %s)", strings.TrimSpace(stderr.String()))
	}
	return nil
}

// streamLogs follows the step container's combined stdout/stderr (the
// Kubernetes logs API does not separate the two streams) until the container
// exits, writing everything to dst.
func (k *Launcher) streamLogs(ctx context.Context, podName string, dst io.Writer) error {
	if dst == nil {
		dst = io.Discard
	}

	// The step container may still be waiting on the init container, so retry
	// briefly until logs are available instead of failing on the first
	// "container not started" error.
	var stream io.ReadCloser
	var err error
	deadline := time.Now().Add(2 * time.Minute)
	for {
		req := k.clientset.CoreV1().Pods(k.namespace).GetLogs(podName, &corev1.PodLogOptions{
			Container: stepContainerName,
			Follow:    true,
		})
		stream, err = req.Stream(ctx)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return errors.Wrap(err, "unable to open action pod log stream")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(podPollInterval):
		}
	}
	defer stream.Close()

	_, err = io.Copy(dst, stream)
	return err
}

func tarDirectory(root string, w io.Writer) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		if info.IsDir() {
			hdr.Name += "/"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
}

// podName turns a container name (which may contain characters invalid in a
// Kubernetes object name) into a valid, lowercase, 63-char-max pod name.
func podName(containerName string) string {
	name := strings.ToLower(containerName)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, name)
	name = strings.Trim(name, "-")
	if name == "" {
		name = "nuon-action"
	}
	if len(name) > 63 {
		name = name[:63]
	}
	return strings.Trim(name, "-")
}

func sanitizeLabelKey(key string) string {
	return "nuon.co/" + podName(key)
}

func sanitizeLabelValue(val string) string {
	v := podName(val)
	if len(v) > 63 {
		v = v[:63]
	}
	return v
}
