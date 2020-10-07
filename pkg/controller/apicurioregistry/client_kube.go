package apicurioregistry

import (
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
	ctx    *Context
	client kubernetes.Interface
}

func NewKubeClient(ctx *Context, config *rest.Config) *KubeClient {
	return &KubeClient{
		client: kubernetes.NewForConfigOrDie(config),
		ctx:    ctx,
	}
}

// ===
// Deployment

func (this *KubeClient) CreateDeployment(namespace string, value *apps.Deployment) (*apps.Deployment, error) {
	res, err := this.client.AppsV1().Deployments(namespace).
		Create(value)
	if err != nil {
		return nil, err
	}
	if err := controllerutil.SetControllerReference(this.ctx.GetConfiguration().GetSpec(), res, this.ctx.GetScheme()); err != nil {
		panic("Could not set controller reference.")
	}
	res, err = this.UpdateDeployment(namespace, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetDeployment(namespace string, name string, options *meta.GetOptions) (*apps.Deployment, error) {
	return this.client.AppsV1().Deployments(namespace).
		Get(name, *options)
}

func (this *KubeClient) UpdateDeployment(namespace string, value *apps.Deployment) (*apps.Deployment, error) {
	return this.client.AppsV1().Deployments(namespace).
		Update(value)
}

func (this *KubeClient) PatchDeployment(namespace, name string, patchData []byte) (*apps.Deployment, error) {
	return this.client.AppsV1().Deployments(namespace).
		Patch(name, types.StrategicMergePatchType, patchData)
}

func (this *KubeClient) GetDeployments(namespace string, options *meta.ListOptions) (*apps.DeploymentList, error) {
	return this.client.AppsV1().Deployments(namespace).
		List(*options)
}

func (this *KubeClient) DeleteDeployment(value *apps.Deployment, options *meta.DeleteOptions) error {
	return this.client.AppsV1().Deployments(value.Namespace).Delete(value.Name, options)
}

// ===
// Service

func (this *KubeClient) CreateService(namespace string, value *core.Service) (*core.Service, error) {
	res, err := this.client.CoreV1().Services(namespace).
		Create(value)
	if err != nil {
		return nil, err
	}
	if err := controllerutil.SetControllerReference(this.ctx.GetConfiguration().GetSpec(), res, this.ctx.GetScheme()); err != nil {
		panic("Could not set controller reference.")
	}
	res, err = this.UpdateService(namespace, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetService(namespace string, name string, options *meta.GetOptions) (*core.Service, error) {
	return this.client.CoreV1().Services(namespace).
		Get(name, meta.GetOptions{})
}

func (this *KubeClient) UpdateService(namespace string, value *core.Service) (*core.Service, error) {
	return this.client.CoreV1().Services(namespace).
		Update(value)
}

func (this *KubeClient) PatchService(namespace, name string, patchData []byte) (*core.Service, error) {
	return this.client.CoreV1().Services(namespace).
		Patch(name, types.StrategicMergePatchType, patchData)
}

func (this *KubeClient) GetServices(namespace string, options *meta.ListOptions) (*core.ServiceList, error) {
	return this.client.CoreV1().Services(namespace).
		List(*options)
}

func (this *KubeClient) DeleteService(value *core.Service, options *meta.DeleteOptions) error {
	return this.client.CoreV1().Services(value.Namespace).Delete(value.Name, options)
}

// ===
// Ingress

func (this *KubeClient) CreateIngress(namespace string, value *extensions.Ingress) (*extensions.Ingress, error) {
	res, err := this.client.ExtensionsV1beta1().Ingresses(namespace).
		Create(value)
	if err != nil {
		return nil, err
	}
	if err := controllerutil.SetControllerReference(this.ctx.GetConfiguration().GetSpec(), res, this.ctx.GetScheme()); err != nil {
		panic("Could not set controller reference.")
	}
	res, err = this.UpdateIngress(namespace, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetIngress(namespace string, name string, options *meta.GetOptions) (*extensions.Ingress, error) {
	return this.client.ExtensionsV1beta1().Ingresses(namespace).
		Get(name, meta.GetOptions{})
}

func (this *KubeClient) UpdateIngress(namespace string, value *extensions.Ingress) (*extensions.Ingress, error) {
	return this.client.ExtensionsV1beta1().Ingresses(namespace).
		Update(value)
}

func (this *KubeClient) PatchIngress(namespace, name string, patchData []byte) (*extensions.Ingress, error) {
	return this.client.ExtensionsV1beta1().Ingresses(namespace).
		Patch(name, types.StrategicMergePatchType, patchData)
}

func (this *KubeClient) GetIngresses(namespace string, options *meta.ListOptions) (*extensions.IngressList, error) {
	return this.client.ExtensionsV1beta1().Ingresses(namespace).
		List(*options)
}

func (this *KubeClient) DeleteIngress(value *extensions.Ingress, options *meta.DeleteOptions) error {
	return this.client.ExtensionsV1beta1().Ingresses(value.Namespace).Delete(value.Name, options)
}

// ===
// PodDisruptionBudget

func (this *KubeClient) CreatePodDisruptionBudget(namespace string, value *policy.PodDisruptionBudget) (*policy.PodDisruptionBudget, error) {
	res, err := this.client.PolicyV1beta1().PodDisruptionBudgets(namespace).
		Create(value)
	if err != nil {
		return nil, err
	}
	if err := controllerutil.SetControllerReference(this.ctx.GetConfiguration().GetSpec(), res, this.ctx.GetScheme()); err != nil {
		panic("Could not set controller reference.")
	}
	res, err = this.UpdatePodDisruptionBudget(namespace, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetPodDisruptionBudget(namespace string, name string, options *meta.GetOptions) (*policy.PodDisruptionBudget, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace).
		Get(name, meta.GetOptions{})
}

func (this *KubeClient) UpdatePodDisruptionBudget(namespace string, value *policy.PodDisruptionBudget) (*policy.PodDisruptionBudget, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace).
		Update(value)
}

func (this *KubeClient) PatchPodDisruptionBudget(namespace, name string, patchData []byte) (*policy.PodDisruptionBudget, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace).
		Patch(name, types.StrategicMergePatchType, patchData)
}

func (this *KubeClient) GetPodDisruptionBudgets(namespace string, options *meta.ListOptions) (*policy.PodDisruptionBudgetList, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace).
		List(*options)
}

func (this *KubeClient) DeletePodDisruptionBudget(value *policy.PodDisruptionBudget, options *meta.DeleteOptions) error {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(value.Namespace).Delete(value.Name, options)
}

// ===
// Pod

func (this *KubeClient) GetPod(namespace string, name string, options *meta.GetOptions) (*core.Pod, error) {
	return this.client.CoreV1().Pods(namespace).
		Get(name, *options)
}
