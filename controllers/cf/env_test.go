package cf

import (
	v1 "github.com/Apicurio/apicurio-registry-operator/api/v1"
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	loop_impl "github.com/Apicurio/apicurio-registry-operator/controllers/loop/impl"
	services2 "github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestEnvCF(t *testing.T) {
	ctx := context.NewLoopContextMock()
	services := services2.NewLoopServicesMock(ctx)
	loop := loop_impl.NewControlLoopImpl(ctx, services)
	loop.AddControlFunction(NewEnvCF(ctx))

	ctx.GetResourceCache().Set(resources.RC_KEY_SPEC, resources.NewResourceCacheEntry(ctx.GetAppName(), &v1.ApicurioRegistry{
		Spec: v1.ApicurioRegistrySpec{
			Configuration: v1.ApicurioRegistrySpecConfiguration{
				Env: []corev1.EnvVar{
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
				},
			},
		},
	}))
	loop.Run()
	c.AssertEquals(t, []corev1.EnvVar{
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
	}, ctx.GetEnvCache().GetSorted())
	// Reordering
	ctx.GetResourceCache().Set(resources.RC_KEY_SPEC, resources.NewResourceCacheEntry(ctx.GetAppName(), &v1.ApicurioRegistry{
		Spec: v1.ApicurioRegistrySpec{
			Configuration: v1.ApicurioRegistrySpecConfiguration{
				Env: []corev1.EnvVar{
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
				},
			},
		},
	}))
	loop.Run()
	c.AssertEquals(t, []corev1.EnvVar{
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
	}, ctx.GetEnvCache().GetSorted())
	// Removing
	ctx.GetResourceCache().Set(resources.RC_KEY_SPEC, resources.NewResourceCacheEntry(ctx.GetAppName(), &v1.ApicurioRegistry{
		Spec: v1.ApicurioRegistrySpec{
			Configuration: v1.ApicurioRegistrySpecConfiguration{
				Env: []corev1.EnvVar{
					{
						Name:  "VAR_3_NAME",
						Value: "VAR_3_VALUE",
					},
					{
						Name:  "VAR_1_NAME",
						Value: "VAR_1_VALUE",
					},
				},
			},
		},
	}))
	loop.Run()
	c.AssertEquals(t, []corev1.EnvVar{
		{
			Name:  "VAR_3_NAME",
			Value: "VAR_3_VALUE",
		},
		{
			Name:  "VAR_1_NAME",
			Value: "VAR_1_VALUE",
		},
	}, ctx.GetEnvCache().GetSorted())
}

func TestEnvOrdering(t *testing.T) {
	ctx := context.NewLoopContextMock()
	services := services2.NewLoopServicesMock(ctx)
	loop := loop_impl.NewControlLoopImpl(ctx, services)
	loop.AddControlFunction(NewEnvCF(ctx))
	loop.AddControlFunction(NewEnvApplyCF(ctx))

	ctx.GetResourceCache().Set(resources.RC_KEY_DEPLOYMENT, resources.NewResourceCacheEntry(ctx.GetAppName(), &apps.Deployment{
		Spec: apps.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: ctx.GetAppName().Str(),
							Env: []corev1.EnvVar{
								{
									Name:  "DEPLOYMENT_VAR_1_NAME",
									Value: "DEPLOYMENT_VAR_1_VALUE",
								},
								{
									Name:  "DEPLOYMENT_VAR_2_NAME",
									Value: "DEPLOYMENT_VAR_2_VALUE",
								},
							},
						},
					},
				},
			},
		},
	}))
	ctx.GetResourceCache().Set(resources.RC_KEY_SPEC, resources.NewResourceCacheEntry(ctx.GetAppName(), &v1.ApicurioRegistry{
		Spec: v1.ApicurioRegistrySpec{
			Configuration: v1.ApicurioRegistrySpecConfiguration{
				Env: []corev1.EnvVar{
					{
						Name:  "SPEC_VAR_1_NAME",
						Value: "SPEC_VAR_1_VALUE",
					},
					{
						Name:  "SPEC_VAR_2_NAME",
						Value: "SPEC_VAR_2_VALUE",
					},
				},
			},
		},
	}))
	loop.Run()
	ctx.GetResourceCache().Set(resources.RC_KEY_SPEC, resources.NewResourceCacheEntry(ctx.GetAppName(), &v1.ApicurioRegistry{
		Spec: v1.ApicurioRegistrySpec{
			Configuration: v1.ApicurioRegistrySpecConfiguration{
				Env: []corev1.EnvVar{
					{
						Name:  "SPEC_VAR_2_NAME",
						Value: "SPEC_VAR_2_VALUE",
					},
					{
						Name:  "SPEC_VAR_3_NAME",
						Value: "SPEC_VAR_3_VALUE",
					},
				},
			},
		},
	}))
	loop.Run()
	sorted := ctx.GetEnvCache().GetSorted()
	sortedI := convert(sorted)
	c.AssertIsInOrder(t, sortedI,
		corev1.EnvVar{
			Name:  "SPEC_VAR_2_NAME",
			Value: "SPEC_VAR_2_VALUE",
		},
		corev1.EnvVar{
			Name:  "SPEC_VAR_3_NAME",
			Value: "SPEC_VAR_3_VALUE",
		})
	c.AssertIsInOrder(t, sortedI,
		corev1.EnvVar{
			Name:  "DEPLOYMENT_VAR_1_NAME",
			Value: "DEPLOYMENT_VAR_1_VALUE",
		},
		corev1.EnvVar{
			Name:  "DEPLOYMENT_VAR_2_NAME",
			Value: "DEPLOYMENT_VAR_2_VALUE",
		})
}

