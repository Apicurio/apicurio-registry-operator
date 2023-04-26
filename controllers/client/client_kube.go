package client

import (
	ctx "context"
	"errors"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	policy_v1 "k8s.io/api/policy/v1"
	policy_v1beta1 "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// =====

type KubeClient struct {
	log    *zap.Logger
	client kubernetes.Interface
	scheme *runtime.Scheme
}

func NewKubeClient(log *zap.Logger, scheme *runtime.Scheme, config *rest.Config) *KubeClient {
	return &KubeClient{
		client: kubernetes.NewForConfigOrDie(config),
		log:    log,
		scheme: scheme,
	}
}

// ===
// Deployment

func (this *KubeClient) CreateDeployment(owner meta.Object, namespace common.Namespace, value *apps.Deployment) (*apps.Deployment, error) {
	if owner == nil {
		return nil, errors.New("Could not find ApicurioRegistry. Retrying.")
	}
	if err := controllerutil.SetControllerReference(owner, value, this.scheme); err != nil {
		return nil, err
	}
	res, err := this.client.AppsV1().Deployments(namespace.Str()).Create(ctx.TODO(), value, meta.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetDeployment(namespace common.Namespace, name common.Name) (*apps.Deployment, error) {
	return this.client.AppsV1().Deployments(namespace.Str()).
		Get(ctx.TODO(), name.Str(), meta.GetOptions{})
}

func (this *KubeClient) UpdateDeployment(namespace common.Namespace, value *apps.Deployment) (*apps.Deployment, error) {
	return this.client.AppsV1().Deployments(namespace.Str()).
		Update(ctx.TODO(), value, meta.UpdateOptions{})
}

func (this *KubeClient) PatchDeployment(namespace common.Namespace, name common.Name, patchData []byte) (*apps.Deployment, error) {
	return this.client.AppsV1().Deployments(namespace.Str()).
		Patch(ctx.TODO(), name.Str(), types.MergePatchType, patchData, meta.PatchOptions{})
}

func (this *KubeClient) GetDeployments(namespace common.Namespace, options meta.ListOptions) (*apps.DeploymentList, error) {
	return this.client.AppsV1().Deployments(namespace.Str()).
		List(ctx.TODO(), options)
}

func (this *KubeClient) DeleteDeployment(value *apps.Deployment) error {
	return this.client.AppsV1().Deployments(value.Namespace).Delete(ctx.TODO(), value.Name, meta.DeleteOptions{})
}

// ===
// Service

func (this *KubeClient) CreateService(owner meta.Object, namespace common.Namespace, value *core.Service) (*core.Service, error) {
	if owner == nil {
		return nil, errors.New("Could not find ApicurioRegistry. Retrying.")
	}
	if err := controllerutil.SetControllerReference(owner, value, this.scheme); err != nil {
		return nil, err
	}
	res, err := this.client.CoreV1().Services(namespace.Str()).Create(ctx.TODO(), value, meta.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetService(namespace common.Namespace, name common.Name) (*core.Service, error) {
	return this.client.CoreV1().Services(namespace.Str()).
		Get(ctx.TODO(), name.Str(), meta.GetOptions{})
}

func (this *KubeClient) UpdateService(namespace common.Namespace, value *core.Service) (*core.Service, error) {
	return this.client.CoreV1().Services(namespace.Str()).
		Update(ctx.TODO(), value, meta.UpdateOptions{})
}

func (this *KubeClient) PatchService(namespace common.Namespace, name common.Name, patchData []byte) (*core.Service, error) {
	return this.client.CoreV1().Services(namespace.Str()).
		Patch(ctx.TODO(), name.Str(), types.MergePatchType, patchData, meta.PatchOptions{})
}

func (this *KubeClient) GetServices(namespace common.Namespace, options meta.ListOptions) (*core.ServiceList, error) {
	return this.client.CoreV1().Services(namespace.Str()).
		List(ctx.TODO(), options)
}

func (this *KubeClient) DeleteService(value *core.Service) error {
	return this.client.CoreV1().Services(value.Namespace).Delete(ctx.TODO(), value.Name, meta.DeleteOptions{})
}

// ===
// Ingress

func (this *KubeClient) CreateIngress(owner meta.Object, namespace common.Namespace, value *networking.Ingress) (*networking.Ingress, error) {
	if owner == nil {
		return nil, errors.New("Could not find ApicurioRegistry. Retrying.")
	}
	if err := controllerutil.SetControllerReference(owner, value, this.scheme); err != nil {
		return nil, err
	}
	res, err := this.client.NetworkingV1().Ingresses(namespace.Str()).
		Create(ctx.TODO(), value, meta.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetIngress(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*networking.Ingress, error) {
	return this.client.NetworkingV1().Ingresses(namespace.Str()).
		Get(ctx.TODO(), name.Str(), meta.GetOptions{})
}

func (this *KubeClient) UpdateIngress(namespace common.Namespace, value *networking.Ingress) (*networking.Ingress, error) {
	return this.client.NetworkingV1().Ingresses(namespace.Str()).
		Update(ctx.TODO(), value, meta.UpdateOptions{})
}

func (this *KubeClient) PatchIngress(namespace common.Namespace, name common.Name, patchData []byte) (*networking.Ingress, error) {
	return this.client.NetworkingV1().Ingresses(namespace.Str()).
		Patch(ctx.TODO(), name.Str(), types.MergePatchType, patchData, meta.PatchOptions{})
}

func (this *KubeClient) GetIngresses(namespace common.Namespace, options meta.ListOptions) (*networking.IngressList, error) {
	return this.client.NetworkingV1().Ingresses(namespace.Str()).
		List(ctx.TODO(), options)
}

func (this *KubeClient) DeleteIngress(value *networking.Ingress) error {
	return this.client.NetworkingV1().Ingresses(value.Namespace).Delete(ctx.TODO(), value.Name, meta.DeleteOptions{})
}

// ===
// Network Policy

func (this *KubeClient) CreateNetworkPolicy(owner meta.Object, namespace common.Namespace, value *networking.NetworkPolicy) (*networking.NetworkPolicy, error) {
	if owner == nil {
		return nil, errors.New("Could not find ApicurioRegistry. Retrying.")
	}
	if err := controllerutil.SetControllerReference(owner, value, this.scheme); err != nil {
		return nil, err
	}
	res, err := this.client.NetworkingV1().NetworkPolicies(namespace.Str()).
		Create(ctx.TODO(), value, meta.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetNetworkPolicy(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*networking.NetworkPolicy, error) {
	return this.client.NetworkingV1().NetworkPolicies(namespace.Str()).
		Get(ctx.TODO(), name.Str(), meta.GetOptions{})
}

func (this *KubeClient) UpdateNetworkPolicy(namespace common.Namespace, value *networking.NetworkPolicy) (*networking.NetworkPolicy, error) {
	return this.client.NetworkingV1().NetworkPolicies(namespace.Str()).
		Update(ctx.TODO(), value, meta.UpdateOptions{})
}

func (this *KubeClient) PatchNetworkPolicy(namespace common.Namespace, name common.Name, patchData []byte) (*networking.NetworkPolicy, error) {
	return this.client.NetworkingV1().NetworkPolicies(namespace.Str()).
		Patch(ctx.TODO(), name.Str(), types.MergePatchType, patchData, meta.PatchOptions{})
}

func (this *KubeClient) GetNetworkPolicies(namespace common.Namespace, options meta.ListOptions) (*networking.NetworkPolicyList, error) {
	return this.client.NetworkingV1().NetworkPolicies(namespace.Str()).
		List(ctx.TODO(), options)
}

func (this *KubeClient) DeleteNetworkPolicy(value *networking.NetworkPolicy) error {
	return this.client.NetworkingV1().NetworkPolicies(value.Namespace).Delete(ctx.TODO(), value.Name, meta.DeleteOptions{})
}

// ===
// PodDisruptionBudget v1beta1

func (this *KubeClient) CreatePodDisruptionBudgetV1beta1(owner meta.Object, namespace common.Namespace, value *policy_v1beta1.PodDisruptionBudget) (*policy_v1beta1.PodDisruptionBudget, error) {
	if owner == nil {
		return nil, errors.New("Could not find ApicurioRegistry. Retrying.")
	}
	if err := controllerutil.SetControllerReference(owner, value, this.scheme); err != nil {
		return nil, err
	}
	res, err := this.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Str()).Create(ctx.TODO(), value, meta.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetPodDisruptionBudgetV1beta1(namespace common.Namespace, name common.Name) (*policy_v1beta1.PodDisruptionBudget, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Str()).
		Get(ctx.TODO(), name.Str(), meta.GetOptions{})
}

func (this *KubeClient) UpdatePodDisruptionBudgetV1beta1(namespace common.Namespace, value *policy_v1beta1.PodDisruptionBudget) (*policy_v1beta1.PodDisruptionBudget, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Str()).
		Update(ctx.TODO(), value, meta.UpdateOptions{})
}

func (this *KubeClient) PatchPodDisruptionBudgetV1beta1(namespace common.Namespace, name common.Name, patchData []byte) (*policy_v1beta1.PodDisruptionBudget, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Str()).
		Patch(ctx.TODO(), name.Str(), types.MergePatchType, patchData, meta.PatchOptions{})
}

func (this *KubeClient) GetPodDisruptionBudgetsV1beta1(namespace common.Namespace, options meta.ListOptions) (*policy_v1beta1.PodDisruptionBudgetList, error) {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Str()).
		List(ctx.TODO(), options)
}

func (this *KubeClient) DeletePodDisruptionBudgetV1beta1(value *policy_v1beta1.PodDisruptionBudget) error {
	return this.client.PolicyV1beta1().PodDisruptionBudgets(value.Namespace).Delete(ctx.TODO(), value.Name, meta.DeleteOptions{})
}

// ===
// PodDisruptionBudget v1

func (this *KubeClient) CreatePodDisruptionBudgetV1(owner meta.Object, namespace common.Namespace, value *policy_v1.PodDisruptionBudget) (*policy_v1.PodDisruptionBudget, error) {
	if owner == nil {
		return nil, errors.New("Could not find ApicurioRegistry. Retrying.")
	}
	if err := controllerutil.SetControllerReference(owner, value, this.scheme); err != nil {
		return nil, err
	}
	res, err := this.client.PolicyV1().PodDisruptionBudgets(namespace.Str()).Create(ctx.TODO(), value, meta.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetPodDisruptionBudgetV1(namespace common.Namespace, name common.Name) (*policy_v1.PodDisruptionBudget, error) {
	return this.client.PolicyV1().PodDisruptionBudgets(namespace.Str()).
		Get(ctx.TODO(), name.Str(), meta.GetOptions{})
}

func (this *KubeClient) UpdatePodDisruptionBudgetV1(namespace common.Namespace, value *policy_v1.PodDisruptionBudget) (*policy_v1.PodDisruptionBudget, error) {
	return this.client.PolicyV1().PodDisruptionBudgets(namespace.Str()).
		Update(ctx.TODO(), value, meta.UpdateOptions{})
}

func (this *KubeClient) PatchPodDisruptionBudgetV1(namespace common.Namespace, name common.Name, patchData []byte) (*policy_v1.PodDisruptionBudget, error) {
	return this.client.PolicyV1().PodDisruptionBudgets(namespace.Str()).
		Patch(ctx.TODO(), name.Str(), types.MergePatchType, patchData, meta.PatchOptions{})
}

func (this *KubeClient) GetPodDisruptionBudgetsV1(namespace common.Namespace, options meta.ListOptions) (*policy_v1.PodDisruptionBudgetList, error) {
	return this.client.PolicyV1().PodDisruptionBudgets(namespace.Str()).
		List(ctx.TODO(), options)
}

func (this *KubeClient) DeletePodDisruptionBudgetV1(value *policy_v1.PodDisruptionBudget) error {
	return this.client.PolicyV1().PodDisruptionBudgets(value.Namespace).Delete(ctx.TODO(), value.Name, meta.DeleteOptions{})
}

// ===
// Pod

func (this *KubeClient) GetPod(namespace common.Namespace, name common.Name) (*core.Pod, error) {
	return this.client.CoreV1().Pods(namespace.Str()).
		Get(ctx.TODO(), name.Str(), meta.GetOptions{})
}

func (this *KubeClient) CreateSecret(owner meta.Object, namespace common.Namespace, value *core.Secret) (*core.Secret, error) {
	if owner == nil {
		return nil, errors.New("Could not find ApicurioRegistry. Retrying.")
	}
	if err := controllerutil.SetControllerReference(owner, value, this.scheme); err != nil {
		return nil, err
	}
	res, err := this.client.CoreV1().Secrets(namespace.Str()).Create(ctx.TODO(), value, meta.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *KubeClient) GetSecret(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*core.Secret, error) {
	return this.client.CoreV1().Secrets(namespace.Str()).
		Get(ctx.TODO(), name.Str(), *options)
}

func (this *KubeClient) DeleteSecret(value *core.Secret, options *meta.DeleteOptions) error {
	return this.client.CoreV1().Secrets(value.Namespace).
		Delete(ctx.TODO(), value.Name, *options)
}
