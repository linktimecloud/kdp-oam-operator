package vela

import (
	velaWorkflowv1alpha1 "github.com/kubevela/workflow/api/v1alpha1"
	velacommon "github.com/oam-dev/kubevela/apis/core.oam.dev/common"
	velav1beta1 "github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
	"kdp-oam-operator/api/bdc/common"
	conditiontype "kdp-oam-operator/api/bdc/condition"
)

func DiffCondition(desiredConditions []conditiontype.Condition, observedConditions []conditiontype.Condition) bool {
	for _, desiredCondition := range desiredConditions {
		for _, observedCondition := range observedConditions {
			if desiredCondition.Type != observedCondition.Type {
				continue
			}
			if !observedCondition.Equal(desiredCondition) {
				return true
			}
		}
	}

	return false
}

func DiffWorkflowStatus(desiredWorkflowStatus *common.WorkflowStatus, observedWorkflowStatus *common.WorkflowStatus) bool {
	if desiredWorkflowStatus == nil && observedWorkflowStatus == nil {
		return false
	}

	if desiredWorkflowStatus == nil || observedWorkflowStatus == nil {
		return true
	}

	if desiredWorkflowStatus.AppRevision != observedWorkflowStatus.AppRevision ||
		desiredWorkflowStatus.Mode != observedWorkflowStatus.Mode ||
		desiredWorkflowStatus.Message != observedWorkflowStatus.Message ||
		desiredWorkflowStatus.Suspend != observedWorkflowStatus.Suspend ||
		desiredWorkflowStatus.SuspendState != observedWorkflowStatus.SuspendState ||
		desiredWorkflowStatus.Terminated != observedWorkflowStatus.Terminated ||
		desiredWorkflowStatus.Finished != observedWorkflowStatus.Finished ||
		desiredWorkflowStatus.ContextBackend != observedWorkflowStatus.ContextBackend ||
		desiredWorkflowStatus.StartTime != observedWorkflowStatus.StartTime ||
		desiredWorkflowStatus.EndTime != observedWorkflowStatus.EndTime {
		return true
	}

	return diffSteps(desiredWorkflowStatus.Steps, observedWorkflowStatus.Steps)
}

func diffSteps(desiredSteps []common.WorkflowStepStatus, observedSteps []common.WorkflowStepStatus) bool {
	if len(desiredSteps) != len(observedSteps) {
		return true
	}

	for i := 0; i < len(desiredSteps); i++ {
		if desiredSteps[i].ID != observedSteps[i].ID ||
			desiredSteps[i].Name != observedSteps[i].Name ||
			desiredSteps[i].Type != observedSteps[i].Type ||
			desiredSteps[i].Message != observedSteps[i].Message ||
			desiredSteps[i].Reason != observedSteps[i].Reason ||
			desiredSteps[i].FirstExecuteTime != observedSteps[i].FirstExecuteTime ||
			desiredSteps[i].LastExecuteTime != observedSteps[i].LastExecuteTime {
			return true
		}

		if diffSubSteps(desiredSteps[i].SubStepsStatus, observedSteps[i].SubStepsStatus) {
			return true
		}
	}

	return false
}

func diffSubSteps(desiredSubSteps []common.StepStatus, observedSubSteps []common.StepStatus) bool {
	if len(desiredSubSteps) != len(observedSubSteps) {
		return true
	}

	for i := 0; i < len(desiredSubSteps); i++ {
		if desiredSubSteps[i].ID != observedSubSteps[i].ID ||
			desiredSubSteps[i].Name != observedSubSteps[i].Name ||
			desiredSubSteps[i].Type != observedSubSteps[i].Type ||
			desiredSubSteps[i].Message != observedSubSteps[i].Message ||
			desiredSubSteps[i].Reason != observedSubSteps[i].Reason ||
			desiredSubSteps[i].FirstExecuteTime != observedSubSteps[i].FirstExecuteTime ||
			desiredSubSteps[i].LastExecuteTime != observedSubSteps[i].LastExecuteTime {
			return true
		}
	}

	return false
}

// parse velaApplication.Status.Conditions to application.Status.Conditions
func DesiredConditionFromVela(velaApplication *velav1beta1.Application) []conditiontype.Condition {
	var bdcCondition []conditiontype.Condition
	bdcCondition = append(bdcCondition, conditiontype.ReconcileVela())

	for _, velaCondition := range velaApplication.Status.Conditions {
		bdcCondition = append(bdcCondition, conditiontype.Condition{
			Type:               conditiontype.ConditionType(velaCondition.Type),
			Status:             velaCondition.Status,
			LastTransitionTime: velaCondition.LastTransitionTime,
			Reason:             conditiontype.ConditionReason(velaCondition.Reason),
			Message:            velaCondition.Message,
		})
	}

	return bdcCondition
}

