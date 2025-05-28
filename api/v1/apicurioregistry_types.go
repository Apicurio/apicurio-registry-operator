/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ### Spec

// ApicurioRegistrySpec defines the desired state of ApicurioRegistry
type ApicurioRegistrySpec struct {
	// Apicurio Registry application configuration
	Configuration ApicurioRegistrySpecConfiguration `json:"configuration,omitempty"`
	// Apicurio Registry deployment configuration
	Deployment ApicurioRegistrySpecDeployment `json:"deployment,omitempty"`
}

type ApicurioRegistrySpecConfiguration struct {
	// Storage:
	//
	// Type of storage used by Apicurio Registry, one of: mem, sql, kafkasql.
	// Default value is `mem`.
	Persistence string `json:"persistence,omitempty"`
	// Configuration of Apicurio Registry SQL storage
	Sql ApicurioRegistrySpecConfigurationSql `json:"sql,omitempty"`
	// Configuration of Apicurio Registry KafkaSQL storage
	Kafkasql ApicurioRegistrySpecConfigurationKafkasql `json:"kafkasql,omitempty"`
	// Configuration of Apicurio Registry web console
	UI ApicurioRegistrySpecConfigurationUI `json:"ui,omitempty"`
	// Third-party (non-Apicurio) library log level
	LogLevel string `json:"logLevel,omitempty"`
	// Apicurio Registry application log level
	RegistryLogLevel string `json:"registryLogLevel,omitempty"`
	// Security configuration
	Security ApicurioRegistrySpecConfigurationSecurity `json:"security,omitempty"`
	// Environment variables:
	//
	// List of additional environment variables that will be
	// provided to the Apicurio Registry application.
	Env []core.EnvVar `json:"env,omitempty"`
}

type ApicurioRegistrySpecConfigurationDataSource struct {
	// Data source URL:
	//
	// URL of the PostgreSQL database, for example:
	// `jdbc:postgresql://<service name>.<namespace>.svc:5432/<database name>`.
	Url string `json:"url,omitempty"`
	// Data source username
	UserName string `json:"userName,omitempty"`
	// Data source password
	Password string `json:"password,omitempty"` // TODO Support Secrets
}

type ApicurioRegistrySpecConfigurationSql struct {
	// SQL data source
	DataSource ApicurioRegistrySpecConfigurationDataSource `json:"dataSource,omitempty"`
}

type ApicurioRegistrySpecConfigurationKafkasql struct {
	// Kafka bootstrap servers URL:
	//
	// URL of one of the Kafka brokers, which provide initial metadata about the Kafka cluster,
	// for example: `<service name>.<namespace>.svc:9092`.
	BootstrapServers string `json:"bootstrapServers,omitempty"`
	// Kafka security configuration:
	//
	// Provide the following configuration options if your Kafka cluster
	// is secured using TLS or SCRAM.
	Security ApicurioRegistrySpecConfigurationKafkaSecurity `json:"security,omitempty"`
}

type ApicurioRegistrySpecConfigurationKafkaSecurity struct {
	// TLS:
	//
	// Kafka is secured using TLS.
	Tls ApicurioRegistrySpecConfigurationKafkaSecurityTls `json:"tls,omitempty"`
	// SCRAM:
	//
	// Kafka is secured using SCRAM.
	Scram ApicurioRegistrySpecConfigurationKafkaSecurityScram `json:"scram,omitempty"`
}

type ApicurioRegistrySpecConfigurationKafkaSecurityTls struct {
	// Truststore Secret name:
	//
	// Name of a Secret that contains TLS truststore (in PKCS12 format)
	// under the `ca.p12` key, and truststore password under the `ca.password` key.
	TruststoreSecretName string `json:"truststoreSecretName,omitempty"`
	// Keystore Secret name:
	//
	// Name of a Secret that contains TLS keystore (in PKCS12 format)
	// under the `user.p12` key, and keystore password under the `user.password` key.
	KeystoreSecretName string `json:"keystoreSecretName,omitempty"`
}

type ApicurioRegistrySpecConfigurationKafkaSecurityScram struct {
	// Truststore Secret name:
	//
	// Name of a Secret that contains TLS truststore (in PKCS12 format)
	// under the `ca.p12` key, and truststore password under the `ca.password` key.
	TruststoreSecretName string `json:"truststoreSecretName,omitempty"`
	// User name
	User string `json:"user,omitempty"`
	// User password Secret name:
	//
	// Name of a Secret that contains password of the SCRAM user
	// under the `password` key.
	PasswordSecretName string `json:"passwordSecretName,omitempty"`
	// Mechanism:
	//
	// Name of the SCRAM mechanism, default value is SCRAM-SHA-512.
	Mechanism string `json:"mechanism,omitempty"`
}

