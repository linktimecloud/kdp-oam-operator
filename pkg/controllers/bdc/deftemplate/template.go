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
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"kdp-oam-operator/api/bdc/common"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	pkgcommon "kdp-oam-operator/pkg/common"
	defcontext "kdp-oam-operator/pkg/controllers/bdc/defcontext"
	cueutil "kdp-oam-operator/pkg/cue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// OutputFieldName is the reference of defcontext base object
	OutputFieldName = "output"
	// OutputsFieldName is the reference of defcontext Auxiliaries
	OutputsFieldName   = "outputs"
	ParameterFieldName = "parameter"
)

const (
	// CapabilityConfigMapNamePrefix is the prefix for capability ConfigMap name
	CapabilityConfigMapNamePrefix = "schema-"
	// OpenapiV3JSONSchema is the key to store OpenAPI v3 JSON schema in ConfigMap
	OpenapiV3JSONSchema string = "openapi-v3-json-schema"
)

// DefinitionTemplate is a helper struct for processing capability including
// It mainly collects schematic and status data of a capability definition.
type DefinitionTemplate struct {
	TemplateStr           string
	Health                string
	CustomStatus          string
	SchematicCategory     common.SchematicCategory
	XDefinition           *bdcv1alpha1.XDefinition
	XDefinitionSchemaName string
}

type TemplateLoaderFn func(context.Context, client.Reader, string, string) (*DefinitionTemplate, error)

// LoadTemplate load template of a capability definition
func (fn TemplateLoaderFn) LoadTemplate(ctx context.Context, c client.Reader, objKind string, refDefName string) (*DefinitionTemplate, error) {
	return fn(ctx, c, objKind, refDefName)
}

// ConvertTemplateJSON2Object convert spec.extension or spec.schematic to object
func ConvertTemplateJSON2Object(capabilityName string, in *runtime.RawExtension, schematic *common.Schematic) (common.Capability, error) {
	var t common.Capability
	t.Name = capabilityName
	if in != nil && in.Raw != nil {
		err := json.Unmarshal(in.Raw, &t)
		if err != nil {
			return t, errors.Wrapf(err, "parse extension fail")
		}
	}
	capTemplate := &DefinitionTemplate{}
	if err := loadSchematicToTemplate(capTemplate, nil, schematic); err != nil {
		return t, errors.WithMessage(err, "cannot resolve schematic")
	}
	if capTemplate.TemplateStr != "" {
		t.CueTemplate = capTemplate.TemplateStr
	}
	return t, nil
}

// LoadTemplate gets the capability definition from cluster and resolve it.
// It returns a helper struct, DefinitionTemplate, which will be used for further
// processing.
func LoadTemplate(ctx context.Context, cli client.Reader, objKind string, refDefName string) (*DefinitionTemplate, error) {
	bdcDef := new(bdcv1alpha1.XDefinition)
	err := GetBDCDefinition(ctx, cli, bdcDef, objKind, refDefName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// make error info correct
			return nil, errors.New(fmt.Sprintf("xdefinition \"%s\" not found", refDefName))
		}
		return nil, errors.WithMessagef(err, "load template from definition")
	}
	tmpl, err := newTemplateOfXDefinition(bdcDef)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func newTemplateOfXDefinition(bdcDef *bdcv1alpha1.XDefinition) (*DefinitionTemplate, error) {
	tmpl := &DefinitionTemplate{
		// Reference:           bdcDef.Spec.Workload,
		XDefinition:           bdcDef,
		XDefinitionSchemaName: bdcDef.Status.SchemaConfigMapRef,
	}
	if err := loadSchematicToTemplate(tmpl, bdcDef.Spec.Status, bdcDef.Spec.Schematic); err != nil {
		return nil, errors.WithMessage(err, "cannot load template")
	}
	return tmpl, nil
}

func loadSchematicToTemplate(tmpl *DefinitionTemplate, status *common.Status, schematic *common.Schematic) error {
	if status != nil {
		tmpl.CustomStatus = status.CustomStatus
		tmpl.Health = status.HealthPolicy
	}

	if schematic != nil {
		if schematic.CUE != nil {
			tmpl.SchematicCategory = common.CUECategory
			tmpl.TemplateStr = schematic.CUE.Template
		}

	}
	return nil
}

type ApiResourceDefMap struct {
	XDefName        string                  `json:"name"`
	XDefinition     bdcv1alpha1.XDefinition `json:"xDefinition"`
	APIResource     bdcv1alpha1.APIResource `json:"apiResource"`
	APIResourceKind string                  `json:"apiResourceKind"`
}

