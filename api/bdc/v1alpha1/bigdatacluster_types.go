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
	"kdp-oam-operator/api/bdc/condition"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Namespace struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

// BigDataClusterSpec defines the desired state of BigDataCluster
type BigDataClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Frozen     bool        `json:"frozen"`
	Disabled   bool        `json:"disabled"`
	Namespaces []Namespace `json:"namespaces"`
}

// BigDataClusterStatusCategory defines the category of a status
type BigDataClusterStatusCategory string

// Active category of SchematicCategory
const (
	ActiveBigDataCluster      BigDataClusterStatusCategory = "Active"
	FrozenBigDataCluster      BigDataClusterStatusCategory = "Frozen"
	DisabledBigDataCluster    BigDataClusterStatusCategory = "Disabled"
	TerminatingBigDataCluster BigDataClusterStatusCategory = "Terminating"
)

// BigDataClusterStatus defines the observed state of BigDataCluster
type BigDataClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status BigDataClusterStatusCategory `json:"status"`
	// ConditionedStatus reflects the observed status of a resource
	condition.ConditionedStatus `json:",inline"`
	SchemaConfigMapRef          string `json:"schemaConfigMapRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.status`

// BigDataCluster is the Schema for the bigdataclusters API
type BigDataCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BigDataClusterSpec   `json:"spec"`
	Status BigDataClusterStatus `json:"status,omitempty"`
}

// SetConditions set condition for BigDataCluster
func (cd *BigDataCluster) SetConditions(c ...condition.Condition) {
	cd.Status.SetConditions(c...)
}

// GetCondition gets condition from BigDataCluster
func (cd *BigDataCluster) GetCondition(conditionType condition.ConditionType) condition.Condition {
	return cd.Status.GetCondition(conditionType)
}

//+kubebuilder:object:root=true

// BigDataClusterList contains a list of BigDataCluster
type BigDataClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BigDataCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BigDataCluster{}, &BigDataClusterList{})
}