func AppliedResourcesFormVela(velaAppliedResources []velacommon.ClusterObjectReference) []common.ClusterObjectReference {
	var appliedResources []common.ClusterObjectReference
	for _, resource := range velaAppliedResources {
		appliedResources = append(appliedResources, common.ClusterObjectReference{
			Cluster:         resource.Cluster,
			Creator:         resource.Creator,
			ObjectReference: resource.ObjectReference,
		})
	}

	return appliedResources
}

func ServicesFromVela(velaServices []velacommon.ApplicationComponentStatus) []common.ApplicationComponentStatus {
	var service []common.ApplicationComponentStatus
	for _, velaService := range velaServices {
		service = append(service, common.ApplicationComponentStatus{
			Name:      velaService.Name,
			Namespace: velaService.Namespace,
			Cluster:   velaService.Cluster,
			Env:       velaService.Env,
			WorkloadDefinition: common.WorkloadGVK{
				APIVersion: velaService.WorkloadDefinition.APIVersion,
				Kind:       velaService.WorkloadDefinition.Kind,
			},
			Healthy: velaService.Healthy,
			Message: velaService.Message,
			Traits:  traitsFromVela(velaService.Traits),
			Scopes:  velaService.Scopes,
		})
	}

	return service
}

func traitsFromVela(velaTraits []velacommon.ApplicationTraitStatus) []common.ApplicationTraitStatus {
	var traits []common.ApplicationTraitStatus
	for _, trait := range velaTraits {
		traits = append(traits, common.ApplicationTraitStatus{
			Type:    trait.Type,
			Healthy: trait.Healthy,
			Message: trait.Message,
		})
	}

	return traits
}

func WorkflowStatusFromVela(velaWorkflowStatus *velacommon.WorkflowStatus) *common.WorkflowStatus {
	if velaWorkflowStatus == nil {
		return nil
	}

	return &common.WorkflowStatus{
		AppRevision:    velaWorkflowStatus.AppRevision,
		Mode:           velaWorkflowStatus.Mode,
		Message:        velaWorkflowStatus.Message,
		Suspend:        velaWorkflowStatus.Suspend,
		SuspendState:   velaWorkflowStatus.SuspendState,
		Terminated:     velaWorkflowStatus.Terminated,
		Finished:       velaWorkflowStatus.Finished,
		ContextBackend: velaWorkflowStatus.ContextBackend,
		Steps:          stepsFromVela(velaWorkflowStatus.Steps),
		StartTime:      velaWorkflowStatus.StartTime,
		EndTime:        velaWorkflowStatus.EndTime,
	}
}

func stepsFromVela(velaSteps []velaWorkflowv1alpha1.WorkflowStepStatus) []common.WorkflowStepStatus {
	steps := make([]common.WorkflowStepStatus, 0)
	if velaSteps == nil || len(velaSteps) == 0 {
		return steps
	}

	for _, step := range velaSteps {
		steps = append(steps, stepFromVela(step))
	}

	return steps
}

func stepFromVela(velaSteps velaWorkflowv1alpha1.WorkflowStepStatus) common.WorkflowStepStatus {
	return common.WorkflowStepStatus{
		StepStatus: common.StepStatus{
			ID:               velaSteps.ID,
			Name:             velaSteps.Name,
			Type:             velaSteps.Type,
			Message:          velaSteps.Message,
			Reason:           velaSteps.Reason,
			FirstExecuteTime: velaSteps.FirstExecuteTime,
			LastExecuteTime:  velaSteps.LastExecuteTime,
		},
		SubStepsStatus: subStepsFromVela(velaSteps.SubStepsStatus),
	}
}

func subStepsFromVela(velaSubSteps []velaWorkflowv1alpha1.StepStatus) []common.StepStatus {
	var subSteps []common.StepStatus
	if velaSubSteps == nil || len(velaSubSteps) == 0 {
		return subSteps
	}

	for _, subStep := range velaSubSteps {
		subSteps = append(subSteps, common.StepStatus{
			ID:               subStep.ID,
			Name:             subStep.Name,
			Type:             subStep.Type,
			Message:          subStep.Message,
			Reason:           subStep.Reason,
			FirstExecuteTime: subStep.FirstExecuteTime,
			LastExecuteTime:  subStep.LastExecuteTime,
		})
	}

	return subSteps
}
