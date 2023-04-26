package envtest

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers"
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	cr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	cr_log "sigs.k8s.io/controller-runtime/pkg/log"
	kzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	"time"
)

var s *SuiteState

var cancel context.CancelFunc
var testEnv *envtest.Environment
var testSupport *c.TestSupport

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {

	log := kzap.NewRaw(kzap.WriteTo(GinkgoWriter), kzap.UseDevMode(true))
	cr_log.SetLogger(zapr.NewLogger(log))
	GinkgoWriter.TeeTo(os.Stdout)

	s = &SuiteState{
		log: log,
	}

	pwd, err := os.Getwd()
	Expect(err).To(BeNil())
	root := pwd + "/../.."

	s.log.Sugar().Infow("path", "root", root)

	Expect(os.Setenv("TEST_ASSET_KUBE_APISERVER", root+"/testbin/bin/kube-apiserver")).To(Succeed())
	Expect(os.Setenv("TEST_ASSET_ETCD", root+"/testbin/bin/etcd")).To(Succeed())
	Expect(os.Setenv("TEST_ASSET_KUBECTL", root+"/testbin/bin/kubectl")).To(Succeed())

	s.ctx, cancel = context.WithCancel(context.TODO())

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{root + "/config/crd/resources"},
		ErrorIfCRDPathMissing: true,
		CRDInstallOptions: envtest.CRDInstallOptions{
			MaxTime: 60 * time.Second,
		},
	}
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	Expect(ar.AddToScheme(scheme.Scheme)).To(Succeed())

	s.k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(s.k8sClient).NotTo(BeNil())

	k8sManager, err := cr.NewManager(cfg, cr.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	testSupport = c.NewTestSupport(s.log, true)

	Expect(os.Setenv("REGISTRY_VERSION", "2.x")).To(Succeed())
	Expect(os.Setenv("OPERATOR_NAME", "apicurio-registry-operator")).To(Succeed())
	Expect(os.Setenv("REGISTRY_IMAGE_MEM", "quay.io/apicurio/apicurio-registry-mem:latest-snapshot")).To(Succeed())
	Expect(os.Setenv("REGISTRY_IMAGE_KAFKASQL", "quay.io/apicurio/apicurio-registry-kafkasql:latest-snapshot")).To(Succeed())
	Expect(os.Setenv("REGISTRY_IMAGE_SQL", "quay.io/apicurio/apicurio-registry-sql:latest-snapshot")).To(Succeed())

	reconciler, err := controllers.NewApicurioRegistryReconciler(k8sManager, s.log, testSupport)
	Expect(err).ToNot(HaveOccurred())
	Expect(reconciler).NotTo(BeNil())

	//+kubebuilder:scaffold:scheme
	go func() {
		defer GinkgoRecover()
		Expect(k8sManager.Start(s.ctx)).To(Succeed())
	}()
})

var _ = AfterSuite(func() {
	cancel()
	Expect(testEnv.Stop()).To(Succeed())
})
