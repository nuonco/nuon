//go:build integration

package teardown_test

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/helm/waypoint"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/server"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/teardown"
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

var _ = Describe("DestroyNamespace", func() {
	var (
		namespace string
		a         *teardown.Activities
		req       teardown.DestroyNamespaceRequest
		e         *testsuite.TestActivityEnvironment
	)

	BeforeEach(func() {
		namespace = uuid.New().String()
		a = teardown.NewActivities()
		a.Kubeconfig = cfg
		testSuite := &testsuite.WorkflowTestSuite{}
		e = testSuite.NewTestActivityEnvironment()
		e.RegisterActivity(a)

		req = teardown.DestroyNamespaceRequest{
			NamespaceName: namespace,
		}

		ns := &corev1.Namespace{ObjectMeta: apimetav1.ObjectMeta{Name: namespace}}
		err := k8sClient.Create(ctx, ns)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("When destroying a namespace", func() {
		It("Should destroy namespace", func() {
			resp, err := e.ExecuteActivity(a.DestroyNamespace, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
		})
	})
})

var _ = Describe("UninstallWaypoint", func() {
	var (
		namespace   string = "default"
		releaseName string = "test"
		a           *teardown.Activities
		req         teardown.UninstallWaypointRequest
		e           *testsuite.TestActivityEnvironment
	)

	BeforeEach(func() {
		namespace = uuid.New().String()
		a = teardown.NewActivities()
		a.Kubeconfig = cfg
		testSuite := &testsuite.WorkflowTestSuite{}
		e = testSuite.NewTestActivityEnvironment()
		e.RegisterActivity(a)

		req = teardown.UninstallWaypointRequest{
			Namespace:   namespace,
			ReleaseName: releaseName,
		}

	})

	Context("When uninstalling waypoint", func() {

		It("Should not error if waypoint is not installed", func() {
			resp, err := e.ExecuteActivity(a.UninstallWaypoint, req)
			Expect(err).NotTo(HaveOccurred())
			// TODO(jdt): this assertion is pretty weak
			Expect(resp).NotTo(BeNil())
		})

		Context("When waypoint is installed", func() {
			BeforeEach(func() {
				sa := server.NewActivities()
				sa.Kubeconfig = cfg
				e.RegisterActivity(sa)

				_, err := e.ExecuteActivity(sa.InstallWaypointServer, server.InstallWaypointServerRequest{
					Namespace:       namespace,
					ReleaseName:     releaseName,
					Chart:           &waypoint.DefaultChart,
					CreateNamespace: true,
					ClusterInfo: kube.ClusterInfo{
						ID:             uuid.NewString(),
						CAData:         uuid.NewString(),
						Endpoint:       "abc.eks.amazonaws.com",
						TrustedRoleARN: "abc://arn",
					},
				})
				Expect(err).NotTo(HaveOccurred())

			})

			It("Should uninstall waypoint", func() {
				resp, err := e.ExecuteActivity(a.UninstallWaypoint, req)
				Expect(err).NotTo(HaveOccurred())
				// TODO(jdt): this assertion is pretty weak
				Expect(resp).NotTo(BeNil())
			})
		})
	})

	AfterEach(func() {
		ns := &corev1.Namespace{ObjectMeta: apimetav1.ObjectMeta{Name: namespace}}
		k8sClient.Delete(ctx, ns)
	})

})
