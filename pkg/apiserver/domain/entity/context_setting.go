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
	"encoding/json"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/utils/log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ContextSettingEntity bigdata cluster delivery model
type ContextSettingEntity struct {
	Name        string                `json:"name"`
	MetaName    string                `json:"metaName"`
	Origin      string                `json:"origin"`
	Type        string                `json:"type"`
	BDC         *BigDataClusterEntity `json:"bdc"`
	Properties  *runtime.RawExtension `json:"properties"`
	CreateTime  metav1.Time           `json:"createTime"`
	UpdateTime  metav1.Time           `json:"updateTime"`
	Labels      map[string]string     `json:"labels,omitempty"`
	Annotations map[string]string     `json:"annotations,omitempty"`
}

func Object2ContextSettingEntity(ctxSetting *bdcv1alpha1.ContextSetting) *ContextSettingEntity {
	var bdcEntity BigDataClusterEntity
	bdcEntity.Name = ctxSetting.Labels[constants.LabelBDCName]
	bdcEntity.OrgName = ctxSetting.Labels[constants.LabelOrgName]
	bdcEntity.DefaultNS = ctxSetting.Annotations[constants.AnnotationBDCDefaultNamespace]
	bdcAppliedCfg := ctxSetting.GetAnnotations()[constants.AnnotationBDCAppliedConfiguration]
	if bdcAppliedCfg != "" {
		err := json.Unmarshal([]byte(bdcAppliedCfg), &bdcEntity)
		if err != nil {
			log.Logger.Errorw("failed to unmarshal bdc applied configuration", "err", err)
			return nil
		}
	}
	updateTime, _ := time.Parse(time.RFC3339, ctxSetting.Annotations[constants.AnnotationBDCUpdatedTime])
	appEntity := &ContextSettingEntity{
		Name:       ctxSetting.Spec.Name,
		MetaName:   ctxSetting.Name,
		Origin:     ctxSetting.Annotations[constants.AnnotationCtxSettingOrigin],
		Type:       ctxSetting.Spec.Type,
		BDC:        &bdcEntity,
		Properties: ctxSetting.Spec.Properties,
		CreateTime: ctxSetting.CreationTimestamp,
		UpdateTime: metav1.NewTime(updateTime),
		Labels:     ctxSetting.Labels,
	}
	return appEntity
}
