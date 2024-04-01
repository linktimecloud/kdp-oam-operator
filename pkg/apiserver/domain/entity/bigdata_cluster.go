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

package entity

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"time"
)

// BigDataClusterEntity bigdata cluster delivery model
type BigDataClusterEntity struct {
	Name        string            `json:"name"`
	Alias       string            `json:"alias"`
	Description string            `json:"description"`
	OrgName     string            `json:"orgName"`
	CreateTime  metav1.Time       `json:"createTime"`
	UpdateTime  metav1.Time       `json:"updateTime"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Status      string            `json:"status"`
	DefaultNS   string            `json:"defaultNS"`
}

func Object2BigDataClusterEntity(bdc *bdcv1alpha1.BigDataCluster) *BigDataClusterEntity {
	updateTime, _ := time.Parse(time.RFC3339, bdc.Annotations[constants.AnnotationBDCUpdatedTime])
	return &BigDataClusterEntity{
		Name:        bdc.Name,
		DefaultNS:   bdc.Labels[constants.AnnotationBDCDefaultNamespace],
		Alias:       bdc.Annotations[constants.AnnotationBDCAlias],
		Description: bdc.Annotations[constants.AnnotationBDCDescription],
		OrgName:     bdc.Annotations[constants.AnnotationOrgName],
		Status:      string(bdc.Status.Status),
		CreateTime:  bdc.CreationTimestamp,
		UpdateTime:  metav1.NewTime(updateTime),
		Labels:      bdc.Labels,
		Annotations: bdc.Annotations,
	}
}
