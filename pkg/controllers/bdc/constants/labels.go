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
	LabelReferredAPIResource = "api-resource.bdc.kdp.io/type"
	// LabelDefinition is the label for definition
	LabelDefinition = "definition.bdc.kdp.io"
	// LabelDefinitionName is the label for definition name
	LabelDefinitionName = "definition.bdc.kdp.io/name"
	LabelBDCOrgName     = "bdc.kdp.io/org"
	LabelAppFormName    = "form.app.bdc.kdp.io/name"
	LabelAppRuntimeName = "runtime.app.bdc.kdp.io/name"
	LabelBDCName        = AnnotationBDCName
	LabelOrgName        = AnnotationOrgName
	LabelName           = "terminal.bdc.kdp.io/name"

	AnnotationCtxSettingSource = "setting.ctx.bdc.kdp.io/source"
	AnnotationCtxSettingType   = "setting.ctx.bdc.kdp.io/type"
)
