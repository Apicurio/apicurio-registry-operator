package client

import (
	ctx "context"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// =====

type CRDClient struct {
	ctx *context.LoopContext
	//ctx.client should be used instead of this rest client
	client *rest.RESTClient
	codec  runtime.ParameterCodec
}

func NewCRDClient(ctx *context.LoopContext, config *rest.Config) *CRDClient {

	ctx.GetScheme().AddKnownTypes(ar.GroupVersion, &ar.ApicurioRegistry{}, &ar.ApicurioRegistryList{})
	meta.AddToGroupVersion(ctx.GetScheme(), ar.GroupVersion)

	config.ContentConfig.GroupVersion = &ar.GroupVersion
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
		codec:  runtime.NewParameterCodec(ctx.GetScheme()),
	}
}

// ===
// ApicurioRegistry

func (this *CRDClient) GetApicurioRegistry(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*ar.ApicurioRegistry, error) {
	result := ar.ApicurioRegistry{}
	err := this.client.
		Get().
		Namespace(namespace.Str()).
		Resource("apicurioregistries"). // TODO
		Name(name.Str()).
		VersionedParams(options, this.codec).
		Do(ctx.TODO()).
		Into(&result)

	return &result, err
}

func (this *CRDClient) UpdateApicurioRegistry(namespace common.Namespace, value *ar.ApicurioRegistry) (*ar.ApicurioRegistry, error) {
	result := ar.ApicurioRegistry{}
	err := this.client.
		Put().
		Namespace(namespace.Str()).
		Resource("apicurioregistries").
		Name(value.Name).
		Body(value).
		Do(ctx.TODO()).
		Into(&result)

	return &result, err
}

func (this *CRDClient) PatchApicurioRegistry(namespace common.Namespace, name common.Name, patchData []byte) (*ar.ApicurioRegistry, error) {
	err := this.client.
		Patch(types.MergePatchType).
		Resource("apicurioregistries").
		Body(patchData).
		Namespace(namespace.Str()).
		Name(name.Str()).
		Do(ctx.TODO()).
		Error()
	if err != nil {
		return nil, err
	}

	result := &ar.ApicurioRegistry{}
	err = this.client.
		Patch(types.MergePatchType).
		Resource("apicurioregistries").
		Body(patchData).
		Namespace(namespace.Str()).
		Name(name.Str()).
		Do(ctx.TODO()).
		Into(result)

	return result, err
}
