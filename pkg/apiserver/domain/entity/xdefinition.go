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
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
)

type XDefinitionEntity struct {
	Name                        string `json:"Name"`
	Description                 string `json:"Description"`
	SchemaConfigMapRef          string `json:"SchemaConfigMapRef"`
	SchemaConfigMapRefNamespace string `json:"SchemaConfigMapRefNamespace"`
	JSONSchema                  string `json:"JSONSchema"`
	UISchema                    string `json:"UISchema"`
}

func Object2XDefinitionEntity(def *bdcv1alpha1.XDefinition) *XDefinitionEntity {
	appEntity := &XDefinitionEntity{
		Name:                        def.Name,
		Description:                 def.Annotations[constants.AnnotationDefinitionDescription],
		SchemaConfigMapRef:          def.Status.SchemaConfigMapRef,
		SchemaConfigMapRefNamespace: def.Status.SchemaConfigMapRefNamespace,
	}
	return appEntity
}
