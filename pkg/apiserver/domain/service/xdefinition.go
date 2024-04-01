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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	v1types "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	entity "kdp-oam-operator/pkg/apiserver/domain/entity"
	"kdp-oam-operator/pkg/apiserver/exception"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	pkgutils "kdp-oam-operator/pkg/utils"
	"kdp-oam-operator/pkg/utils/log"
	"strings"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DefinitionQueryOption struct {
	RelatedResourceType string `json:"relatedResourceType"`
	RelatedResourceKind string `json:"relatedResourceKind"`
}

// XDefinitionService x-application service
type XDefinitionService interface {
	GetXDefinition(ctx context.Context, options DefinitionQueryOption, bdcName string) (*entity.XDefinitionEntity, error)
}

// NewXDefinitionService new x-application service
func NewXDefinitionService() XDefinitionService {
	kubeConfig, err := clients.GetKubeConfig()
	if err != nil {
		log.Logger.Fatalf("get kube config failure %s", err.Error())
	}
	kubeClient, err := clients.GetKubeClient()
	if err != nil {
		log.Logger.Fatalf("get kube client failure %s", err.Error())
	}
	return &xDefinitionServiceImpl{
		KubeClient: kubeClient,
		KubeConfig: kubeConfig,
	}
}

type xDefinitionServiceImpl struct {
	KubeClient client.Client
	KubeConfig *rest.Config
}

func (a xDefinitionServiceImpl) GetXDefinition(ctx context.Context, options DefinitionQueryOption, bdcName string) (*entity.XDefinitionEntity, error) {
	list := new(bdcv1alpha1.XDefinitionList)
	if err := a.KubeClient.List(ctx, list); err != nil {
		return nil, err
	}

	var defs *entity.XDefinitionEntity
	for _, item := range list.Items {
		if item.Spec.APIResource.Definition.Kind == options.RelatedResourceKind && item.Spec.APIResource.Definition.Type == options.RelatedResourceType {
			var cm v1.ConfigMap
			if err := a.KubeClient.Get(ctx, k8stypes.NamespacedName{
				Namespace: item.Status.SchemaConfigMapRefNamespace,
				Name:      item.Status.SchemaConfigMapRef,
			}, &cm); err != nil && !apierrors.IsNotFound(err) {
				return nil, err
			}
			defs = entity.Object2XDefinitionEntity(&item)
			updatedJSONSchema, err := a.renderXDefinitionDynamicParameter(ctx, &item, cm.Data["openapi-v3-json-schema"], bdcName)
			if err != nil {
				return nil, err
			}
			defs.JSONSchema = *updatedJSONSchema
			defs.UISchema = cm.Data["ui-schema"]

			break
		}
	}
	if defs == nil {
		return nil, exception.ErrDefinitionNotFound
	}

	return defs, nil
}

func (a xDefinitionServiceImpl) renderXDefinitionDynamicParameter(ctx context.Context, def *bdcv1alpha1.XDefinition, defaultSchema, bdcName string) (*string, error) {
	labels := map[string]string{}
	labels[constants.AnnotationBDCName] = bdcName

	listOptions := v1types.ListOptions{Labels: labels}

	dynamicParameterValues := make(map[string][]interface{})
	if def.Spec.DynamicParameterMeta != nil {
		// dynamic parameter
		for _, param := range def.Spec.DynamicParameterMeta {
			dynamicResKind := param.Type
			dynamicResRefType := param.RefType
			dynamicResRefKey := param.RefKey

			matchLabels := metav1.LabelSelector{MatchLabels: listOptions.Labels}

			dynamicResRefValues := make([]interface{}, 0)
			if dynamicResKind == kindContextSecret {
				list := new(bdcv1alpha1.ContextSecretList)
				selector, err := metav1.LabelSelectorAsSelector(&matchLabels)
				if err != nil {
					return nil, err
				}
				if err := a.KubeClient.List(ctx, list, &client.ListOptions{
					LabelSelector: selector,
				}); err != nil {
					return nil, err
				}
				for _, item := range list.Items {
					if dynamicResRefType == item.Spec.Type {
						var dynamicResRefValue interface{}
						if dynamicResRefKey == "" {
							dynamicResRefValue = item.Spec.Name
						} else {
							propertiesMap, err := pkgutils.RawExtension2Map(item.Spec.Properties)
							if err != nil {
								return nil, err
							}
							dynamicResRefValue = propertiesMap[dynamicResRefKey]
						}
						dynamicResRefValues = append(dynamicResRefValues, dynamicResRefValue)
					}
				}
			}
			if dynamicResKind == kindContextSetting {
				list := new(bdcv1alpha1.ContextSettingList)
				selector, err := metav1.LabelSelectorAsSelector(&matchLabels)
				if err != nil {
					return nil, err
				}
				if err := a.KubeClient.List(ctx, list, &client.ListOptions{
					LabelSelector: selector,
				}); err != nil {
					return nil, err
				}
				for _, item := range list.Items {
					if dynamicResRefType == item.Spec.Type {
						var dynamicResRefValue interface{}
						if dynamicResRefKey == "" {
							dynamicResRefValue = item.Spec.Name
						} else {
							propertiesMap, err := pkgutils.RawExtension2Map(item.Spec.Properties)
							if err != nil {
								return nil, err
							}
							dynamicResRefValue = propertiesMap[dynamicResRefKey]
						}
						dynamicResRefValues = append(dynamicResRefValues, dynamicResRefValue)
					}
				}
			}
			dynamicParameterValues[param.Name] = dynamicResRefValues
		}
	}

	return renderXDefinitionJSONSchem(defaultSchema, dynamicParameterValues)
}

func renderXDefinitionJSONSchem(defaultSchema string, dynamicParameterValues map[string][]interface{}) (*string, error) {
	schemaMap := pkgutils.StringToMap(defaultSchema)
	properties := schemaMap["properties"].(map[string]interface{})

	for key, value := range dynamicParameterValues {
		currentSchema := properties
		schemaPath := strings.Split(key, ".")
		for i, path := range schemaPath {
			if _, ok := currentSchema[path]; !ok {
				currentSchema[path] = make(map[string]interface{})
			}
			if i == len(schemaPath)-1 {
				if _, ok := currentSchema[path].(map[string]interface{})["enum"]; !ok {
					currentSchema[path].(map[string]interface{})["enum"] = []interface{}{}
				}
				currentEnum := currentSchema[path].(map[string]interface{})["enum"].([]interface{})
				for _, v := range value {
					if !contains(currentEnum, v) {
						currentEnum = append(currentEnum, v)
					}
				}
				currentSchema[path].(map[string]interface{})["enum"] = currentEnum
			} else {
				if _, ok := currentSchema[path].(map[string]interface{})["properties"]; !ok {
					currentSchema[path].(map[string]interface{})["properties"] = make(map[string]interface{})
				}
				currentSchema = currentSchema[path].(map[string]interface{})["properties"].(map[string]interface{})
			}
		}
	}
	updatedSchemaJSON, _ := json.Marshal(schemaMap)
	updatedSchemaJSONStr := string(updatedSchemaJSON)
	fmt.Println(string(updatedSchemaJSON))
	return &updatedSchemaJSONStr, nil
}

func contains(slice []interface{}, item interface{}) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