type ApicurioRegistrySpecConfigurationUI struct {
	// Read-only:
	//
	// Set the web console to read-only mode.
	// WARNING: This does not affect access to the Apicurio REST API.
	ReadOnly bool `json:"readOnly,omitempty"`
}

type ApicurioRegistrySpecConfigurationSecurity struct {
	// Keycloak:
	//
	// Configure Apicurio Registry to use Keycloak for Identity and Access Management (IAM).
	Keycloak ApicurioRegistrySpecConfigurationSecurityKeycloak `json:"keycloak,omitempty"`
	// HTTPS:
	//
	// Configure Apicurio Registry to be accessible using HTTPS.
	Https ApicurioRegistrySpecConfigurationSecurityHttps `json:"https,omitempty"`
}

type ApicurioRegistrySpecConfigurationSecurityHttps struct {
	// Disable HTTP:
	//
	// Disable HTTP if HTTPS is enabled.
	DisableHttp bool `json:"disableHttp,omitempty"`
	// HTTPS certificate and private key Secret name:
	//
	// Name of a Secret that contains HTTPS certificate under the `tls.crt` key,
	// and the private key under the `tls.key` key.
	SecretName string `json:"secretName,omitempty"`
}

type ApicurioRegistrySpecConfigurationSecurityKeycloak struct {
	// Keycloak auth URL:
	//
	// URL of the Keycloak auth endpoint, must end with `/auth`.
	Url string `json:"url,omitempty"`
	// Keycloak realm
	Realm string `json:"realm,omitempty"`
	// Client ID for the REST API
	ApiClientId string `json:"apiClientId,omitempty"`
	// Client ID for the UI
	UiClientId string `json:"uiClientId,omitempty"`
}

type ApicurioRegistrySpecDeploymentMetadata struct {
	// Annotations:
	//
	// Additional Apicurio Registry Pod annotations.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels:
	//
	// Additional Apicurio Registry Pod labels.
	Labels map[string]string `json:"labels,omitempty"`
}

type ApicurioRegistrySpecDeployment struct {
	// Replicas:
	//
	// The required number of Apicurio Registry pods. Default value is 1.
	Replicas int32 `json:"replicas,omitempty"`
	// Hostname:
	//
	// Apicurio Registry application hostname (part of the URL without the protocol and path).
	Host string `json:"host,omitempty"`
	// Affinity
	Affinity *core.Affinity `json:"affinity,omitempty"`
	// Tolerations
	Tolerations []core.Toleration `json:"tolerations,omitempty"`
	// Metadata of the Apicurio Registry pod
	Metadata ApicurioRegistrySpecDeploymentMetadata `json:"metadata,omitempty"`
	// Apicurio Registry image:
	//
	// Replaces the default Apicurio Registry application image.
	// Overrides the values in the REGISTRY_IMAGE_MEM, REGISTRY_IMAGE_KAFKASQL and REGISTRY_IMAGE_SQL Operator environment variables.
	Image string `json:"image,omitempty"`
	// Apicurio Registry image pull secrets:
	//
	// List of Secrets to use when pulling the Apicurio Registry image.
	ImagePullSecrets []core.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Apicurio Registry managed resources:
	//
	// Configure how the Operator manages Kubernetes resources.
	ManagedResources ApicurioRegistrySpecDeploymentManagedResources `json:"managedResources,omitempty"`
	// Configure Apicurio Registry pod template:
	//
	// With some restrictions, the Apicurio Registry Operator forwards the data from this field
	// to the corresponding "spec.template" field in the Apicurio Registry Deployment.
	// This feature provides greater configuration flexibility, without the need for the Operator to natively support each use case.
	// WARNING: This feature is a Technology Preview feature only.
	PodTemplateSpecPreview ApicurioRegistryPodTemplateSpec `json:"podTemplateSpecPreview,omitempty"`
	// Additional configuration for the operator-managed Ingress.
	Ingress ApicurioRegistrySpecDeploymentIngress `json:"ingress,omitempty"`
}

type ApicurioRegistrySpecDeploymentIngress struct {
	// Additional annotations for the operator-managed Ingress.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Name of an IngressClass cluster resource. Ingress
	// controller implementations use this field to know whether they should be
	// serving this Ingress resource, by a transitive connection
	// (controller -> IngressClass -> Ingress resource).
	//
	// See https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class
	IngressClassName string `json:"ingressClassName,omitempty"`
}

