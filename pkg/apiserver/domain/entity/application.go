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
	"kdp-oam-operator/api/bdc/common"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/utils/log"
	"time"

	velacommon "github.com/oam-dev/kubevela/apis/core.oam.dev/common"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ApplicationEntity application delivery model
type ApplicationEntity struct {
	Name            string                        `json:"name"`
	AppFormName     string                        `json:"appFormName"`
	AppTemplateType string                        `json:"appTemplateType"`
	AppRuntimeName  string                        `json:"appRuntimeName"`
	AppRuntimeNs    string                        `json:"appRuntimeNs"`
	BDC             *BigDataClusterEntity         `json:"bdc"`
	Properties      *runtime.RawExtension         `json:"properties"`
	CreateTime      metav1.Time                   `json:"createTime"`
	UpdateTime      metav1.Time                   `json:"updateTime"`
	Labels          map[string]string             `json:"labels,omitempty"`
	Annotations     map[string]string             `json:"annotations,omitempty"`
	Status          bdcv1alpha1.ApplicationStatus `json:"status"`
}

func Object2ApplicationEntity(app *bdcv1alpha1.Application) *ApplicationEntity {
	var bdcEntity BigDataClusterEntity
	bdcEntity.Name = app.Labels[constants.LabelBDCName]
	bdcEntity.OrgName = app.Labels[constants.LabelOrgName]
	bdcEntity.DefaultNS = app.Annotations[constants.AnnotationBDCDefaultNamespace]
	bdcAppliedCfg := app.GetAnnotations()[constants.AnnotationBDCAppliedConfiguration]
	if bdcAppliedCfg != "" {
		err := json.Unmarshal([]byte(bdcAppliedCfg), &bdcEntity)
		if err != nil {
			log.Logger.Errorw("failed to unmarshal bdc applied configuration", "err", err)
			return nil
		}
	}
	updateTime, _ := time.Parse(time.RFC3339, app.Annotations[constants.AnnotationBDCUpdatedTime])

	appEntity := &ApplicationEntity{
		Name:            app.Name,
		AppFormName:     app.Spec.Name,
		AppTemplateType: app.Spec.Type,
		AppRuntimeName:  app.Labels[constants.LabelAppRuntimeName],
		AppRuntimeNs:    app.Annotations[constants.AnnotationBDCDefaultNamespace],
		BDC:             &bdcEntity,
		Properties:      app.Spec.Properties,
		CreateTime:      app.CreationTimestamp,
		UpdateTime:      metav1.NewTime(updateTime),
		Labels:          app.Labels,
		Annotations:     app.Annotations,
		Status:          app.Status,
	}
	appStatus := app.Status.Status
	var applicationFinalStatusPhase common.ApplicationFinalPhase
	if appStatus == common.ApplicationStarting || appStatus == common.ApplicationInitializing || appStatus == string(velacommon.ApplicationRendering) || appStatus == string(velacommon.ApplicationPolicyGenerating) || appStatus == string(velacommon.ApplicationRunningWorkflow) || appStatus == string(velacommon.ApplicationWorkflowSuspending) {
		applicationFinalStatusPhase = common.ApplicationFinalPhaseExecuting
	} else if appStatus == string(velacommon.ApplicationWorkflowTerminated) || appStatus == string(velacommon.ApplicationWorkflowFailed) {
		applicationFinalStatusPhase = common.ApplicationFinalPhaseFailed
	} else if appStatus == common.ApplicationInitializeError || appStatus == common.ApplicationOutputDefError || appStatus == string(velacommon.ApplicationUnhealthy) {
		applicationFinalStatusPhase = common.ApplicationFinalPhaseException
	} else if appStatus == common.ApplicationDeleting {
		applicationFinalStatusPhase = common.ApplicationFinalPhaseStopping
	} else if appStatus == string(velacommon.ApplicationRunning) {
		applicationFinalStatusPhase = common.ApplicationFinalPhaseRunning
	} else {
		applicationFinalStatusPhase = common.ApplicationFinalPhaseException
	}
	appEntity.Status.Status = string(applicationFinalStatusPhase)
	return appEntity
}
