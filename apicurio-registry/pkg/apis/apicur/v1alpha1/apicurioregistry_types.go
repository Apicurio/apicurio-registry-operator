package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApicurioRegistryList contains a list of ApicurioRegistry
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ApicurioRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApicurioRegistry `json:"items"`
}

// ApicurioRegistry is the Schema for the apicurioregistries API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ApicurioRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApicurioRegistrySpec   `json:"spec,omitempty"`
	Status ApicurioRegistryStatus `json:"status,omitempty"`
}

// ApicurioRegistrySpec defines the desired state of ApicurioRegistry
// +k8s:openapi-gen=true
type ApicurioRegistrySpec struct {
	Image         ApicurioRegistrySpecImage         `json:"image,omitempty"`
	Configuration ApicurioRegistrySpecConfiguration `json:"configuration,omitempty"`
	Deployment    ApicurioRegistrySpecDeployment    `json:"deployment,omitempty"`
}

// +k8s:openapi-gen=true
type ApicurioRegistrySpecImage struct {
	Registry string `json:"registry,omitempty"`
	Version  string `json:"version,omitempty"`
	Override string `json:"override,omitempty"`
}

// +k8s:openapi-gen=true
type ApicurioRegistrySpecConfiguration struct {
	// +kubebuilder:validation:Enum=mem,jpa,kafka,streams
	Persistence string                                      `json:"persistence,omitempty"`
	DataSource  ApicurioRegistrySpecConfigurationDataSource `json:"dataSource,omitempty"`
	Kafka       ApicurioRegistrySpecConfigurationKafka      `json:"kafka,omitempty"`
	Streams     ApicurioRegistrySpecConfigurationStreams    `json:"streams,omitempty"`
}

// +k8s:openapi-gen=true
type ApicurioRegistrySpecConfigurationDataSource struct {
	Url      string `json:"url,omitempty"`
	UserName string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
}

// +k8s:openapi-gen=true
type ApicurioRegistrySpecConfigurationKafka struct {
	BootstrapServers string `json:"bootstrapServers,omitempty"`
}

// +k8s:openapi-gen=true
type ApicurioRegistrySpecConfigurationStreams struct {
	BootstrapServers string `json:"bootstrapServers,omitempty"`
	ApplicationServer    string `json:"applicationServer,omitempty"`
	ApplicationId    string `json:"applicationId,omitempty"`
}

// +k8s:openapi-gen=true
type ApicurioRegistrySpecDeployment struct {
	Replicas  int32                                  `json:"replicas,omitempty"`
	Route     string                                  `json:"route,omitempty"`
	Resources ApicurioRegistrySpecDeploymentResources `json:"resources,omitempty"`
}

// +k8s:openapi-gen=true
type ApicurioRegistrySpecDeploymentResources struct {
	Cpu    ApicurioRegistrySpecDeploymentResourcesRequestsLimit `json:"cpu,omitempty"`
	Memory ApicurioRegistrySpecDeploymentResourcesRequestsLimit `json:"memory,omitempty"`
}

// +k8s:openapi-gen=true
type ApicurioRegistrySpecDeploymentResourcesRequestsLimit struct {
	Requests string `json:"requests,omitempty"`
	Limit    string `json:"limit,omitempty"`
}

// ApicurioRegistryStatus defines the observed state of ApicurioRegistry
// +k8s:openapi-gen=true
type ApicurioRegistryStatus struct {
	//Replicas int32    `json:"replicas,omitempty"`
	Image string `json:"image,omitempty"`
	DeploymentName string `json:"deploymentName,omitempty"`
	ServiceName string `json:"serviceName,omitempty"`
	IngressName string `json:"ingressName,omitempty"`
	ReplicaCount int32 `json:"replicaCount,omitempty"`
	Route string `json:"route,omitempty"`
	//CpuRequests string `json:"cpuRequests,omitempty"`
	//CpuLimits string `json:"cpuLimits,omitempty"`
	//MemoryRequests string `json:"memoryRequests,omitempty"`
	//MemoryLimits string `json:"memoryLimits,omitempty"`
	//PersistenceConfigurationValid bool `json:"persistenceConfigurationValid,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ApicurioRegistry{}, &ApicurioRegistryList{})
}