type ApicurioRegistrySpecDeploymentManagedResources struct {
	// Disable Ingress:
	//
	// Operator will not create or manage an Ingress for Apicurio Registry, so it can be done manually.
	DisableIngress bool `json:"disableIngress,omitempty"`
	// Disable NetworkPolicy:
	//
	// Operator will not create or manage a NetworkPolicy for Apicurio Registry, so it can be done manually.
	DisableNetworkPolicy bool `json:"disableNetworkPolicy,omitempty"`
	// Disable PodDisruptionBudget:
	//
	// Operator will not create or manage a PodDisruptionBudget for Apicurio Registry, so it can be done manually.
	DisablePodDisruptionBudget bool `json:"disablePodDisruptionBudget,omitempty"`
}

// ### Status

type ApicurioRegistryStatus struct {
	// Information about the Apicurio Registry application
	Info ApicurioRegistryStatusInfo `json:"info,omitempty"`
	// Conditions:
	//
	// Apicurio Registry application and Operator conditions.
	Conditions []meta.Condition `json:"conditions,omitempty"`
	// Managed Resources:
	//
	// Kubernetes resources managed by the Apicurio Registry Operator.
	ManagedResources []ApicurioRegistryStatusManagedResource `json:"managedResources,omitempty"`
}

type ApicurioRegistryStatusInfo struct {
	// Apicurio Registry URL
	Host string `json:"host,omitempty"`
}

type ApicurioRegistryStatusManagedResource struct {
	Kind      string `json:"kind,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// ### Roots

// ApicurioRegistry represents an Apicurio Registry instance
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type ApicurioRegistry struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApicurioRegistrySpec   `json:"spec,omitempty"`
	Status ApicurioRegistryStatus `json:"status,omitempty"`
}

// ApicurioRegistryList contains a list of ApicurioRegistry
// +kubebuilder:object:root=true
type ApicurioRegistryList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ApicurioRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApicurioRegistry{}, &ApicurioRegistryList{})
}

// These are slightly modified copies of core.PodTemplateSpec and some nested structs,
// for the purpose of:
// - allowing some fields to be empty or nil,
// - for generating a better CRD,
// - working deep equality operation
//
// The modified struct must be de/serializable from/to the original PodTemplateSpec via JSON.
//
// Comments removed to avoid an error "Too long: must have at most 262144 bytes" when executing "kubectl apply".
// By using kubectl apply to create/update resources, an annotation "kubectl.kubernetes.io/last-applied-configuration"
// is created by K8s API to store the latest version of the resource.
// However, it has a size limit and if the CRD has many long descriptions, it will result the error.

// ApicurioRegistryPodTemplateSpec describes the data a pod should have when created from a template
type ApicurioRegistryPodTemplateSpec struct {

	// +optional
	Metadata ApicurioRegistryObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// +optional
	Spec ApicurioRegistryPodSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// ApicurioRegistryObjectMeta is metadata that all persisted resources must have, which includes all objects
// users must create.
type ApicurioRegistryObjectMeta struct {

	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// +optional
	GenerateName string `json:"generateName,omitempty" protobuf:"bytes,2,opt,name=generateName"`

	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`

	// +optional
	SelfLink string `json:"selfLink,omitempty" protobuf:"bytes,4,opt,name=selfLink"`

	// +optional
	UID types.UID `json:"uid,omitempty" protobuf:"bytes,5,opt,name=uid,casttype=k8s.io/kubernetes/pkg/types.UID"`

	// +optional
	ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,6,opt,name=resourceVersion"`

	// +optional
	Generation int64 `json:"generation,omitempty" protobuf:"varint,7,opt,name=generation"`

	// +optional
	CreationTimestamp *meta.Time `json:"creationTimestamp,omitempty" protobuf:"bytes,8,opt,name=creationTimestamp"` // Modified

	// +optional
	DeletionTimestamp *meta.Time `json:"deletionTimestamp,omitempty" protobuf:"bytes,9,opt,name=deletionTimestamp"`

	// +optional
	DeletionGracePeriodSeconds *int64 `json:"deletionGracePeriodSeconds,omitempty" protobuf:"varint,10,opt,name=deletionGracePeriodSeconds"`

	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`

	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`

	// +optional
	// +patchMergeKey=uid
	// +patchStrategy=merge
	OwnerReferences []meta.OwnerReference `json:"ownerReferences,omitempty" patchStrategy:"merge" patchMergeKey:"uid" protobuf:"bytes,13,rep,name=ownerReferences"`

	// +optional
	// +patchStrategy=merge
	Finalizers []string `json:"finalizers,omitempty" patchStrategy:"merge" protobuf:"bytes,14,rep,name=finalizers"`

	// +optional
	ClusterName string `json:"clusterName,omitempty" protobuf:"bytes,15,opt,name=clusterName"`

	// +optional
	ManagedFields []meta.ManagedFieldsEntry `json:"managedFields,omitempty" protobuf:"bytes,17,rep,name=managedFields"`
}

