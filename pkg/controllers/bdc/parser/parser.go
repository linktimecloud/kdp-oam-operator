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

package parser

import (
	"context"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"kdp-oam-operator/api/bdc/common"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	pkgcommon "kdp-oam-operator/pkg/common"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/controllers/bdc/defcontext"
	"kdp-oam-operator/pkg/controllers/bdc/deftemplate"
	"kdp-oam-operator/pkg/controllers/utils/uuid"
	"kdp-oam-operator/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Parser struct {
	Client      client.Client
	TemplLoader deftemplate.TemplateLoaderFn
}

// NewParser create parser
func NewParser(cli client.Client) *Parser {
	return &Parser{
		Client:      cli,
		TemplLoader: deftemplate.LoadTemplate,
	}
}

func (p *Parser) GenerateBigDataClusterFile(ctx context.Context, bdcObj runtime.Object) (*BDCFile, error) {
	var objName string
	var objAnnotations map[string]string
	var objLabels map[string]string
	var objSpec interface{}
	var objKind string
	var objUID string
	var downstreamNs string
	refDefName := ""
	specName := ""
	var err error

	switch bdcObject := bdcObj.(type) {
	case *bdcv1alpha1.BigDataCluster:
		objName = bdcObject.Name
		objAnnotations = bdcObject.Annotations
		objLabels = bdcObject.Labels
		objSpec = bdcObject.Spec
		objKind = bdcObject.Kind
		downstreamNs, _ = GenerateDownStreamNamespace(bdcObject)
		objUID = string(bdcObject.UID)
	case *bdcv1alpha1.ContextSecret:
		objName = bdcObject.Name
		objAnnotations = bdcObject.Annotations
		objLabels = bdcObject.Labels
		objSpec = bdcObject.Spec.Properties
		objKind = bdcObject.Kind
		if downstreamNs, err = p.GenerateOtherDownStreamNamespace(ctx, bdcObject.Annotations[constants.AnnotationBDCName]); err != nil {
			return nil, err
		}
		objUID = string(bdcObject.UID)
		refDefName = bdcObject.Spec.Type
		specName = bdcObject.Spec.Name
	case *bdcv1alpha1.ContextSetting:
		objName = bdcObject.Name
		objAnnotations = bdcObject.Annotations
		objLabels = bdcObject.Labels
		objSpec = bdcObject.Spec.Properties
		objKind = bdcObject.Kind
		if downstreamNs, err = p.GenerateOtherDownStreamNamespace(ctx, bdcObject.Annotations[constants.AnnotationBDCName]); err != nil {
			return nil, err
		}
		objUID = string(bdcObject.UID)
		refDefName = bdcObject.Spec.Type
		specName = bdcObject.Spec.Name
	case *bdcv1alpha1.Application:
		objName = bdcObject.Name
		objAnnotations = bdcObject.Annotations
		objLabels = bdcObject.Labels
		objSpec = bdcObject.Spec.Properties
		objKind = bdcObject.Kind
		if downstreamNs, err = p.GenerateOtherDownStreamNamespace(ctx, bdcObject.Annotations[constants.AnnotationBDCName]); err != nil {
			return nil, err
		}
		objUID = string(bdcObject.UID)
		refDefName = bdcObject.Spec.Type
		specName = bdcObject.Spec.Name
	}

	// generate a file object
	bdcFile := &BDCFile{
		BDCName:                   objName,
		Name:                      specName,
		CRUID:                     objUID,
		BigDataClusterLabels:      make(map[string]string),
		BigDataClusterAnnotations: make(map[string]string),
		RelatedXDefinitions:       make(map[string]*bdcv1alpha1.XDefinition),
		Parser:                    p,
		SetOwnerReference:         true,
	}
	for k, v := range objAnnotations {
		bdcFile.BigDataClusterAnnotations[k] = v
	}
	for k, v := range objLabels {
		bdcFile.BigDataClusterLabels[k] = v
	}
	bdcFile.DownStreamNamespace = downstreamNs

	// resolve and generate a template
	var templ *deftemplate.DefinitionTemplate
	templ, err = p.TemplLoader.LoadTemplate(ctx, p.Client, objKind, refDefName)
	if err != nil {
		return nil, errors.WithMessagef(err, "fail to load xdefition template")
	}

	//bdcTemplate, err := p.makeTemplate(ctx, objSpec)

	var settings map[string]interface{}
	settings, err = utils.Object2Map(objSpec)
	if err != nil {
		return nil, errors.WithMessagef(err, "fail to parse settings for")
	}
	bdcTemplate := &BDCTemplate{
		Name:              objName,
		SchematicCategory: templ.SchematicCategory,
		FullTemplate:      templ,
		Params:            settings,
		Engine:            deftemplate.NewBigDataClusterDefAbstractEngine(objName),
	}
	bdcTemplate.Ctx.Namespace = downstreamNs
	bdcFile.BDCTemplate = bdcTemplate

	if bdcTemplate.FullTemplate.XDefinition != nil {
		cd := bdcTemplate.FullTemplate.XDefinition.DeepCopy()
		cd.Status = bdcv1alpha1.XDefinitionStatus{}
		bdcFile.RelatedXDefinitions[bdcTemplate.FullTemplate.XDefinition.Name] = cd
	}

	return bdcFile, nil
}

