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
)

type BigDataClusterModel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              bdcv1alpha1.BigDataClusterSpec `json:"spec"`
}

type CreateBigDataClusterRequest struct {
	Org string `json:"org"`
	bdcv1alpha1.BigDataCluster
}

type UpdateBigDataClusterRequest struct {
	Properties string `json:"properties,omitempty"`
}

type BigDataClusterBase struct {
	Name        string            `json:"name"`
	DefaultNS   string            `json:"defaultNS"`
	Alias       string            `json:"alias"`
	Description string            `json:"description"`
	OrgName     string            `json:"orgName"`
	Status      string            `json:"status"`
	CreateTime  metav1.Time       `json:"createTime"`
	UpdateTime  metav1.Time       `json:"updateTime"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type GetBigDataClusterResponse struct {
	Data    *BigDataClusterBase `json:"data"`
	Message string              `json:"message"`
	Status  int                 `json:"status"`
}

// ListBigDataClustersResponse list bigdata clusters by query params
type ListBigDataClustersResponse struct {
	Data    []*BigDataClusterBase `json:"data"`
	Message string                `json:"message"`
	Status  int                   `json:"status"`
}

type BigDataClusterStatus struct {
	Status bdcv1alpha1.BigDataClusterStatus `json:"status,omitempty"`
}
