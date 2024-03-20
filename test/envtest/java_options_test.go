package envtest

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	envv "github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
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
		Eventually(func() bool {
			deployment := &apps.Deployment{}
			var env []core.EnvVar = nil
			if err := s.k8sClient.Get(s.ctx, deploymentKey, deployment); err == nil {
				env = deployment.Spec.Template.Spec.Containers[0].Env
			} else {
				env = []core.EnvVar{}
			}
			// =============
			success := 0 // F****** Go
			if v, e := get(env, "JAVA_OPTIONS"); e {
				if v.Value == "-Dcolor=green -Dcute=false" {
					success++
				}
			}
			if v, e := get(env, "JAVA_OPTS_APPEND"); e {
				if parsed, err := envv.ParseShellArgs(v.Value); err == nil {
					if reflect.DeepEqual(parsed, map[string]string{
						"-Danimal": "frog",
						"-Dcute":   "true",
						"-Dcolor":  "green",
					}) {
						success++
					}
				}
			}
			if v, e := get(env, "VAR_3_NAME"); e {
				if v.Value == "VAR_3_VALUE" {
					success++
				}
			}
			return success == 3
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(BeTrue())
	})
})
