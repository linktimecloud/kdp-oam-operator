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
	"k8s.io/apimachinery/pkg/runtime"
	"kdp-oam-operator/api/bdc/condition"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ContextSettingSpec defines the desired state of ContextSetting
type ContextSettingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name       string                `json:"name"`
	Type       string                `json:"type,omitempty"`
	Properties *runtime.RawExtension `json:"properties,omitempty"`
}

// ContextSettingStatus defines the observed state of ContextSetting
type ContextSettingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status string `json:"status"`
	// ConditionedStatus reflects the observed status of a resource
	condition.ConditionedStatus `json:",inline"`
	SchemaConfigMapRef          string `json:"schemaConfigMapRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ContextSetting is the Schema for the contextsettings API
type ContextSetting struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContextSettingSpec   `json:"spec"`
	Status ContextSettingStatus `json:"status,omitempty"`
}

// SetConditions set condition for ContextSetting
func (in *ContextSetting) SetConditions(c ...condition.Condition) {
	in.Status.SetConditions(c...)
}

// GetCondition gets condition from ContextSetting
func (in *ContextSetting) GetCondition(conditionType condition.ConditionType) condition.Condition {
	return in.Status.GetCondition(conditionType)
}

//+kubebuilder:object:root=true

// ContextSettingList contains a list of ContextSetting
type ContextSettingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContextSetting `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ContextSetting{}, &ContextSettingList{})
}
