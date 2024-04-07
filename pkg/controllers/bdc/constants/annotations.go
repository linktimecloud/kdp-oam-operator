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

package constants

const (
	// AnnotationBDCName is the annotation which describe what is the name of BigDataCluster
	AnnotationBDCName = "bdc.kdp.io/name"
	// AnnotationOrgName is the annotation describe the org which BigDataCluster belongs to
	AnnotationOrgName           = "bdc.kdp.io/org"
	AnnotationLastAppliedConfig = "bdc.kdp.io/last-applied-configuration"
	// AnnotationDefinitionDescription is the annotation which describe what is the capability used for in a Definition Object
	AnnotationDefinitionDescription = "definition.bdc.kdp.io/description"
	// AnnotationCtxSettingAdopt is the annotation which describe what is the capability used for in a Context Setting Object
	AnnotationCtxSettingAdopt = "setting.ctx.bdc.kdp.io/adopt"

	AnnotationBDCDefaultNamespace           = "bdc.kdp.io/namespace"
	AnnotationBDCAlias                      = "bdc.kdp.io/alias"
	AnnotationBDCDescription                = "bdc.kdp.io/description"
	AnnotationBDCUpdatedTime                = "bdc.kdp.io/updateTime"
	AnnotationBDCAppliedConfiguration       = "bdc.kdp.io/applied-configuration"
	AnnotationCtxSettingOrigin              = "setting.ctx.bdc.kdp.io/origin"
	AnnotationCtxSettingReferencedConfigMap = "setting.ctx.bdc.kdp.io/referenced-configmap"

	// FinalizerResourceTracker finalizer for gc
	FinalizerResourceTracker = "bdc.kdp.io/resource-tracker-finalizer"
)
