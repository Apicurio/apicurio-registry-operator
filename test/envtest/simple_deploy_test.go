package envtest

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gt "github.com/onsi/gomega/types"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	policy_v1 "k8s.io/api/policy/v1"
	policy_v1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("operator processing a simple spec", Ordered, func() {

	var registry *ar.ApicurioRegistry
	var registryKey types.NamespacedName
	var deploymentKey types.NamespacedName
	var serviceKey types.NamespacedName
	var ingressKey types.NamespacedName
	var pdbKey types.NamespacedName
	var npKey types.NamespacedName

	BeforeAll(func() {
		// Consistency in case the specs are reordered
		testSupport.SetMockCanMakeHTTPRequestToOperand(false)
		testSupport.SetMockOperandMetricsReportReady(false)
		ns := &core.Namespace{
			ObjectMeta: meta.ObjectMeta{
				Name: "test-namespace",
			},
		}
		Expect(s.k8sClient.Create(context.TODO(), ns)).To(Succeed())
		registry = &ar.ApicurioRegistry{
			ObjectMeta: meta.ObjectMeta{
				Name:      "test",
				Namespace: ns.ObjectMeta.Name,
			},
			Spec: ar.ApicurioRegistrySpec{},
		}
		Expect(s.k8sClient.Create(s.ctx, registry)).To(Succeed())
		registryKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name}
	})

	It("should create a deployment", func() {
		deployment := &apps.Deployment{}
		deploymentKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name + "-deployment"}
		Eventually(func() error {
			return s.k8sClient.Get(s.ctx, deploymentKey, deployment)
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Succeed())
	})

	It("should create a PDB", func() {
		if testSupport.GetSupportedFeatures().PreferredPDBVersion == "v1beta1" {
			pdb := &policy_v1beta1.PodDisruptionBudget{}
			pdbKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name + "-pdb"}
			Eventually(func() error {
				return s.k8sClient.Get(s.ctx, pdbKey, pdb)
			}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Succeed())
		} else if testSupport.GetSupportedFeatures().PreferredPDBVersion == "v1" {
			pdb := &policy_v1.PodDisruptionBudget{}
			pdbKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name + "-pdb"}
			Eventually(func() error {
				return s.k8sClient.Get(s.ctx, pdbKey, pdb)
			}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Succeed())
		} else {
			Fail("Unexpected preferred PDB version: " + testSupport.GetSupportedFeatures().PreferredPDBVersion)
		}
	})

	It("should create a service", func() {
		service := &core.Service{}
		serviceKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name + "-service"}
		Eventually(func() error {
			return s.k8sClient.Get(s.ctx, serviceKey, service)
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Succeed())
	})

	It("should create an ingress", func() {
		ingress := &networking.Ingress{}
		ingressKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name + "-ingress"}
		Eventually(func() error {
			return s.k8sClient.Get(s.ctx, ingressKey, ingress)
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Succeed())
	})

	It("should create an network policy", func() {
		np := &networking.NetworkPolicy{}
		npKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name + "-networkpolicy"}
		Eventually(func() error {
			return s.k8sClient.Get(s.ctx, npKey, np)
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(Succeed())
	})

	It("should report managed resources", func() {
		registry2 := &ar.ApicurioRegistry{}
		Eventually(func() []ar.ApicurioRegistryStatusManagedResource {
			if err := s.k8sClient.Get(s.ctx, registryKey, registry2); err == nil {
				return registry2.Status.ManagedResources
			} else {
				return []ar.ApicurioRegistryStatusManagedResource{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(ConsistOf([]ar.ApicurioRegistryStatusManagedResource{
			{
				Kind:      "Deployment",
				Name:      "test-deployment",
				Namespace: registry.Namespace,
			},
			{
				Kind:      "Service",
				Name:      "test-service",
				Namespace: registry.Namespace,
			},
			{
				Kind:      "Ingress",
				Name:      "test-ingress",
				Namespace: registry.Namespace,
			},
		}))
	})

	It("should report initializing condition", func() {
		registry2 := &ar.ApicurioRegistry{}
		Eventually(func() []meta.Condition {
			if err := s.k8sClient.Get(s.ctx, registryKey, registry2); err == nil {
				return registry2.Status.Conditions
			} else {
				return []meta.Condition{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Type":    Equal("Ready"),
			"Status":  Equal(meta.ConditionFalse),
			"Reason":  Equal("Initializing"),
			"Message": Equal(""),
		})))
	})

	It("should report operand not healthy condition", func() {
		testSupport.SetMockCanMakeHTTPRequestToOperand(true)
		registry2 := &ar.ApicurioRegistry{}
		Eventually(func() []meta.Condition {
			if err := s.k8sClient.Get(s.ctx, registryKey, registry2); err == nil {
				return registry2.Status.Conditions
			} else {
				return []meta.Condition{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(ConsistOf([]gt.GomegaMatcher{
			MatchFields(IgnoreExtras, Fields{
				"Type":    Equal("Ready"),
				"Status":  Equal(meta.ConditionFalse),
				"Reason":  Equal("Error"),
				"Message": Equal("An error occurred in the operator or the application. Please check other conditions and logs."),
			}),
			MatchFields(IgnoreExtras, Fields{
				"Type":    Equal("ApplicationNotHealthy"),
				"Status":  Equal(meta.ConditionTrue),
				"Reason":  Equal("ReadinessProbeFailed"),
				"Message": Equal("Readiness probe is failing. Please check application logs."),
			}),
		}))
	})

	It("should report operand is healthy condition", func() {
		testSupport.SetMockOperandMetricsReportReady(true)
		registry2 := &ar.ApicurioRegistry{}
		Eventually(func() []meta.Condition {
			if err := s.k8sClient.Get(s.ctx, registryKey, registry2); err == nil {
				return registry2.Status.Conditions
			} else {
				return []meta.Condition{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Type":    Equal("Ready"),
			"Status":  Equal(meta.ConditionTrue),
			"Reason":  Equal("Reconciled"),
			"Message": Equal(""),
		})))
	})

	It("should get to a stable control state", func() {
		Eventually(func() bool {
			return testSupport.TimerDuration() > 10*time.Second
		}, 20*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(BeTrue())
	})

	It("should delete created resources during cleanup", func() {
		Expect(s.k8sClient.Delete(s.ctx, registry)).To(Succeed())
		Eventually(func() bool {
			pdb := false
			if testSupport.GetSupportedFeatures().PreferredPDBVersion == "v1beta1" {
				pdb = errors.IsNotFound(s.k8sClient.Get(s.ctx, pdbKey, &policy_v1beta1.PodDisruptionBudget{}))
			}
			if testSupport.GetSupportedFeatures().PreferredPDBVersion == "v1" {
				pdb = errors.IsNotFound(s.k8sClient.Get(s.ctx, pdbKey, &policy_v1.PodDisruptionBudget{}))
			}
			return errors.IsNotFound(s.k8sClient.Get(s.ctx, deploymentKey, &apps.Deployment{})) &&
				errors.IsNotFound(s.k8sClient.Get(s.ctx, serviceKey, &core.Service{})) &&
				errors.IsNotFound(s.k8sClient.Get(s.ctx, ingressKey, &networking.Ingress{})) &&
				pdb &&
				errors.IsNotFound(s.k8sClient.Get(s.ctx, npKey, &networking.NetworkPolicy{}))
		}, 20*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(BeTrue())
	})
})