// ApicurioRegistryPodSpec is a description of a pod.
type ApicurioRegistryPodSpec struct {

	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Volumes []core.Volume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,1,rep,name=volumes"`

	// +patchMergeKey=name
	// +patchStrategy=merge
	InitContainers []core.Container `json:"initContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,20,rep,name=initContainers"`

	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Containers []core.Container `json:"containers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"` // Modified

	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	EphemeralContainers []core.EphemeralContainer `json:"ephemeralContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,34,rep,name=ephemeralContainers"`

	// +optional
	RestartPolicy core.RestartPolicy `json:"restartPolicy,omitempty" protobuf:"bytes,3,opt,name=restartPolicy,casttype=RestartPolicy"`

	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty" protobuf:"varint,4,opt,name=terminationGracePeriodSeconds"`

	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty" protobuf:"varint,5,opt,name=activeDeadlineSeconds"`

	// +optional
	DNSPolicy core.DNSPolicy `json:"dnsPolicy,omitempty" protobuf:"bytes,6,opt,name=dnsPolicy,casttype=DNSPolicy"`

	// +optional
	// +mapType=atomic
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`

	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,8,opt,name=serviceAccountName"`

	// +k8s:conversion-gen=false
	// +optional
	DeprecatedServiceAccount string `json:"serviceAccount,omitempty" protobuf:"bytes,9,opt,name=serviceAccount"`

	// +optional
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty" protobuf:"varint,21,opt,name=automountServiceAccountToken"`

	// +optional
	NodeName string `json:"nodeName,omitempty" protobuf:"bytes,10,opt,name=nodeName"`

	// +k8s:conversion-gen=false
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty" protobuf:"varint,11,opt,name=hostNetwork"`

	// +k8s:conversion-gen=false
	// +optional
	HostPID bool `json:"hostPID,omitempty" protobuf:"varint,12,opt,name=hostPID"`

	// +k8s:conversion-gen=false
	// +optional
	HostIPC bool `json:"hostIPC,omitempty" protobuf:"varint,13,opt,name=hostIPC"`

	// +k8s:conversion-gen=false
	// +optional
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty" protobuf:"varint,27,opt,name=shareProcessNamespace"`

	// +optional
	SecurityContext *core.PodSecurityContext `json:"securityContext,omitempty" protobuf:"bytes,14,opt,name=securityContext"`

	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []core.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`

	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,16,opt,name=hostname"`

	// +optional
	Subdomain string `json:"subdomain,omitempty" protobuf:"bytes,17,opt,name=subdomain"`

	// +optional
	Affinity *core.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`

	// +optional
	SchedulerName string `json:"schedulerName,omitempty" protobuf:"bytes,19,opt,name=schedulerName"`

	// +optional
	Tolerations []core.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`

	// +optional
	// +patchMergeKey=ip
	// +patchStrategy=merge
	HostAliases []core.HostAlias `json:"hostAliases,omitempty" patchStrategy:"merge" patchMergeKey:"ip" protobuf:"bytes,23,rep,name=hostAliases"`

	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`

	// +optional
	Priority *int32 `json:"priority,omitempty" protobuf:"bytes,25,opt,name=priority"`

	// +optional
	DNSConfig *core.PodDNSConfig `json:"dnsConfig,omitempty" protobuf:"bytes,26,opt,name=dnsConfig"`

	// +optional
	ReadinessGates []core.PodReadinessGate `json:"readinessGates,omitempty" protobuf:"bytes,28,opt,name=readinessGates"`

	// +optional
	RuntimeClassName *string `json:"runtimeClassName,omitempty" protobuf:"bytes,29,opt,name=runtimeClassName"`

	// +optional
	EnableServiceLinks *bool `json:"enableServiceLinks,omitempty" protobuf:"varint,30,opt,name=enableServiceLinks"`

	// +optional
	PreemptionPolicy *core.PreemptionPolicy `json:"preemptionPolicy,omitempty" protobuf:"bytes,31,opt,name=preemptionPolicy"`

	// +optional
	Overhead core.ResourceList `json:"overhead,omitempty" protobuf:"bytes,32,opt,name=overhead"`

	// +optional
	// +patchMergeKey=topologyKey
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=topologyKey
	// +listMapKey=whenUnsatisfiable
	TopologySpreadConstraints []core.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty" patchStrategy:"merge" patchMergeKey:"topologyKey" protobuf:"bytes,33,opt,name=topologySpreadConstraints"`

	// +optional
	SetHostnameAsFQDN *bool `json:"setHostnameAsFQDN,omitempty" protobuf:"varint,35,opt,name=setHostnameAsFQDN"`

	// +optional
	OS *core.PodOS `json:"os,omitempty" protobuf:"bytes,36,opt,name=os"`
}
