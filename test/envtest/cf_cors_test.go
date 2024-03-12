package envtest

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
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
			Spec: ar.ApicurioRegistrySpec{
				Configuration: ar.ApicurioRegistrySpecConfiguration{
					Persistence: "sql",
				},
			},
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

	It("should update the CORS_ALLOWED_ORIGINS env. variable, with host", func() {
		registry := &ar.ApicurioRegistry{}
		Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
		registry.Spec.Deployment.Host = "host"
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
				Value: "http://host,https://host",
			},
		}))
	})

	It("should update the CORS_ALLOWED_ORIGINS env. variable, with Keycloak only", func() {
		registry := &ar.ApicurioRegistry{}
		Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
		registry.Spec.Deployment.Host = ""
		registry.Spec.Configuration.Security.Keycloak.Url = "httporhttps://keycloak.cloud/path"
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
				Value: "httporhttps://keycloak.cloud",
			},
		}))
	})

	It("should update the CORS_ALLOWED_ORIGINS env. variable, with host and Keycloak", func() {
		registry := &ar.ApicurioRegistry{}
		Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
		registry.Spec.Deployment.Host = "host"
		registry.Spec.Configuration.Security.Keycloak.Url = "httporhttps://keycloak.cloud/path"
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
				Value: "http://host,https://host,httporhttps://keycloak.cloud",
			},
		}))
	})

	It("should not prevent custom CORS_ALLOWED_ORIGINS env. variable", func() {
		registry := &ar.ApicurioRegistry{}
		Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
		registry.Spec.Deployment.Host = "host"
		registry.Spec.Configuration.Security.Keycloak.Url = "httporhttps://keycloak.cloud/path"
		registry.Spec.Configuration.Env = []core.EnvVar{
			{
				Name:  "CORS_ALLOWED_ORIGINS",
				Value: "custom",
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
				Value: "custom",
			},
		}))
	})

	It("should remove the CORS_ALLOWED_ORIGINS env. variable", func() {
		registry := &ar.ApicurioRegistry{}
		Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
		registry.Spec.Deployment.Host = ""
		registry.Spec.Configuration.Security.Keycloak.Url = ""
		registry.Spec.Configuration.Env = []core.EnvVar{}
		Expect(s.k8sClient.Update(s.ctx, registry)).To(Succeed())
		deployment := &apps.Deployment{}
		Eventually(func() []core.EnvVar {
			if err := s.k8sClient.Get(s.ctx, deploymentKey, deployment); err == nil {
				return deployment.Spec.Template.Spec.Containers[0].Env
			} else {
				return []core.EnvVar{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Not(ContainElements(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("CORS_ALLOWED_ORIGINS"),
		}))))
	})
})
