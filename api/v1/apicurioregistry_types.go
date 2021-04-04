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

type ApicurioRegistrySpecDeployment struct {
	Replicas    int32               `json:"replicas,omitempty"`
	Host        string              `json:"host,omitempty"`
	Affinity    *corev1.Affinity    `json:"affinity,omitempty"`
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// ### Status

// ApicurioRegistryStatus defines the observed state of ApicurioRegistry
type ApicurioRegistryStatus struct {
	Image          string `json:"image,omitempty"`
	DeploymentName string `json:"deploymentName,omitempty"`
	ServiceName    string `json:"serviceName,omitempty"`
	IngressName    string `json:"ingressName,omitempty"`
	ReplicaCount   int32  `json:"replicaCount,omitempty"`
	Host           string `json:"host,omitempty"`
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