func GetBDCDefinition(ctx context.Context, cli client.Reader, definition client.Object, objKind string, refDefName string) error {
	bdcDefNs := GetDefinitionNamespaceWithCtx(ctx)
	cmName := fmt.Sprintf("%s", "bdc-definition-map")
	var cm v1.ConfigMap

	// Lookup Definition and APIResource map, get definition name
	err := cli.Get(ctx, client.ObjectKey{Namespace: bdcDefNs, Name: cmName}, &cm)
	if err != nil && apierrors.IsNotFound(err) {
		klog.ErrorS(err, "ConfigMap not found", "configMap", bdcDefNs, cmName)
		return err
	}
	bdcDefName := cm.Data[fmt.Sprintf("%s-%s", common.DefaultAPIResourceType, objKind)]
	if refDefName != "" {
		bdcDefName = cm.Data[fmt.Sprintf("%s-%s", refDefName, objKind)]
		if bdcDefName == "" {
			bdcDefName = refDefName
		}
	}

	if err := cli.Get(ctx, types.NamespacedName{Name: bdcDefName, Namespace: bdcDefNs}, definition); err != nil {
		if apierrors.IsNotFound(err) {
			return err
		}
		return err
	}
	return nil
}

type namespaceContextKey int

const (
	// BDCDefinitionNamespace is defcontext key to define app namespace
	BDCDefinitionNamespace namespaceContextKey = iota
)

func GetDefinitionNamespaceWithCtx(ctx context.Context) string {
	var BDCDefNs string
	if bdcDef := ctx.Value(BDCDefinitionNamespace); bdcDef == nil {
		BDCDefNs = pkgcommon.SystemDefaultNamespace
	} else {
		BDCDefNs = bdcDef.(string)
	}
	return BDCDefNs
}

func checkRequestNamespaceError(err error) bool {
	return err != nil && err.Error() == "an empty namespace may not be set when a resource name is provided"
}

type def struct {
	name string
}

type BigDataClusterDef struct {
	def
}

func NewBigDataClusterDefAbstractEngine(name string) AbstractEngine {
	return &BigDataClusterDef{
		def: def{
			name: name,
		},
	}
}

type AbstractEngine interface {
	RenderCUETemplate(ctx defcontext.ContextData, abstractTemplate string, params interface{}) ([]*unstructured.Unstructured, error)
}

func (wd *BigDataClusterDef) RenderCUETemplate(ctx defcontext.ContextData, abstractTemplate string, params interface{}) ([]*unstructured.Unstructured, error) {
	var paramFile = ParameterFieldName + ": {}"
	if params != nil {
		bt, err := json.Marshal(params)
		if err != nil {
			return nil, errors.WithMessagef(err, "marshal parameter of workload %s", wd.name)
		}
		if string(bt) != "null" {
			paramFile = fmt.Sprintf("%s: %s", ParameterFieldName, string(bt))
		}
	}

	var finalContext = strings.Builder{}
	// user custom parameter but be the first data and generated data should be appended at last
	// in case the user defined data has packages

	finalContext.WriteString(abstractTemplate + "\n")
	// parameter definition
	finalContext.WriteString(paramFile + "\n")

	baseCtx, err := ctx.BaseContextFile()
	if err != nil {
		return nil, err
	}
	finalContext.WriteString(baseCtx + "\n")

	finalContextStr := finalContext.String()

	var (
		c *cue.Context
		v cue.Value
	)
	// create a defcontext
	c = cuecontext.New()
	// compile some CUE into a Value
	v = c.CompileString(finalContextStr)
	output := v.Eval().LookupPath(cue.ParsePath(OutputFieldName))
	outputs := v.Eval().LookupPath(cue.ParsePath(OutputsFieldName))
	//klog.InfoS("BigDataClusterDefinition", "output manifests", output)
	//klog.InfoS("BigDataClusterDefinition", "outputs manifests", outputs)

	var finalOutputs []cue.Value
	finalOutputs = append(finalOutputs, output)

	iter, err := outputs.Fields(cue.Definitions(true), cue.Hidden(true), cue.All())
	if err != nil {
		return nil, errors.WithMessagef(err, "invalid outputs of workload %s", wd.name)
	}
	for iter.Next() {
		finalOutputs = append(finalOutputs, iter.Value())
	}

	err = v.Err()
	if err != nil {
		fmt.Println("Error during build:", v.Err())
		return nil, errors.WithMessagef(err, "Error during build: %s", err)
	}

	var workloads []*unstructured.Unstructured

	for _, fo := range finalOutputs {
		unstructuredOutput, err := cueutil.Unstructured(fo)
		if err != nil {
			return nil, errors.WithMessagef(err, "evaluate template %s", err)
		}
		workloads = append(workloads, unstructuredOutput)
	}

	return workloads, nil
}
