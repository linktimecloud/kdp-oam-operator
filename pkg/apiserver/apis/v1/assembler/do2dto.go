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

package assembler

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/apiserver/domain/entity"
	"kdp-oam-operator/pkg/apiserver/domain/service"
	pkgutils "kdp-oam-operator/pkg/utils"
	"time"
)

func ConvertBigDataClusterEntityToDTO(entity *entity.BigDataClusterEntity) (*v1dto.BigDataClusterBase, error) {
	bdcBase := &v1dto.BigDataClusterBase{
		Name:        entity.Name,
		DefaultNS:   entity.DefaultNS,
		Alias:       entity.Alias,
		Description: entity.Description,
		OrgName:     entity.OrgName,
		Status:      entity.Status,
		CreateTime:  entity.CreateTime,
		UpdateTime:  entity.UpdateTime,
		Labels:      entity.Labels,
	}
	return bdcBase, nil
}

func ConvertApplicationEntityToDTO(entity *entity.ApplicationEntity) (*v1dto.ApplicationBase, error) {
	bdcBase, err := ConvertBigDataClusterEntityToDTO(entity.BDC)
	if err != nil {
		return nil, err
	}
	appStatus := pkgutils.Object2RawExtension(entity.Status)
	appBase := &v1dto.ApplicationBase{
		Name:            entity.Name,
		AppFormName:     entity.AppFormName,
		AppTemplateType: entity.AppTemplateType,
		AppRuntimeName:  entity.AppRuntimeName,
		AppRuntimeNs:    entity.AppRuntimeNs,
		Status:          appStatus,
		CreateTime:      entity.CreateTime,
		UpdateTime:      entity.UpdateTime,
		BDC:             bdcBase,
		Properties:      entity.Properties,
		Labels:          entity.Labels,
		Annotations:     entity.Annotations,
	}
	return appBase, nil
}

func ConvertContextSecretEntityToDTO(entity *entity.ContextSecretEntity) (*v1dto.ContextSecretBase, error) {
	bdcBase, err := ConvertBigDataClusterEntityToDTO(entity.BDC)
	if err != nil {
		return nil, err
	}
	ctxSecretBase := &v1dto.ContextSecretBase{
		Name:        entity.Name,
		MetaName:    entity.MetaName,
		Origin:      entity.Origin,
		Type:        entity.Type,
		CreateTime:  entity.CreateTime,
		UpdateTime:  entity.UpdateTime,
		Properties:  entity.Properties,
		BDC:         bdcBase,
		Labels:      entity.Labels,
		Annotations: entity.Annotations,
	}
	return ctxSecretBase, nil
}

func ConvertContextSettingEntityToDTO(entity *entity.ContextSettingEntity) (*v1dto.ContextSettingBase, error) {
	bdcBase, err := ConvertBigDataClusterEntityToDTO(entity.BDC)
	if err != nil {
		return nil, err
	}
	ctxSettingBase := &v1dto.ContextSettingBase{
		Name:        entity.Name,
		MetaName:    entity.MetaName,
		Origin:      entity.Origin,
		Type:        entity.Type,
		CreateTime:  entity.CreateTime,
		UpdateTime:  entity.UpdateTime,
		Properties:  entity.Properties,
		BDC:         bdcBase,
		Labels:      entity.Labels,
		Annotations: entity.Annotations,
	}
	return ctxSettingBase, nil
}

func ConvertXDefinitionEntityToDTO(entity *entity.XDefinitionEntity) (*v1dto.XDefinitionBase, error) {
	defBase := &v1dto.XDefinitionBase{
		Name:                        entity.Name,
		SchemaConfigMapRef:          entity.SchemaConfigMapRef,
		SchemaConfigMapRefNamespace: entity.SchemaConfigMapRefNamespace,
		Description:                 entity.Description,
	}
	defBase.JSONSchema = pkgutils.StringToMap(entity.JSONSchema)
	defBase.UISchema = pkgutils.StringToMap(entity.UISchema)
	return defBase, nil
}

func ConvertWebTerminalEntityToDTO(entity *unstructured.Unstructured) (*v1dto.TerminalBase, error) {
	rules, err := service.ParseExtractionRules()
	if err != nil {
		fmt.Println("parse transform file data err:", err)
		return nil, err
	}

	data, err := service.ExtractData(entity, rules)
	if err != nil {
		fmt.Println("Error extracting data:", err)
		return nil, err
	}

	accessUrl, phase, ttl, err := service.GetTerminalData(entity)
	if err != nil {
		fmt.Println("get terminal url by response err: ", err.Error())
		return nil, err
	}
	terminalUrl := service.GetTerminalUrl(accessUrl)

	terBase := &v1dto.TerminalBase{
		Name:       entity.GetName(),
		NameSpace:  entity.GetNamespace(),
		Phase:      phase,
		AccessUrl:  terminalUrl,
		CreateTime: entity.GetCreationTimestamp(),
		EndTime:    calculateEndTime(entity.GetCreationTimestamp(), data["ttl"].(int64)),
		Ttl:        ttl,
	}
	return terBase, nil
}

// calculateEndTime get end time
func calculateEndTime(createTime metav1.Time, ttl int64) metav1.Time {
	return metav1.NewTime(createTime.Add(time.Duration(ttl) * time.Second))
}