type BDCFile struct {
	BDCName                   string // metadata.name
	Name                      string // spec.name
	DownStreamNamespace       string // output manifest namespace
	CRUID                     string
	BigDataClusterLabels      map[string]string
	BigDataClusterAnnotations map[string]string
	RelatedXDefinitions       map[string]*bdcv1alpha1.XDefinition
	ReferredObjects           []*unstructured.Unstructured
	Parser                    *Parser
	BDCTemplate               *BDCTemplate
	SetOwnerReference         bool
}

type BDCTemplate struct {
	Name              string
	SchematicCategory common.SchematicCategory
	Params            map[string]interface{}
	FullTemplate      *deftemplate.DefinitionTemplate
	Ctx               defcontext.ContextData
	Engine            deftemplate.AbstractEngine
}

func (bdcf *BDCFile) PrepareManifests(ctx context.Context, req ctrl.Request) (manifests []*unstructured.Unstructured, err error) {
	ctxData, err := GenerateContextDataFromBigDataClusterFile(bdcf, ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to generate context data")
	}
	bdcCtx := defcontext.NewBDCContext(ctxData)

	// generate manifest，context added last step will be used to render the output
	manifests, err = bdcf.EvalContext(bdcCtx)
	if err != nil {
		return nil, errors.Wrapf(err, "evaluate base template app=%s in namespace=%s", bdcf.Name, "")
	}
	commonLabels := SetCommonContextLabels(bdcCtx)
	for _, mf := range manifests {
		utils.AddLabels(mf, utils.MergeMapOverrideWithDst(commonLabels, map[string]string{constants.LabelReferredAPIResource: bdcf.BDCTemplate.FullTemplate.XDefinition.Spec.APIResource.Definition.Kind}))
		if bdcf.SetOwnerReference {
			ownerReference := []metav1.OwnerReference{{
				APIVersion:         bdcf.BDCTemplate.FullTemplate.XDefinition.Spec.APIResource.Definition.APIVersion,
				Kind:               bdcf.BDCTemplate.FullTemplate.XDefinition.Spec.APIResource.Definition.Kind,
				Name:               bdcf.BDCTemplate.Name,
				UID:                types.UID(bdcf.CRUID),
				Controller:         pointer.Bool(true),
				BlockOwnerDeletion: pointer.Bool(true),
			}}
			mf.SetOwnerReferences(ownerReference)
		}
	}

	return manifests, nil
}