func TestEnvPriority(t *testing.T) {
	ctx := context.NewLoopContextMock()
	services := services2.NewLoopServicesMock(ctx)
	loop := loop_impl.NewControlLoopImpl(ctx, services)
	loop.AddControlFunction(NewEnvCF(ctx))
	loop.AddControlFunction(NewEnvApplyCF(ctx))

	// In reverse priority
	// Spec Sourced
	ctx.GetResourceCache().Set(resources.RC_KEY_SPEC, resources.NewResourceCacheEntry(ctx.GetAppName(), &v1.ApicurioRegistry{
		Spec: v1.ApicurioRegistrySpec{
			Configuration: v1.ApicurioRegistrySpecConfiguration{
				Env: []corev1.EnvVar{
					{
						Name:  "VAR_3_NAME",
						Value: "SPEC",
					},
				},
			},
		},
	}))
	loop.Run()

	// TODO When overwriting an entry, previous dependencies are removed!
	// Operator Managed
	ctx.GetEnvCache().Set(env.NewSimpleEnvCacheEntryBuilder("VAR_2_NAME", "OPERATOR").Build())
	ctx.GetEnvCache().Set(env.NewSimpleEnvCacheEntryBuilder("VAR_3_NAME", "OPERATOR").Build())
	loop.Run()

	// Deployment - User Managed
	ctx.GetResourceCache().Set(resources.RC_KEY_DEPLOYMENT, resources.NewResourceCacheEntry(ctx.GetAppName(), &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name: "test",
		},
		Spec: apps.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: ctx.GetAppName().Str(),
							Env: []corev1.EnvVar{
								{
									Name:  "VAR_1_NAME",
									Value: "DEPLOYMENT",
								},
								{
									Name:  "VAR_2_NAME",
									Value: "DEPLOYMENT",
								},
								{
									Name:  "VAR_3_NAME",
									Value: "DEPLOYMENT",
								},
							},
						},
					},
				},
			},
		},
	}))
	loop.Run()

	sorted := ctx.GetEnvCache().GetSorted()
	sortedI := convert(sorted)
	c.AssertSliceContains(t, sortedI, corev1.EnvVar{
		Name:  "VAR_1_NAME",
		Value: "DEPLOYMENT",
	})
	c.AssertSliceContains(t, sortedI, corev1.EnvVar{
		Name:  "VAR_2_NAME",
		Value: "OPERATOR",
	})
	c.AssertSliceContains(t, sortedI, corev1.EnvVar{
		Name:  "VAR_3_NAME",
		Value: "OPERATOR",
	})
	c.AssertIsInOrder(t, sortedI,
		corev1.EnvVar{
			Name:  "VAR_1_NAME",
			Value: "DEPLOYMENT",
		},
		corev1.EnvVar{
			Name:  "VAR_2_NAME",
			Value: "OPERATOR",
		})
	c.AssertIsInOrder(t, sortedI,
		corev1.EnvVar{
			Name:  "VAR_1_NAME",
			Value: "DEPLOYMENT",
		},
		corev1.EnvVar{
			Name:  "VAR_3_NAME",
			Value: "OPERATOR",
		})
}

func convert(data []corev1.EnvVar) []interface{} {
	res := make([]interface{}, len(data))
	for i, v := range data {
		res[i] = v
	}
	return res
}
