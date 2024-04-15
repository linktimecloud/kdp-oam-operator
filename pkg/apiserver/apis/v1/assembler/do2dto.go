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
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/apiserver/domain/entity"
	pkgutils "kdp-oam-operator/pkg/utils"
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

func ConvertWebTerminalEntityToDTO(entity *entity.WebTerminalEntity) (*v1dto.TerminalBase, error) {
	terBase := &v1dto.TerminalBase{
		Name:       entity.Name,
		NameSpace:  entity.NameSpace,
		Phase:      entity.Phase,
		AccessUrl:  entity.AccessUrl,
		CreateTime: entity.CreateTime,
		EndTime:    entity.EndTime,
		Ttl:        entity.Ttl,
	}
	return terBase, nil
}
