/*
Copyright 2024 KDP(Kubernetes Data Platform).

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

package dto

import (
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ApplicationSpecPropertiesMap struct {
	Properties map[string]interface{} `json:"properties"`
}

type CreateApplicationRequestModel struct {
	AppFormName     string `json:"appFormName"`
	AppTemplateType string `json:"appTemplateType"`
	ApplicationSpecPropertiesMap
}

type UpdateApplicationRequestModel struct {
	ApplicationSpecPropertiesMap
}

type ApplicationSpecProperties struct {
	Properties *runtime.RawExtension `json:"properties"`
}

type CreateApplicationRequestBody struct {
	AppFormName     string `json:"appFormName"`
	AppTemplateType string `json:"appTemplateType"`
	ApplicationSpecProperties
}

type CreateApplicationRequest struct {
	CreateApplicationRequestBody
	BDC *BigDataClusterBase `json:"bdc,omitempty"`
}

type UpdateApplicationRequest struct {
	AppName string              `json:"appName,omitempty"`
	BDC     *BigDataClusterBase `json:"bdc,omitempty"`
	UpdateApplicationRequestBody
}

type UpdateApplicationRequestBody struct {
	ApplicationSpecProperties
}

type ApplicationBase struct {
	Name            string                `json:"name"`
	AppFormName     string                `json:"appFormName"`
	AppTemplateType string                `json:"appTemplateType"`
	AppRuntimeName  string                `json:"appRuntimeName"`
	AppRuntimeNs    string                `json:"appRuntimeNs"`
	BDC             *BigDataClusterBase   `json:"bdc"`
	Properties      *runtime.RawExtension `json:"properties"`
	CreateTime      metav1.Time           `json:"createTime"`
	UpdateTime      metav1.Time           `json:"updateTime"`
	Labels          map[string]string     `json:"labels,omitempty"`
	Annotations     map[string]string     `json:"annotations,omitempty"`
	Status          *runtime.RawExtension `json:"status"`
}

type GetApplicationsResponse struct {
	Data    *ApplicationBase `json:"data"`
	Message string           `json:"message"`
	Status  int              `json:"status"`
}

// ListApplicationsResponse list applications by query params
type ListApplicationsResponse struct {
	Data    []*ApplicationBase `json:"data"`
	Message string             `json:"message"`
	Status  int                `json:"status"`
}

type ApplicationStatus struct {
	// AppliedResources record the resources that the  workflow step apply.
	AppliedResources bdcv1alpha1.ApplicationStatus `json:"appliedResources,omitempty"`
}

type ApplicationRawResponse struct {
	Data    *runtime.RawExtension `json:"data"`
	Message string                `json:"message"`
	Status  int                   `json:"status"`
}
