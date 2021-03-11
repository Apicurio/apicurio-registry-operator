package client

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// =====

type KubeClient struct {
	ctx    *context.LoopContext
	client kubernetes.Interface
}

func NewKubeClient(ctx *context.LoopContext, config *rest.Config) *KubeClient {
	return &KubeClient{
		client: kubernetes.NewForConfigOrDie(config),
		ctx:    ctx,
	}
}

// ===
// Deployment

func (this *KubeClient) CreateDeployment(namespace common.Namespace, value *apps.Deployment) (*apps.Deployment, error) {
	if err := controllerutil.SetControllerReference(getSpec(this.ctx), value, this.ctx.GetScheme()); err != nil {
		return nil, err
	}
	res, err := this.client.AppsV1().Deployments(namespace.Str()).Create(value)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetDeployment(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*apps.Deployment, error) {
	return this.client.AppsV1().Deployments(namespace.Str()).
		Get(name.Str(), *options)
}

func (this *KubeClient) UpdateDeployment(namespace common.Namespace, value *apps.Deployment) (*apps.Deployment, error) {
	return this.client.AppsV1().Deployments(namespace.Str()).
		Update(value)
}

func (this *KubeClient) PatchDeployment(namespace common.Namespace, name common.Name, patchData []byte) (*apps.Deployment, error) {
	return this.client.AppsV1().Deployments(namespace.Str()).
		Patch(name.Str(), types.StrategicMergePatchType, patchData)
}

func (this *KubeClient) GetDeployments(namespace common.Namespace, options *meta.ListOptions) (*apps.DeploymentList, error) {
	return this.client.AppsV1().Deployments(namespace.Str()).
		List(*options)
}

func (this *KubeClient) DeleteDeployment(value *apps.Deployment, options *meta.DeleteOptions) error {
	return this.client.AppsV1().Deployments(value.Namespace).Delete(value.Name, options)
}

// ===
// Service

func (this *KubeClient) CreateService(namespace common.Namespace, value *core.Service) (*core.Service, error) {
	if err := controllerutil.SetControllerReference(getSpec(this.ctx), value, this.ctx.GetScheme()); err != nil {
		return nil, err
	}
	res, err := this.client.CoreV1().Services(namespace.Str()).Create(value)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetService(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*core.Service, error) {
	return this.client.CoreV1().Services(namespace.Str()).
		Get(name.Str(), meta.GetOptions{})
}

func (this *KubeClient) UpdateService(namespace common.Namespace, value *core.Service) (*core.Service, error) {
	return this.client.CoreV1().Services(namespace.Str()).
		Update(value)
}

func (this *KubeClient) PatchService(namespace common.Namespace, name common.Name, patchData []byte) (*core.Service, error) {
	return this.client.CoreV1().Services(namespace.Str()).
		Patch(name.Str(), types.StrategicMergePatchType, patchData)
}

func (this *KubeClient) GetServices(namespace common.Namespace, options *meta.ListOptions) (*core.ServiceList, error) {
	return this.client.CoreV1().Services(namespace.Str()).
		List(*options)
}

func (this *KubeClient) DeleteService(value *core.Service, options *meta.DeleteOptions) error {
	return this.client.CoreV1().Services(value.Namespace).Delete(value.Name, options)
}

// ===
// Ingress

func (this *KubeClient) CreateIngress(namespace common.Namespace, value *extensions.Ingress) (*extensions.Ingress, error) {
	if err := controllerutil.SetControllerReference(getSpec(this.ctx), value, this.ctx.GetScheme()); err != nil {
		return nil, err
	}
	res, err := this.client.ExtensionsV1beta1().Ingresses(namespace.Str()).
		Create(value)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetIngress(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*extensions.Ingress, error) {
	return this.client.ExtensionsV1beta1().Ingresses(namespace.Str()).
		Get(name.Str(), meta.GetOptions{})
}

func (this *KubeClient) UpdateIngress(namespace common.Namespace, value *extensions.Ingress) (*extensions.Ingress, error) {
	return this.client.ExtensionsV1beta1().Ingresses(namespace.Str()).
		Update(value)
}

func (this *KubeClient) PatchIngress(namespace common.Namespace, name common.Name, patchData []byte) (*extensions.Ingress, error) {
	return this.client.ExtensionsV1beta1().Ingresses(namespace.Str()).
		Patch(name.Str(), types.StrategicMergePatchType, patchData)
}

func (this *KubeClient) GetIngresses(namespace common.Namespace, options *meta.ListOptions) (*extensions.IngressList, error) {
	return this.client.ExtensionsV1beta1().Ingresses(namespace.Str()).
		List(*options)
}

func (this *KubeClient) DeleteIngress(value *extensions.Ingress, options *meta.DeleteOptions) error {
	return this.client.ExtensionsV1beta1().Ingresses(value.Namespace).Delete(value.Name, options)
}

// ===
// PodDisruptionBudget

func (this *KubeClient) CreatePodDisruptionBudget(namespace common.Namespace, value *policy.PodDisruptionBudget) (*policy.PodDisruptionBudget, error) {
	if err := controllerutil.SetControllerReference(getSpec(this.ctx), value, this.ctx.GetScheme()); err != nil {
		return nil, err
	}
	res, err := this.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Str()).Create(value)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetPodDisruptionBudget(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*policy.PodDisruptionBudget, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Str()).
		Get(name.Str(), meta.GetOptions{})
}

func (this *KubeClient) UpdatePodDisruptionBudget(namespace common.Namespace, value *policy.PodDisruptionBudget) (*policy.PodDisruptionBudget, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Str()).
		Update(value)
}

func (this *KubeClient) PatchPodDisruptionBudget(namespace common.Namespace, name common.Name, patchData []byte) (*policy.PodDisruptionBudget, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Str()).
		Patch(name.Str(), types.StrategicMergePatchType, patchData)
}

func (this *KubeClient) GetPodDisruptionBudgets(namespace common.Namespace, options *meta.ListOptions) (*policy.PodDisruptionBudgetList, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Str()).
		List(*options)
}

func (this *KubeClient) DeletePodDisruptionBudget(value *policy.PodDisruptionBudget, options *meta.DeleteOptions) error {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(value.Namespace).Delete(value.Name, options)
}

// ===
// Pod

func (this *KubeClient) GetPod(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*core.Pod, error) {
	return this.client.CoreV1().Pods(namespace.Str()).
		Get(name.Str(), *options)
}
