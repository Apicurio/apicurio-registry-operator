package envtest

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("java options environment variables", Ordered, func() {

	var registryKey types.NamespacedName
	var deploymentKey types.NamespacedName

	const testNamespace = "java-options-env-test-namespace"
	const registryName = "test"

	BeforeAll(func() {
		// Speed the tests up
		testSupport.SetMockCanMakeHTTPRequestToOperand(testNamespace, true)
		testSupport.SetMockOperandMetricsReportReady(testNamespace, true)
		ns := &core.Namespace{
			ObjectMeta: meta.ObjectMeta{
				Name: testNamespace,
			},
		}
		Expect(s.k8sClient.Create(context.TODO(), ns)).To(Succeed())
		registry := &ar.ApicurioRegistry{
			ObjectMeta: meta.ObjectMeta{
				Name:      registryName,
				Namespace: ns.ObjectMeta.Name,
			},
			Spec: ar.ApicurioRegistrySpec{
				Configuration: ar.ApicurioRegistrySpecConfiguration{
					Persistence: "sql",
				},
			},
		}
		Expect(s.k8sClient.Create(s.ctx, registry)).To(Succeed())
		registryKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name}
		deploymentKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name + "-deployment"}
	})

	It("should be set in the deployment", func() {
		Eventually(func() error {
			registry := &ar.ApicurioRegistry{}
			Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
			registry.Spec.Configuration.Env = []core.EnvVar{
				{
					Name:  "JAVA_OPTIONS",
					Value: "-Dcolor=green -Dcute=false",
				},
				{
					Name:  "JAVA_OPTS_APPEND",
					Value: "-Danimal=frog -Dcute=true",
				},
				{
					Name:  "VAR_3_NAME",
					Value: "VAR_3_VALUE",
				},
			}
			return s.k8sClient.Update(s.ctx, registry)
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Succeed())
		Eventually(func() []core.EnvVar {
			deployment := &apps.Deployment{}
			if err := s.k8sClient.Get(s.ctx, deploymentKey, deployment); err == nil {
				return deployment.Spec.Template.Spec.Containers[0].Env
			} else {
				return []core.EnvVar{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(
			ContainElements([]core.EnvVar{
				{
					Name:  "JAVA_OPTIONS",
					Value: "-Dcolor=green -Dcute=false",
				},
				{
					Name:  "JAVA_OPTS_APPEND",
					Value: "-Danimal=frog -Dcolor=green -Dcute=true",
				},
				{
					Name:  "VAR_3_NAME",
					Value: "VAR_3_VALUE",
				},
			}),
		)
	})
})
