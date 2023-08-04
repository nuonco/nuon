//go:build integration

package runner_test

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/waypoint/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/services/workers-installs/internal/runner"
	"go.temporal.io/sdk/testsuite"

	"testing"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client // You'll be using this client in your tests.
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
}, 60)

var _ = AfterSuite(func() {
	cancel()

	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("InstallWaypoint", func() {
	var (
		namespace string = "default"
		a         *runner.Activities
		req       runner.InstallWaypointRequest
		e         *testsuite.TestActivityEnvironment
	)

	BeforeEach(func() {
		namespace = uuid.New().String()
		a = runner.NewActivities(workers.Config{})
		a.Kubeconfig = cfg
		testSuite := &testsuite.WorkflowTestSuite{}
		e = testSuite.NewTestActivityEnvironment()
		e.RegisterActivity(a)

		req = runner.InstallWaypointRequest{
			InstallID:   uuid.NewString(),
			Namespace:   namespace,
			ReleaseName: "test",
			Chart:       &waypoint.DefaultChart,
			Atomic:      false,
			ClusterInfo: kube.ClusterInfo{
				ID:             "cluster-id",
				Endpoint:       "endpoint",
				CAData:         "ca-data",
				TrustedRoleARN: "arn",
			},
			RunnerConfig: runner.RunnerConfig{
				ID:            namespace,
				Cookie:        "cookie",
				ServerAddr:    "addr",
				OdrIAMRoleArn: "arn",
			},
			CreateNamespace: true,
		}
	})

	Context("When installing waypoint", func() {
		It("Should install waypoint", func() {
			resp, err := e.ExecuteActivity(a.InstallWaypoint, req)
			Expect(err).NotTo(HaveOccurred())
			// TODO(jdt): this assertion is pretty weak
			Expect(resp).NotTo(BeNil())
		})

		It("Should fail with invalid version", func() {
			req.Chart.Version = "doesnotexist"
			_, err := e.ExecuteActivity(a.InstallWaypoint, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("version %q not found", req.Chart.Version))
		})
	})

	AfterEach(func() {
		ns := &corev1.Namespace{ObjectMeta: apimetav1.ObjectMeta{Name: namespace}}
		k8sClient.Delete(ctx, ns)
	})
})
