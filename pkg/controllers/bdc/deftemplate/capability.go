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

package deftemplate

import (
	"bytes"
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/encoding/openapi"
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"kdp-oam-operator/api/bdc/common"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	pkgcommon "kdp-oam-operator/pkg/common"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CapabilityBaseDefinition struct {
}

type CapabilityDefinition struct {
	Name        string                  `json:"name"`
	XDefinition bdcv1alpha1.XDefinition `json:"xDefinition"`
	CapabilityBaseDefinition
}

// NewCapabilityXDef will create a CapabilityXDefinition
func NewCapabilityXDef(xDefinition *bdcv1alpha1.XDefinition) CapabilityDefinition {
	var def CapabilityDefinition
	def.Name = xDefinition.Name
	def.XDefinition = *xDefinition.DeepCopy()
	return def
}

// GetOpenAPISchema gets OpenAPI v3 schema by StepDefinition name
func (def *CapabilityDefinition) GetOpenAPIAndUischemaSchema(name string) ([]byte, []byte, error) {
	capability, err := ConvertTemplateJSON2Object(name, nil, def.XDefinition.Spec.Schematic)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert WorkflowStepDefinition to Capability Object")
	}
	return getSchemas(capability)
}

// StoreOpenAPISchema stores OpenAPI v3 schema from StepDefinition in ConfigMap
func (def *CapabilityDefinition) StoreOpenAPISchema(ctx context.Context, k8sClient client.Client, namespace, name string) (string, error) {
	jsonSchema, uiSchema, err := def.GetOpenAPIAndUischemaSchema(name)
	if err != nil {
		return "", fmt.Errorf("failed to generate OpenAPI v3 JSON schema for capability %s: %w", def.Name, err)
	}

	resourceDefinition := def.XDefinition
	ownerReference := []metav1.OwnerReference{{
		APIVersion:         resourceDefinition.APIVersion,
		Kind:               resourceDefinition.Kind,
		Name:               resourceDefinition.Name,
		UID:                resourceDefinition.GetUID(),
		Controller:         pointer.Bool(true),
		BlockOwnerDeletion: pointer.Bool(true),
	}}
	data := map[string]string{
		common.OpenapiV3JSONSchema: string(jsonSchema),
		common.UISchema:            string(uiSchema),
	}
	cmName, err := def.CreateOrUpdateConfigMap(ctx, k8sClient, namespace, resourceDefinition, data, ownerReference)
	if err != nil {
		return cmName, err
	}

	return cmName, nil
}

func (def *CapabilityBaseDefinition) CreateOrUpdateConfigMap(ctx context.Context, k8sClient client.Client, namespace string,
	definition bdcv1alpha1.XDefinition, data map[string]string, ownerReferences []metav1.OwnerReference) (string, error) {
	apiResourceDefinitionType := common.DefaultAPIResourceType
	if definition.Spec.APIResource.Definition.Type != "" {
		apiResourceDefinitionType = definition.Spec.APIResource.Definition.Type
	}
	cmName := fmt.Sprintf("%s%s-%s", common.CapabilityConfigMapNamePrefix, definition.Name, apiResourceDefinitionType)
	var cm v1.ConfigMap
	labels := definition.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[constants.LabelDefinition] = "schema"
	labels[constants.LabelDefinitionName] = definition.Name
	annotations := make(map[string]string)

	// No need to check the existence of namespace, if it doesn't exist, API server will return the error message
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: cmName}, &cm)
	if err != nil && apierrors.IsNotFound(err) {
		cm = v1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            cmName,
				Namespace:       pkgcommon.SystemDefaultNamespace,
				OwnerReferences: ownerReferences,
				Labels:          labels,
				Annotations:     annotations,
			},
			Data: data,
		}
		err = k8sClient.Create(ctx, &cm)
		if err != nil {
			return cmName, fmt.Errorf(constants.ErrUpdateCapabilityInConfigMap, definition.Name, err)
		}
		klog.InfoS("Successfully stored Capability Schema in ConfigMap", "configMap", klog.KRef(namespace, cmName))
		return cmName, nil
	}

	cm.Data = data
	cm.Labels = labels
	cm.Annotations = annotations
	if err = k8sClient.Update(ctx, &cm); err != nil {
		return cmName, fmt.Errorf(constants.ErrUpdateCapabilityInConfigMap, definition.Name, err)
	}
	klog.InfoS("Successfully update Capability Schema in ConfigMap", "configMap", klog.KRef(namespace, cmName))
	return cmName, nil
}

func getSchemas(capability common.Capability) ([]byte, []byte, error) {
	openAPISchema, err := generateOpenAPISchemaFromCapabilityParameter(capability)
	if err != nil {
		return nil, nil, err
	}
	schema, err := ConvertOpenAPISchema2SwaggerObject(openAPISchema)
	if err != nil {
		return nil, nil, err
	}

	uiSchema := GetAdditionUiSchema(schema)

	FixOpenAPISchema(schema)
	parameter, err := schema.MarshalJSON()
	if err != nil {
		return nil, nil, err
	}
	return parameter, uiSchema, nil

}

const BaseTemplate = `
context: {
 name: string
 config?: [...{
   name: string
   value: string
 }]
 ...
}
`

func ParseToCUEValue(s string) (*cue.Value, error) {
	// the cue script must be first, it could include the imports
	template := s + "\n" + BaseTemplate
	bi := build.NewContext().NewInstance("", nil)

	file, err := parser.ParseFile("-", template, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	if err := bi.AddSyntax(file); err != nil {
		return nil, err
	}
	inst := cuecontext.New().BuildInstance(bi)
	if err != nil {
		return nil, fmt.Errorf("fail to parse the template:%w", err)
	}
	return &inst, nil
}

func FillParameterDefinitionFieldIfNotExist(val cue.Value) cue.Value {
	defaultValue := cuecontext.New().CompileString("#parameter: {}")
	defPath := cue.ParsePath("#" + ParameterFieldName)
	if paramVal := val.LookupPath(cue.ParsePath(ParameterFieldName)); paramVal.Exists() {
		if paramVal.IncompleteKind() == cue.BottomKind {
			return defaultValue
		}
		paramOnlyVal := val.Context().CompileString("{}").FillPath(defPath, paramVal)
		return paramOnlyVal
	}
	return defaultValue
}

func generateOpenAPISchemaFromCapabilityParameter(capability common.Capability) ([]byte, error) {
	val, err := ParseToCUEValue(capability.CueTemplate)
	if err != nil {
		return nil, err
	}
	paramOnlyVal := FillParameterDefinitionFieldIfNotExist(*val)
	defaultConfig := &openapi.Config{}
	b, err := openapi.Gen(paramOnlyVal, defaultConfig)
	if err != nil {
		return nil, err
	}
	var out = &bytes.Buffer{}
	_ = json.Indent(out, b, "", "   ")
	return out.Bytes(), nil

}

func ConvertOpenAPISchema2SwaggerObject(data []byte) (*openapi3.Schema, error) {
	openapiDoc, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		return nil, err
	}
	schemaRef, ok := openapiDoc.Components.Schemas[ParameterFieldName]
	if !ok {
		return nil, errors.New(constants.ErrGenerateOpenAPIV3JSONSchemaForCapability)
	}
	return schemaRef.Value, nil
}
