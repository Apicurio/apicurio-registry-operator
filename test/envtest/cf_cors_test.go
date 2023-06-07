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

var _ = Describe("cf_cors", Ordered, func() {

	var registryKey types.NamespacedName
	var deploymentKey types.NamespacedName

	const testNamespace = "cf-cors-test-namespace"
	const registryName = "test"

	BeforeAll(func() {
		// Consistency in case the specs are reordered
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
			Spec: ar.ApicurioRegistrySpec{},
		}
		Expect(s.k8sClient.Create(s.ctx, registry)).To(Succeed())
		registryKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name}
	})

	It("should create the CORS_ALLOWED_ORIGINS env. variable", func() {
		deployment := &apps.Deployment{}
		deploymentKey = types.NamespacedName{Namespace: registryKey.Namespace, Name: registryKey.Name + "-deployment"}
		Eventually(func() []core.EnvVar {
			if err := s.k8sClient.Get(s.ctx, deploymentKey, deployment); err == nil {
				return deployment.Spec.Template.Spec.Containers[0].Env
			} else {
				return []core.EnvVar{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(ContainElements([]core.EnvVar{
			{
				Name: "CORS_ALLOWED_ORIGINS",
				Value: "http://" + registryKey.Name + "." + registryKey.Namespace + "," +
					"https://" + registryKey.Name + "." + registryKey.Namespace,
			},
		}))
	})

	It("should update the CORS_ALLOWED_ORIGINS env. variable", func() {
		registry := &ar.ApicurioRegistry{}
		Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
		registry.Spec.Deployment.Host = "baz"
		Expect(s.k8sClient.Update(s.ctx, registry)).To(Succeed())
		deployment := &apps.Deployment{}
		Eventually(func() []core.EnvVar {
			if err := s.k8sClient.Get(s.ctx, deploymentKey, deployment); err == nil {
				return deployment.Spec.Template.Spec.Containers[0].Env
			} else {
				return []core.EnvVar{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Not(ContainElements([]core.EnvVar{
			{
				Name:  "CORS_ALLOWED_ORIGINS",
				Value: "baz",
			},
		})))
	})

	It("should remove the CORS_ALLOWED_ORIGINS env. variable", func() {
		registry := &ar.ApicurioRegistry{}
		Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
		registry.Spec.Deployment.Host = ""
		Expect(s.k8sClient.Update(s.ctx, registry)).To(Succeed())
		deployment := &apps.Deployment{}
		Eventually(func() []core.EnvVar {
			if err := s.k8sClient.Get(s.ctx, deploymentKey, deployment); err == nil {
				return deployment.Spec.Template.Spec.Containers[0].Env
			} else {
				return []core.EnvVar{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Not(ContainElements([]core.EnvVar{
			{
				Name: "CORS_ALLOWED_ORIGINS",
				Value: "http://" + registryKey.Name + "." + registryKey.Namespace + "," +
					"https://" + registryKey.Name + "." + registryKey.Namespace,
			},
		})))
	})

	It("should not prevent custom CORS_ALLOWED_ORIGINS env. variable", func() {
		registry := &ar.ApicurioRegistry{}
		Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
		registry.Spec.Configuration.Env = []core.EnvVar{
			{
				Name:  "CORS_ALLOWED_ORIGINS",
				Value: "foo",
			},
		}
		Expect(s.k8sClient.Update(s.ctx, registry)).To(Succeed())
		deployment := &apps.Deployment{}
		Eventually(func() []core.EnvVar {
			if err := s.k8sClient.Get(s.ctx, deploymentKey, deployment); err == nil {
				return deployment.Spec.Template.Spec.Containers[0].Env
			} else {
				return []core.EnvVar{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(ContainElements([]core.EnvVar{
			{
				Name:  "CORS_ALLOWED_ORIGINS",
				Value: "foo",
			},
		}))
	})
})
