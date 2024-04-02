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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ContextSecretBase struct {
	Name        string                `json:"name"`
	MetaName    string                `json:"metaName"`
	Origin      string                `json:"origin"`
	Type        string                `json:"type"`
	BDC         *BigDataClusterBase   `json:"bdc"`
	Properties  *runtime.RawExtension `json:"properties"`
	CreateTime  metav1.Time           `json:"createTime"`
	UpdateTime  metav1.Time           `json:"updateTime"`
	Labels      map[string]string     `json:"labels,omitempty"`
	Annotations map[string]string     `json:"annotations,omitempty"`
	Status      string                `json:"status"`
}

type GetContextSecretResponse struct {
	Data    *ContextSecretBase `json:"data"`
	Message string             `json:"message"`
	Status  int                `json:"status"`
}

type ListContextSecretsResponse struct {
	Data    []*ContextSecretBase `json:"data"`
	Message string               `json:"message"`
	Status  int                  `json:"status"`
}

type GetContextSecretDefSchemaResponse struct {
	Data    *runtime.RawExtension `json:"data"`
	Message string                `json:"message"`
	Status  int                   `json:"status"`
}
