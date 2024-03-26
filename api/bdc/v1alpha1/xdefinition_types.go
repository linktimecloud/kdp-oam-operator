/*
Copyright 2023 KDP(Kubernetes Data Platform).

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kdp-oam-operator/api/bdc/common"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Definition struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Type       string `json:"type,omitempty"`
}

type APIResource struct {
	Definition Definition `json:"definition"`
}

type ParameterMeta struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	RefType     string `json:"refType"`
	RefKey      string `json:"refKey"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// XDefinitionSpec defines the desired state of XDefinition
type XDefinitionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Status defines the custom health policy and status message for workload
	// +optional
	Status *common.Status `json:"status,omitempty"`
	// Schematic defines the data format and template of the encapsulation of the definition.
	// Only CUE schematic is supported for now.
	Schematic   *common.Schematic `json:"schematic"`
	APIResource APIResource       `json:"apiResource"`
	// +optional
	DynamicParameterMeta []ParameterMeta `json:"dynamicParameterMeta"`
}

// XDefinitionStatus defines the observed state of XDefinition
type XDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SchemaConfigMapRef          string `json:"schemaConfigMapRef"`
	SchemaConfigMapRefNamespace string `json:"schemaConfigMapRefNamespace"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="SchemaConfigMapRefNamespace",type="string",JSONPath=`.status.schemaConfigMapRefNamespace`
//+kubebuilder:printcolumn:name="SchemaConfigMapRef",type="string",JSONPath=`.status.schemaConfigMapRef`
//+kubebuilder:printcolumn:name="RelatedResourceAPIVersion",type="string",JSONPath=`.spec.apiResource.definition.apiVersion`
//+kubebuilder:printcolumn:name="RelatedResourceKind",type="string",JSONPath=`.spec.apiResource.definition.kind`
//+kubebuilder:printcolumn:name="RelatedResourceType",type="string",JSONPath=`.spec.apiResource.definition.type`

// XDefinition is the Schema for the xdefinitions API
type XDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   XDefinitionSpec   `json:"spec"`
	Status XDefinitionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// XDefinitionList contains a list of XDefinition
type XDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []XDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&XDefinition{}, &XDefinitionList{})
}
