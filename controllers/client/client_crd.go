package client

import (
	ctx "context"
	"encoding/json"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// =====

type CRDClient struct {
	//ctx context.LoopContext
	log *zap.Logger
	//ctx.client should be used instead of this rest client
	client *rest.RESTClient
	codec  runtime.ParameterCodec
}

func NewCRDClient(log *zap.Logger, scheme *runtime.Scheme, config *rest.Config) *CRDClient {

	scheme.AddKnownTypes(ar.GroupVersion, &ar.ApicurioRegistry{}, &ar.ApicurioRegistryList{})
	meta.AddToGroupVersion(scheme, ar.GroupVersion)

	config2 := rest.CopyConfig(config)
	config2.ContentConfig.GroupVersion = &ar.GroupVersion
	config2.APIPath = "/apis"
	config2.NegotiatedSerializer = serializer.NewCodecFactory(scheme)
	config2.UserAgent = rest.DefaultKubernetesUserAgent()

	c, err := rest.UnversionedRESTClientFor(config2)
	if err != nil {
		log.Sugar().Error(err)
		panic("Could not create Kubernetes client for ApicurioRegistry CRD.")
	}
	return &CRDClient{
		client: c,
		log:    log,
		codec:  runtime.NewParameterCodec(scheme),
	}
}

// ===
// ApicurioRegistry

// TODO Returns nil if the resource is not found, but request was OK
func (this *CRDClient) GetApicurioRegistry(namespace common.Namespace, name common.Name) (*ar.ApicurioRegistry, error) {
	result := &ar.ApicurioRegistry{}
	err := this.client.
		Get().
		Resource("apicurioregistries").
		Namespace(namespace.Str()).
		Name(name.Str()).
		Do(ctx.TODO()).
		Into(result)
	if errors.IsNotFound(err) {
		return nil, nil
	}
	return result, err
}

func (this *CRDClient) UpdateApicurioRegistry(namespace common.Namespace, value *ar.ApicurioRegistry) (*ar.ApicurioRegistry, error) {
	result := &ar.ApicurioRegistry{}
	err := this.client.
		Put().
		Resource("apicurioregistries").
		Namespace(namespace.Str()).
		Name(value.Name).
		Body(value).
		Do(ctx.TODO()).
		Into(result)

	return result, err
}

func (this *CRDClient) PatchApicurioRegistry(namespace common.Namespace, name common.Name, patchData []byte) (*ar.ApicurioRegistry, error) {
	result := &ar.ApicurioRegistry{}
	err := this.client.
		Patch(types.MergePatchType).
		Resource("apicurioregistries").
		Namespace(namespace.Str()).
		Name(name.Str()).
		Body(patchData).
		Do(ctx.TODO()).
		Into(result)

	return result, err
}

func (this *CRDClient) PatchApicurioRegistryStatus(namespace common.Namespace, name common.Name, patchData []byte) (*ar.ApicurioRegistryStatus, error) {
	// Add "status" prefix to the patch path
	var original map[string]interface{}
	if err := json.Unmarshal(patchData, &original); err != nil {
		this.log.Sugar().Error(err)
		panic("Could not patch ApicurioRegistry status.")
	}
	extended := map[string]interface{}{
		"status": original,
	}
	extendedPatchData, err := json.Marshal(extended)
	if err != nil {
		this.log.Sugar().Error(err)
		panic("Could not patch ApicurioRegistry status.")
	}

	result := &ar.ApicurioRegistry{}
	err = this.client.
		Patch(types.MergePatchType).
		Resource("apicurioregistries").
		SubResource("status").
		Namespace(namespace.Str()).
		Name(name.Str()).
		Body(extendedPatchData).
		Do(ctx.TODO()).
		Into(result)

	return &result.Status, err
}
