package envtest

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"time"
)

var _ = Describe("operator configuring https", Ordered, func() {

	var registry *ar.ApicurioRegistry
	//var registryKey types.NamespacedName
	var secretKey types.NamespacedName
	var serviceKey types.NamespacedName

	BeforeAll(func() {
		const registryName = "test"
		// Consistency in case the specs are reordered
		testSupport.SetMockCanMakeHTTPRequestToOperand(false)
		testSupport.SetMockOperandMetricsReportReady(false)
		ns := &core.Namespace{
			ObjectMeta: meta.ObjectMeta{
				Name: "https-test-namespace",
			},
		}
		Expect(s.k8sClient.Create(context.TODO(), ns)).To(Succeed())
		secretKey = types.NamespacedName{Namespace: ns.ObjectMeta.Name, Name: registryName + "-https-secret"}
		secret := &core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name:      secretKey.Name,
				Namespace: secretKey.Namespace,
			},
			Data: map[string][]byte{
				"tls.crt": []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURaVENDQWsyZ0F3SUJBZ0lVQktQYTZzcTRTN0NnWHY0QWxnNDVpS205cTVBd0RRWUpLb1pJaHZjTkFRRUwKQlFBd1FqRUxNQWtHQTFVRUJoTUNXRmd4RlRBVEJnTlZCQWNNREVSbFptRjFiSFFnUTJsMGVURWNNQm9HQTFVRQpDZ3dUUkdWbVlYVnNkQ0JEYjIxd1lXNTVJRXgwWkRBZUZ3MHlNekExTVRneE16UTRORFphRncwek16QTFNVFV4Ck16UTRORFphTUVJeEN6QUpCZ05WQkFZVEFsaFlNUlV3RXdZRFZRUUhEQXhFWldaaGRXeDBJRU5wZEhreEhEQWEKQmdOVkJBb01FMFJsWm1GMWJIUWdRMjl0Y0dGdWVTQk1kR1F3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQgpEd0F3Z2dFS0FvSUJBUUQwMEZKalZtTFNVMjR3NTQ0SXk0cFBYZXBUMk10RDFzbURvWFdLU25nRnNJaVpMcXllCjNka01MeG5mbGIwVWFzYUZJWklvVXQ2ZlRFQXAvdVRCZk5MTHl4MHB3aGFQam9BK2hrMmlYT0I2bWV0WHIzbnMKa3JnVzZQTWdDUmpCOTJNZ2ZDc0g3anZlbnhHWjFqR0cvNmNuUFo2SGRyb0MyU0N5SkZKZnVTY3FqUWdGLzhzaQpMNDFhQ1Bxa25ic3Jia1RubGFVYlhXMjhJN3ZYQlpyV1NYZHIrWW82SVg1ZkphTDlIdFl6SlVNWXJ5RlkremtNCnVkaW5HNndkcENhTjcrYzdmZENTalZmWHBEWnpPOFNTS1crdVJFMjRTaHIvNGpjelRIUXh6amllOE9rd3FrQmsKR2hWejZIaDAxd1FsdmNiODNCT2Q3TWZNNFA4cTRaTEUvanAxQWdNQkFBR2pVekJSTUIwR0ExVWREZ1FXQkJRbgpnMUdKWlVSandyVEwzMEVDWTRNUlRkUUtYakFmQmdOVkhTTUVHREFXZ0JRbmcxR0paVVJqd3JUTDMwRUNZNE1SClRkUUtYakFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUNaTFdRWU9OSFYKbURxb1B6Zk8rOFNBVEF3cHdndk1iT2FFbnZqL2xybHdwbFdqbmpSUmFkWTAxU01Hb3B2dTI1cGNFYzYwTTJjdgo4c2NXZTFCZWE3bXYrelJEbjM0MHUzWklNRGFYMjE4QlVsdDhtSEhuNjVTWjZWcHBKVElGNmRPWm5VK0Erc01xCkk3SUpaYUQ1QlRhOGRjc0N5STFWeUZ4bEFibFM5MllFbEowanBuSmsrVFdidGdseFduMk9EQUxYMzBzL3UzS3EKSkQ3cnY5YTlZVjRkOVZsOGpKNzljeCt0MGMyUTNBK2N4YkNYZzJYMXV0MjZmRVlMdDJSU2x5SGdSOUtRckhydQpHNjREMGV3dnpGYk96aXpSUlJ0dmVOaWNINkFyUUthMWJITDZ4QW5YckhZNUFLUENGcG41V1kzemVPb2tORmpECjg5Yjh4dzNnMEFHbgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="),
				"tls.key": []byte("LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRRDAwRkpqVm1MU1UyNHcKNTQ0SXk0cFBYZXBUMk10RDFzbURvWFdLU25nRnNJaVpMcXllM2RrTUx4bmZsYjBVYXNhRklaSW9VdDZmVEVBcAovdVRCZk5MTHl4MHB3aGFQam9BK2hrMmlYT0I2bWV0WHIzbnNrcmdXNlBNZ0NSakI5Mk1nZkNzSDdqdmVueEdaCjFqR0cvNmNuUFo2SGRyb0MyU0N5SkZKZnVTY3FqUWdGLzhzaUw0MWFDUHFrbmJzcmJrVG5sYVViWFcyOEk3dlgKQlpyV1NYZHIrWW82SVg1ZkphTDlIdFl6SlVNWXJ5RlkremtNdWRpbkc2d2RwQ2FONytjN2ZkQ1NqVmZYcERaegpPOFNTS1crdVJFMjRTaHIvNGpjelRIUXh6amllOE9rd3FrQmtHaFZ6NkhoMDF3UWx2Y2I4M0JPZDdNZk00UDhxCjRaTEUvanAxQWdNQkFBRUNnZ0VBTmVKNGorYmV2MzZmbldJS01FTmt3UTFoMjJ5M2FNb285ckVlSnY4M0pjRnkKZjR6M2I4eFN6c3k3UEN4QVB2TTFtTzRIdHBwdTU4OG52RmFmVVR0QlJwd0JZa1NYSktmdjhGTXRXVlJxRUhJNgppOFZTNTlCdmRwTjFtQktJZ1lFTEw0WkZEbXpRZnJLeWRCTGlPZDJobEJDTENUUUh3MEs1WUp5QUNSTyszQzFhClp4bmdxa2FXcEJSVFhQVTVsdFRXRHNWNmk3VFdpalFqTHkwSEJSMmVDTTV1Sm9oKytwOEJaczY1VmNZSlVQL3kKZkRNbk9BQ2h3RUZseHBNM0ZuZml6RVplZWh1TUcrdFRrY0Zhb1RMKzZhbU42RnhMUFU1aURmRFhPZlBDQzd1TgowZmEwRFdKbFZhdjVQSGRQbUxYZFY3NnB6QUQ2ZlJKLzFKcDh0OFhqb1FLQmdRRDZnMDJGczBMcHlacnNudzE0Cld0MDJWL2w5VnlmSS9USjVwNGswUXdtekl1QWlzdXdRdGQ3SlJZT0JRSkExZkt5VDNXenV2ZTYvZWNmcS9VYzUKRmthb1BRZkJLNXR5bjF0emlydlBYWFVYQ0phQmdIb3RuOWsvNDQxRzRsekxoYlNHaVBLdTBJdVE5Y3AxK3ZvRQo4OGZwbEhhRU45MjNTT1NWMSs2Y2RxTzIyUUtCZ1FENkxRL3hvU0JFTGh5L2dRM2h6d3p6VHdCaWM4OFk0WnVqCjZaZzVkYmw5a3ZpN3BrRWgzS01Da3ZVQk5XU2FNME0vMzE4RnpUNCt1MHplTnRTd3JBUGdCdkQyV2twM2R0NjcKZUlTUXNlcDlnNEg0YU5ObEh2czVOODdZTjZCOXZJenlZTzdRanc4THkxaTM0TEo5NCsvQ1RDQXRlaEJNVGVBdApmMFo5eks3Mi9RS0JnQ0VFdVcwTDZaL2k0TGFiYUMwYTNObFMweUdBSVZCT2Z4Nmx4R0hORERRK1BvaVVTS1VUCk02QVh0M09MelBZZnpxZFdvZ3I5b2NBL0R1aWNKWTBTc0pGd0tkdCtJZWtEdEF3UWx4eUgxdTBJUnI0ZTd2dWcKZkFQOXZCdEJyclZzbEJTL2JDMDZjNHJSdXJPK05zSDhWN2NqeUZNNUFkSXNtMlJjcDZpYndveFJBb0dCQU1iKwpTdjFXdlpTZDNTNFNtQmt5R1VuN1lBSHZ2aDQ3YmhKdVB6QU5UUkx1Y2J6SkhHdXoxVkc1MVBvMkh5UnNmQ1IxCkoxODFCenJjdnVMT1dGV0RMYjNucDRrOC9waVJ5ODd3cVBsekcyTGsxTi9qZWFxb2Z3bmZNejlXMStqTHJvMG8Kdnl6VGJoTmlsdG9EOTlZZEZWdkdNNTRZeHBmN0pjTHF4d1pQWmloOUFvR0JBTkFJZTJQSXdGVklCV3dFaVhHdgo2N2FPTXVpOTUxelF4MER0ZkI1QXRJRXl2SThGOSt3LzJqdWNhcnZBS2xmN2k1b0Y3ak1mZ2cxeEwzWlZmOTl1ClF4NlcxY04wQUJkYTdMWUhYVElpUlZvd0VEZ2J1T0FzNU1YSk05WHRsdWYxcFVkemlTWjVPYUtkQkpySFNUWHIKS05PMFBKUTJ5Rjc0bXFCYTNkeHMzKy9tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K"),
			},
		}
		Expect(s.k8sClient.Create(context.TODO(), secret)).To(Succeed())
		registry = &ar.ApicurioRegistry{
			ObjectMeta: meta.ObjectMeta{
				Name:      registryName,
				Namespace: ns.ObjectMeta.Name,
			},
			Spec: ar.ApicurioRegistrySpec{
				Configuration: ar.ApicurioRegistrySpecConfiguration{
					Security: ar.ApicurioRegistrySpecConfigurationSecurity{
						Https: ar.ApicurioRegistrySpecConfigurationSecurityHttps{
							SecretName: secretKey.Name,
						},
					},
				},
			},
		}
		Expect(s.k8sClient.Create(s.ctx, registry)).To(Succeed())
		//registryKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name}
	})

	// TODO Add tests for other resources. This specifically tests issue #149.
	It("should create a service", func() {
		service := &core.Service{}
		serviceKey = types.NamespacedName{Namespace: registry.Namespace, Name: registry.Name + "-service"}
		Eventually(func() []core.ServicePort {
			if err := s.k8sClient.Get(s.ctx, serviceKey, service); err == nil {
				return service.Spec.Ports
			} else {
				return []core.ServicePort{}
			}
		}, 10*time.Second*T_SCALE, EVENTUALLY_CHECK_PERIOD).Should(ConsistOf(
			core.ServicePort{
				Name:       "https",
				Protocol:   core.ProtocolTCP,
				Port:       8443,
				TargetPort: intstr.FromInt(8443),
			},
			core.ServicePort{
				Name:       "http",
				Protocol:   core.ProtocolTCP,
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
			},
		))
	})
})
