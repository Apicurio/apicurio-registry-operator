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

var _ = Describe("dynamic environment variables", Ordered, func() {

	var registryKey types.NamespacedName
	var deploymentKey types.NamespacedName

	const testNamespace = "dynamic-env-test-namespace"
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
			Spec: ar.ApicurioRegistrySpec{},
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
					Name:  "VAR_1_NAME",
					Value: "VAR_1_VALUE",
				},
				{
					Name:  "VAR_2_NAME",
					Value: "VAR_2_VALUE",
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
			And(
				ContainElements([]core.EnvVar{
					{
						Name:  "VAR_1_NAME",
						Value: "VAR_1_VALUE",
					},
					{
						Name:  "VAR_2_NAME",
						Value: "VAR_2_VALUE",
					},
					{
						Name:  "VAR_3_NAME",
						Value: "VAR_3_VALUE",
					},
				}),
				Satisfy(func(vars []core.EnvVar) bool {
					return isBefore(vars, "VAR_1_NAME", "VAR_2_NAME") &&
						isBefore(vars, "VAR_2_NAME", "VAR_3_NAME")
				}),
			),
		)
	})

	It("should be reordered", func() {
		Eventually(func() error {
			registry := &ar.ApicurioRegistry{}
			Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
			registry.Spec.Configuration.Env = []core.EnvVar{
				{
					Name:  "VAR_3_NAME",
					Value: "VAR_3_VALUE",
				},
				{
					Name:  "VAR_2_NAME",
					Value: "VAR_2_VALUE",
				},
				{
					Name:  "VAR_1_NAME",
					Value: "VAR_1_VALUE",
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
			And(
				ContainElements([]core.EnvVar{
					{
						Name:  "VAR_1_NAME",
						Value: "VAR_1_VALUE",
					},
					{
						Name:  "VAR_2_NAME",
						Value: "VAR_2_VALUE",
					},
					{
						Name:  "VAR_3_NAME",
						Value: "VAR_3_VALUE",
					},
				}),
				Satisfy(func(vars []core.EnvVar) bool {
					return isBefore(vars, "VAR_2_NAME", "VAR_1_NAME") &&
						isBefore(vars, "VAR_3_NAME", "VAR_2_NAME")
				}),
			),
		)
	})

	It("should not override environment variables set by the operator", func() {
		Eventually(func() error {
			registry := &ar.ApicurioRegistry{}
			Expect(s.k8sClient.Get(s.ctx, registryKey, registry)).To(Succeed())
			registry.Spec.Configuration.LogLevel = "DEBUG"
			registry.Spec.Configuration.RegistryLogLevel = "DEBUG"
			registry.Spec.Configuration.Env = []core.EnvVar{
				{
					Name:  "LOG_LEVEL",
					Value: "OVERRIDDEN_FROM_SPEC",
				},
				{
					Name:  "REGISTRY_LOG_LEVEL",
					Value: "OVERRIDDEN_FROM_SPEC",
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
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(ContainElements([]core.EnvVar{
			{
				Name:  "LOG_LEVEL",
				Value: "DEBUG",
			},
			{
				Name:  "REGISTRY_LOG_LEVEL",
				Value: "DEBUG",
			},
		}))
	})

	It("should work with variables set in the deployment", func() {
		Eventually(func() error {
			deployment := &apps.Deployment{}
			Expect(s.k8sClient.Get(s.ctx, deploymentKey, deployment)).To(Succeed())
			deployment.Spec.Template.Spec.Containers[0].Env = []core.EnvVar{
				{
					Name:  "VAR_4_NAME",
					Value: "VAR_4_VALUE",
				},
			}
			return s.k8sClient.Update(s.ctx, deployment)
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Succeed())
		Eventually(func() []core.EnvVar {
			deployment := &apps.Deployment{}
			if err := s.k8sClient.Get(s.ctx, deploymentKey, deployment); err == nil {
				return deployment.Spec.Template.Spec.Containers[0].Env
			} else {
				return []core.EnvVar{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(ContainElements([]core.EnvVar{
			{
				Name:  "VAR_4_NAME",
				Value: "VAR_4_VALUE",
			},
		}))
		Eventually(func() error {
			deployment := &apps.Deployment{}
			Expect(s.k8sClient.Get(s.ctx, deploymentKey, deployment)).To(Succeed())
			deployment.Spec.Template.Spec.Containers[0].Env = []core.EnvVar{}
			return s.k8sClient.Update(s.ctx, deployment)
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Succeed())
		Eventually(func() []core.EnvVar {
			deployment := &apps.Deployment{}
			if err := s.k8sClient.Get(s.ctx, deploymentKey, deployment); err == nil {
				return deployment.Spec.Template.Spec.Containers[0].Env
			} else {
				return []core.EnvVar{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Not(ContainElements([]core.EnvVar{
			{
				Name:  "VAR_4_NAME",
				Value: "VAR_4_VALUE",
			},
		})))
	})

})
