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

package common

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type CUE struct {
	// Template defines the abstraction template data of the capability, it will replace the old CUE template in extension field.
	// Template is a required field if CUE is defined in Capability Definition.
	Template string `json:"template"`
}

// Schematic defines the encapsulation of this capability
type Schematic struct {
	CUE *CUE `json:"cue,omitempty"`
}

type Status struct {
	// CustomStatus defines the custom status message that could display to user
	// +optional
	CustomStatus string `json:"customStatus,omitempty"`
	// HealthPolicy defines the health check policy for the abstraction
	// +optional
	HealthPolicy string `json:"healthPolicy,omitempty"`
}

// BDCObjectReference defines the object reference with cluster.
type BDCObjectReference struct {
	corev1.ObjectReference `json:",inline"`
	BDCName                string               `json:"bdcName,omitempty"`
	Status                 runtime.RawExtension `json:"status,omitempty"`
}

// ClusterObjectReference defines the object reference with cluster.
type ClusterObjectReference struct {
	Cluster                string `json:"cluster,omitempty"`
	Creator                string `json:"creator,omitempty"`
	corev1.ObjectReference `json:",inline"`
}

// ApplicationComponentStatus record the health status of App component
type ApplicationComponentStatus struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Cluster   string `json:"cluster,omitempty"`
	Env       string `json:"env,omitempty"`
	// WorkloadDefinition is the definition of a WorkloadDefinition, such as deployments/apps.v1
	WorkloadDefinition WorkloadGVK              `json:"workloadDefinition,omitempty"`
	Healthy            bool                     `json:"healthy"`
	Message            string                   `json:"message,omitempty"`
	Traits             []ApplicationTraitStatus `json:"traits,omitempty"`
	Scopes             []corev1.ObjectReference `json:"scopes,omitempty"`
}

// WorkflowStatus record the status of workflow
type WorkflowStatus struct {
	AppRevision string `json:"appRevision,omitempty"`
	Mode        string `json:"mode"`
	Message     string `json:"message,omitempty"`

	Suspend      bool   `json:"suspend"`
	SuspendState string `json:"suspendState,omitempty"`

	Terminated bool `json:"terminated"`
	Finished   bool `json:"finished"`

	ContextBackend *corev1.ObjectReference `json:"contextBackend,omitempty"`
	Steps          []WorkflowStepStatus    `json:"steps,omitempty"`

	StartTime metav1.Time `json:"startTime,omitempty"`
	// +nullable
	EndTime metav1.Time `json:"endTime,omitempty"`
}

type WorkflowStepStatus struct {
	StepStatus     `json:",inline"`
	SubStepsStatus []StepStatus `json:"subSteps,omitempty"`
}

type StepStatus struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	// A human readable message indicating details about why the workflowStep is in this state.
	Message string `json:"message,omitempty"`
	// A brief CamelCase message indicating details about why the workflowStep is in this state.
	Reason string `json:"reason,omitempty"`
	// FirstExecuteTime is the first time this step execution.
	FirstExecuteTime metav1.Time `json:"firstExecuteTime,omitempty"`
	// LastExecuteTime is the last time this step execution.
	LastExecuteTime metav1.Time `json:"lastExecuteTime,omitempty"`
}

// WorkloadGVK refer to a Workload Type
type WorkloadGVK struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

// ApplicationTraitStatus records the trait health status
type ApplicationTraitStatus struct {
	Type    string `json:"type"`
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}

const (
	// ApplicationInitializing means the app is preparing for initializing
	ApplicationInitializing string = "initializing"
	// ApplicationInitializeError means the initialize Error
	ApplicationInitializeError string = "initializeError"
	// ApplicationOutputDefError means outputs or outpts definition Error
	ApplicationOutputDefError string = "outputDefError"
	// ApplicationStarting means vela application starting but not running
	ApplicationStarting string = "starting"
	// ApplicationDeleting means the app is preparing for deleting
	ApplicationDeleting string = "deleting"
)

type ApplicationFinalPhase string

const (
	ApplicationFinalPhaseExecuting ApplicationFinalPhase = "executing"
	ApplicationFinalPhaseFailed    ApplicationFinalPhase = "failed"
	ApplicationFinalPhaseRunning   ApplicationFinalPhase = "running"
	ApplicationFinalPhaseException ApplicationFinalPhase = "exception"
	ApplicationFinalPhaseStopping  ApplicationFinalPhase = "stopping"
)
