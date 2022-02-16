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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ### Spec

// ApicurioRegistrySpec defines the desired state of ApicurioRegistry
type ApicurioRegistrySpec struct {
	Configuration ApicurioRegistrySpecConfiguration `json:"configuration,omitempty"`
	Deployment    ApicurioRegistrySpecDeployment    `json:"deployment,omitempty"`
}

type ApicurioRegistrySpecConfiguration struct {
	Persistence string                                    `json:"persistence,omitempty"`
	Sql         ApicurioRegistrySpecConfigurationSql      `json:"sql,omitempty"`
	Kafkasql    ApicurioRegistrySpecConfigurationKafkasql `json:"kafkasql,omitempty"`
	UI          ApicurioRegistrySpecConfigurationUI       `json:"ui,omitempty"`
	LogLevel    string                                    `json:"logLevel,omitempty"`
	Security    ApicurioRegistrySpecConfigurationSecurity `json:"security,omitempty"`
}

type ApicurioRegistrySpecConfigurationDataSource struct {
	Url      string `json:"url,omitempty"`
	UserName string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
}

type ApicurioRegistrySpecConfigurationSql struct {
	DataSource ApicurioRegistrySpecConfigurationDataSource `json:"dataSource,omitempty"`
}

type ApicurioRegistrySpecConfigurationKafkasql struct {
	BootstrapServers string                                         `json:"bootstrapServers,omitempty"`
	Security         ApicurioRegistrySpecConfigurationKafkaSecurity `json:"security,omitempty"`
}

type ApicurioRegistrySpecConfigurationKafkaSecurity struct {
	Tls   ApicurioRegistrySpecConfigurationKafkaSecurityTls   `json:"tls,omitempty"`
	Scram ApicurioRegistrySpecConfigurationKafkaSecurityScram `json:"scram,omitempty"`
}

type ApicurioRegistrySpecConfigurationKafkaSecurityTls struct {
	TruststoreSecretName string `json:"truststoreSecretName,omitempty"`
	KeystoreSecretName   string `json:"keystoreSecretName,omitempty"`
}

type ApicurioRegistrySpecConfigurationKafkaSecurityScram struct {
	TruststoreSecretName string `json:"truststoreSecretName,omitempty"`
	User                 string `json:"user,omitempty"`
	PasswordSecretName   string `json:"passwordSecretName,omitempty"`
	Mechanism            string `json:"mechanism,omitempty"`
}

type ApicurioRegistrySpecConfigurationUI struct {
	ReadOnly bool `json:"readOnly,omitempty"`
}

type ApicurioRegistrySpecConfigurationSecurity struct {
	Keycloak ApicurioRegistrySpecConfigurationSecurityKeycloak `json:"keycloak,omitempty"`
	// TLS cert and key used to enable SSL access in the deployed application.
	Tls ApicurioRegistrySpecConfigurationSecurityTls `json:"tls,omitempty"`
}

type ApicurioRegistrySpecConfigurationSecurityKeycloak struct {
	Url         string `json:"url,omitempty"`
	Realm       string `json:"realm,omitempty"`
	ApiClientId string `json:"apiClientId,omitempty"`
	UiClientId  string `json:"uiClientId,omitempty"`
}

type ApicurioRegistrySpecConfigurationSecurityTls struct {
	// The name of the Secret containing the TLS certificate and key
	SecretName string `json:"secretName,omitempty"`
	// The name of the certificate file in the Secret. Optional, defaults to 'tls.crt'
	Certificate string `json:"certificate,omitempty"`
	// The name of the private key file in the Secret. Optional, defaults to 'tls.key'
	Key string `json:"key,omitempty"`
}

type ApicurioRegistrySpecDeploymentMetadata struct {
	// Annotations added to the Deployment pod template.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels added to the Deployment pod template.
	Labels map[string]string `json:"labels,omitempty"`
}

type ApicurioRegistrySpecDeployment struct {
	Replicas    int32               `json:"replicas,omitempty"`
	Host        string              `json:"host,omitempty"`
	Affinity    *corev1.Affinity    `json:"affinity,omitempty"`
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Metadata applied to the Deployment pod template.
	Metadata ApicurioRegistrySpecDeploymentMetadata `json:"metadata,omitempty"`
	// Image set in the Deployment pod template. Overrides the values in the REGISTRY_IMAGE_MEM, REGISTRY_IMAGE_KAFKASQL and REGISTRY_IMAGE_SQL operator environment variables.
	Image string `json:"image,omitempty"`
	// List of secrets in the same namespace to use for pulling the Deployment pod image.
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// ### Status

type ApicurioRegistryStatus struct {
	// Information about the deployed application.
	Info ApicurioRegistryStatusInfo `json:"info,omitempty"`
	// List of status conditions.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// List of resources managed by this operator.
	ManagedResources []ApicurioRegistryStatusManagedResource `json:"managedResources,omitempty"`
}

type ApicurioRegistryStatusInfo struct {
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
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApicurioRegistrySpec   `json:"spec,omitempty"`
	Status ApicurioRegistryStatus `json:"status,omitempty"`
}

// ApicurioRegistryList contains a list of ApicurioRegistry
// +kubebuilder:object:root=true
type ApicurioRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApicurioRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApicurioRegistry{}, &ApicurioRegistryList{})
}
