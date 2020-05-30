// Package v1alpha1 contains API Schema definitions for the registry v1alpha1 API group
// +k8s:deepcopy-gen=package,register
// +groupName=apicur.io
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
)

const GroupName = "apicur.io"
const GroupResource = "apicurioregistries"
const GroupVersion = "v1alpha1"

var SchemeGroupResource = schema.GroupResource{Group: GroupName, Resource: GroupResource}
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: GroupVersion}

var SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
