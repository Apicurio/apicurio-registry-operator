package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// =====

type CRDClient struct {
	ctx    *Context
	client *rest.RESTClient
}

func NewCRDClient(ctx *Context, config *rest.Config) *CRDClient {

	ctx.scheme.AddKnownTypes(ar.SchemeGroupVersion, &ar.ApicurioRegistry{}, &ar.ApicurioRegistryList{})

	config.ContentConfig.GroupVersion = &ar.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.NewCodecFactory(ctx.GetScheme())
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	c, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		panic("Could not create CRD client.")
	}
	return &CRDClient{
		client: c,
		ctx:    ctx,
	}
}

// ===
// ApicurioRegistry

func (this *CRDClient) GetApicurioRegistry(namespace string, name string, options *meta.GetOptions) (*ar.ApicurioRegistry, error) {
	result := ar.ApicurioRegistry{}
	err := this.client.
		Get().
		Namespace(namespace).
		Resource(ar.GroupResource).
		Name(name).
		VersionedParams(&meta.GetOptions{}, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (this *CRDClient) UpdateApicurioRegistry(namespace string, value *ar.ApicurioRegistry) (*ar.ApicurioRegistry, error) {
	result := ar.ApicurioRegistry{}
	err := this.client.
		Put().
		Namespace(namespace).
		Resource(ar.GroupResource).
		Name(value.Name).
		Body(value).
		Do().
		Into(&result)

	return &result, err
}

func (this *CRDClient) PatchApicurioRegistry(namespace, name string, patchData []byte) (*ar.ApicurioRegistry, error) {
	err := this.client.
		Patch(types.MergePatchType).
		Resource(ar.GroupResource).
		SubResource("status").
		Body(patchData).
		Namespace(namespace).
		Name(name).
		Do().
		Error()
	if err != nil {
		return nil, err
	}

	result := &ar.ApicurioRegistry{}
	err = this.client.
		Patch(types.MergePatchType).
		Resource(ar.GroupResource).
		Body(patchData).
		Namespace(namespace).
		Name(name).
		Do().
		Into(result)

	return result, err
}