func (bdcf *BDCFile) EvalContext(ctx defcontext.ContextData) ([]*unstructured.Unstructured, error) {
	return bdcf.BDCTemplate.Engine.RenderCUETemplate(ctx, bdcf.BDCTemplate.FullTemplate.TemplateStr, bdcf.BDCTemplate.Params)
}

// SetCommonContextLabels get base context labels
func SetCommonContextLabels(ctx defcontext.ContextData) map[string]string {
	baseLabels := ctx.BaseContextLabels()
	baseLabels[defcontext.ContextBDCName] = ctx.GetData(defcontext.ContextBDCName).(string)

	return baseLabels
}

func GenerateDownStreamNamespace(bdc *bdcv1alpha1.BigDataCluster) (string, error) {
	defaultNs := pkgcommon.SystemDefaultNamespace
	nss := bdc.Spec.Namespaces
	for _, ns := range nss {
		if ns.IsDefault {
			return ns.Name, nil
		}
	}

	return defaultNs, nil
}

func (p *Parser) GenerateOtherDownStreamNamespace(ctx context.Context, bdcInstanceName string) (string, error) {
	// Lookup BigDataCluster
	var bdcObj bdcv1alpha1.BigDataCluster
	err := p.Client.Get(ctx, client.ObjectKey{Name: bdcInstanceName}, &bdcObj)
	if err != nil && apierrors.IsNotFound(err) {
		klog.ErrorS(err, "not found", "BigDataCluster", bdcInstanceName)
		return pkgcommon.SystemDefaultNamespace, err
	}
	return GenerateDownStreamNamespace(&bdcObj)
}

func GenerateContextDataFromBigDataClusterFile(bdcFile *BDCFile, ctx context.Context, req ctrl.Request) (defcontext.ContextData, error) {
	data := defcontext.ContextData{
		Name:    bdcFile.Name,
		BDCName: bdcFile.BDCName,

		Namespace: bdcFile.DownStreamNamespace,
	}
	if bdcFile.BigDataClusterAnnotations != nil {
		data.BDCAnnotations = bdcFile.BigDataClusterAnnotations
		// 添加bdc group
		if bdcFile.BigDataClusterAnnotations[constants.AnnotationBDCName] != "" {
			data.PushData(defcontext.Bdc, bdcFile.BigDataClusterAnnotations[constants.AnnotationBDCName])
		}
		if bdcFile.BigDataClusterAnnotations[constants.AnnotationOrgName] != "" {
			data.PushData(defcontext.Group, bdcFile.BigDataClusterAnnotations[constants.AnnotationOrgName])
		}
	}
	if bdcFile.BigDataClusterLabels != nil {
		data.BDCLabels = bdcFile.BigDataClusterLabels
	}

	contextData, err := bdcFile.Parser.GetK8sBdcCtxNamespaced(ctx)
	if err != nil {
		return data, err
	}
	if contextData != nil {
		if contextData != nil {
			for key, val := range contextData {
				data.PushData(key, val)
			}
		}
	}

	data.PushData(defcontext.AppUuid, uuid.GenAppUUID(req.Namespace, req.Name, 8))

	return data, nil
}

func (p *Parser) GetK8sBdcCtxNamespaced(ctx context.Context) (map[string]string, error) {
	labelSelector := labels.Set(map[string]string{pkgcommon.KdpContextLabelKey: pkgcommon.KdpContextLabelValue}).AsSelector()
	listOptions := []client.ListOption{
		client.InNamespace(pkgcommon.SystemDefaultNamespace),
		client.MatchingLabelsSelector{Selector: labelSelector},
	}
	var configMaps corev1.ConfigMapList
	err := p.Client.List(ctx, &configMaps, listOptions...)
	if err != nil {
		klog.ErrorS(err, "fail to get context ConfigMap!")
		return nil, err
	}

	contextMap := make(map[string]string)
	for _, configMap := range configMaps.Items {
		contextMap = utils.MergeMapOverrideWithDst(contextMap, configMap.Data)
	}

	return contextMap, nil
}
