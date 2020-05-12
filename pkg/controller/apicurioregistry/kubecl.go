package apicurioregistry

import (
	"encoding/json"
	"errors"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"os"
)

type KubeCl struct {
	client kubernetes.Interface
	ctx    *Context
}

// This is provides a way to update resources or otherwise talk to Kubernetes
func NewKubeCl(ctx *Context) *KubeCl {
	return &KubeCl{client: newKubeClOrDie(), ctx: ctx}
}

func newKubeClOrDie() kubernetes.Interface {
	cfg, err := inClusterConfig()
	if err != nil {
		panic(err)
	}
	return kubernetes.NewForConfigOrDie(cfg)
}

func inClusterConfig() (*rest.Config, error) {
	// Credit to etcd-operator
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		addrs, err := net.LookupHost("kubernetes.default.svc")
		if err != nil {
			panic(err)
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func createPatch(old, new, datastruct interface{}) ([]byte, error) {
	// Credit to etcd-operator
	o, err := json.Marshal(old)
	if err != nil {
		return nil, err
	}
	n, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}
	return strategicpatch.CreateTwoWayMergePatch(o, n, datastruct)
}

type DeploymentUF = func(spec *apps.Deployment)
type ServiceUF = func(spec *core.Service)
type IngressUF = func(spec *extensions.Ingress)

func (this *KubeCl) PatchDeployment(namespace, name string, updateFunc func(*apps.Deployment)) error {
	o, err := this.client.AppsV1().Deployments(namespace).Get(name, meta.GetOptions{})
	if err != nil {
		return err
	}
	n := o.DeepCopy()
	updateFunc(n)
	patchData, err := createPatch(o, n, apps.Deployment{})
	if err != nil {
		return err
	}
	_, err = this.client.AppsV1beta1().Deployments(namespace).Patch(name, types.StrategicMergePatchType, patchData)
	return err
}

func (this *KubeCl) PatchService(namespace, name string, updateFunc ServiceUF) error {
	o, err := this.client.CoreV1().Services(namespace).Get(name, meta.GetOptions{})
	if err != nil {
		return err
	}
	n := o.DeepCopy()
	updateFunc(n)
	patchData, err := createPatch(o, n, core.Service{})
	if err != nil {
		return err
	}
	_, err = this.client.CoreV1().Services(namespace).Patch(name, types.StrategicMergePatchType, patchData)
	return err
}

func (this *KubeCl) PatchIngress(namespace, name string, updateFunc func(*extensions.Ingress)) error {
	o, err := this.client.ExtensionsV1beta1().Ingresses(namespace).Get(name, meta.GetOptions{})
	if err != nil {
		return err
	}
	n := o.DeepCopy()
	updateFunc(n)
	patchData, err := createPatch(o, n, extensions.Ingress{})
	if err != nil {
		return err
	}
	_, err = this.client.ExtensionsV1beta1().Ingresses(namespace).Patch(name, types.StrategicMergePatchType, patchData)
	return err
}

// vvv === This should be on a higher level of abstraction

func (this *KubeCl) GetDeployment() (*apps.Deployment, error) {
	// TODO cache?
	if name := this.ctx.configuration.GetConfig(CFG_STA_DEPLOYMENT_NAME); name != "" {
		deployment, err := this.client.AppsV1().Deployments(this.ctx.configuration.GetSpecNamespace()).Get(name, meta.GetOptions{})
		return deployment, err
	}
	return nil, errors.New("No deployment name in status yet.")
}

func (this *KubeCl) GetService() (*core.Service, error) {
	// TODO cache?
	if name := this.ctx.configuration.GetConfig(CFG_STA_SERVICE_NAME); name != "" {
		service, err := this.client.CoreV1().Services(this.ctx.configuration.GetSpecNamespace()).Get(name, meta.GetOptions{})
		return service, err
	}
	return nil, errors.New("No service name in status yet.")
}

func (this *KubeCl) GetIngress() (*extensions.Ingress, error) {
	// TODO cache?
	if name := this.ctx.configuration.GetConfig(CFG_STA_INGRESS_NAME); name != "" {
		ingress, err := this.client.ExtensionsV1beta1().Ingresses(this.ctx.configuration.GetSpecNamespace()).Get(name, meta.GetOptions{})
		return ingress, err
	}
	return nil, errors.New("No ingress name in status yet.")
}

func (this *KubeCl) Client() kubernetes.Interface {
	return this.client
}
