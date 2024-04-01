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

package bigdatacluster

import (
	"context"
	_ "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	bdc "kdp-oam-operator/api/bdc/v1alpha1"
)

// ValidateCreate validates the BigDataCluster on creation
func (h *ValidatingHandler) ValidateCreate(ctx context.Context, bdc *bdc.BigDataCluster) field.ErrorList {
	var allErrs field.ErrorList

	return allErrs
}

// ValidateUpdate validates the BigDataCluster on update
func (h *ValidatingHandler) ValidateUpdate(ctx context.Context, newBdc, oldBDC *bdc.BigDataCluster) field.ErrorList {
	// check if the newBdc is valid
	errs := h.ValidateCreate(ctx, newBdc)
	// TODO: add more validating
	return errs
}
