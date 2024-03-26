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

package condition

import (
	"context"
	"kdp-oam-operator/api/bdc/condition"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// A Conditioned may have conditions set or retrieved. Conditions are typically
// indicate the status of both a resource and its reconciliation process.
type Conditioned interface {
	SetConditions(c ...condition.Condition)
	GetCondition(condition.ConditionType) condition.Condition
}

// A ConditionedObject is an Object type with condition field
type ConditionedObject interface {
	client.Object
	Conditioned
}

// PatchCondition will patch status with condition and return, it generally used by cases which don't want to reconcile after patch
func PatchCondition(ctx context.Context, r client.StatusClient, bdcObj ConditionedObject, condition ...condition.Condition) error {
	if len(condition) == 0 {
		return nil
	}
	workloadPatch := client.MergeFrom(bdcObj.DeepCopyObject().(client.Object))
	bdcObj.SetConditions(condition...)
	return r.Status().Patch(ctx, bdcObj, workloadPatch, client.FieldOwner(bdcObj.GetUID()))
}
