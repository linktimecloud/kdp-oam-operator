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

import corev1 "k8s.io/api/core/v1"

var (
	// SystemName name of kdp
	SystemName = "kdp"
	// SystemDefaultNamespace global value for controller and webhook system-level namespace
	SystemDefaultNamespace = "kdp-system"
	// BDCControllerName means the controller is KDP BigDataCluster
	BDCControllerName = "kdp-bdc"
	// AppControllerName means the controller is KDP Application
	AppControllerName = "kdp-app"
	// KdpContextLabelKey means the label key of context cm
	KdpContextLabelKey = "kdp-operator-context"
	// KdpContextLabelValue means the label value of context cm
	KdpContextLabelValue = "KDP"
)

type ObjectReference struct {
	corev1.ObjectReference `json:",inline"`
}

// Equal check if two references are equal
func (in ObjectReference) Equal(r ObjectReference) bool {
	return in.APIVersion == r.APIVersion && in.Kind == r.Kind && in.Name == r.Name && in.Namespace == r.Namespace
}
