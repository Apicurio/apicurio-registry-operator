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

package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApicurioRegistrySpec defines the desired state of ApicurioRegistry
type ApicurioRegistrySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ApicurioRegistry. Edit ApicurioRegistry_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// ApicurioRegistryStatus defines the observed state of ApicurioRegistry
type ApicurioRegistryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ApicurioRegistry is the Schema for the apicurioregistries API
type ApicurioRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApicurioRegistrySpec   `json:"spec,omitempty"`
	Status ApicurioRegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApicurioRegistryList contains a list of ApicurioRegistry
type ApicurioRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApicurioRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApicurioRegistry{}, &ApicurioRegistryList{})
}
