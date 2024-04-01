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

package apply

import (
	"context"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Applicator interface {
	Apply(context.Context, client.Object, ...ApplyOption) error
}

type applyAction struct {
	skipUpdate       bool
	updateAnnotation bool
}

// ApplyOption is called before applying state to the object.
// ApplyOption is still called even if the object does NOT exist.
// If the object does not exist, `existing` will be assigned as `nil`.
// nolint: golint
type ApplyOption func(act *applyAction, existing, desired client.Object) error

// NewAPIApplicator creates an Applicator that applies state to an
// object or creates the object if not exist.
func NewAPIApplicator(c client.Client) *APIApplicator {
	return &APIApplicator{
		creator: creatorFn(createOrGetExisting),
		patcher: patcherFn(threeWayMergePatch),
		c:       c,
	}
}

type creator interface {
	createOrGetExisting(context.Context, *applyAction, client.Client, client.Object, ...ApplyOption) (client.Object, error)
}

type creatorFn func(context.Context, *applyAction, client.Client, client.Object, ...ApplyOption) (client.Object, error)

func (fn creatorFn) createOrGetExisting(ctx context.Context, act *applyAction, c client.Client, o client.Object, ao ...ApplyOption) (client.Object, error) {
	return fn(ctx, act, c, o, ao...)
}

type patcher interface {
	patch(c, m client.Object, a *applyAction) (client.Patch, error)
}

type patcherFn func(c, m client.Object, a *applyAction) (client.Patch, error)

func (fn patcherFn) patch(c, m client.Object, a *applyAction) (client.Patch, error) {
	return fn(c, m, a)
}

// APIApplicator implements Applicator
type APIApplicator struct {
	creator
	patcher
	c client.Client
}

// loggingApply will record a log with desired object applied
func loggingApply(msg string, desired client.Object) {
	d, ok := desired.(metav1.Object)
	if !ok {
		klog.InfoS(msg, "resource", desired.GetObjectKind().GroupVersionKind().String())
		return
	}
	klog.InfoS(msg, "name", d.GetName(), "resource", desired.GetObjectKind().GroupVersionKind().String())
}

// Apply applies new state to an object or create it if not exist
func (a *APIApplicator) Apply(ctx context.Context, desired client.Object, ao ...ApplyOption) error {
	applyAct := &applyAction{updateAnnotation: excludeLastAppliedConfigurationForSpecialResources(desired)}
	existing, err := a.createOrGetExisting(ctx, applyAct, a.c, desired, ao...)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}

	// the object already exists, apply new state
	if err := executeApplyOptions(applyAct, existing, desired, ao); err != nil {
		return err
	}

	if applyAct.skipUpdate {
		loggingApply("skip update", desired)
		return nil
	}

	loggingApply("patching object", desired)
	patch, err := a.patcher.patch(existing, desired, applyAct)
	if err != nil {
		return errors.Wrap(err, "cannot calculate patch by computing a three way diff")
	}
	return errors.Wrapf(a.c.Patch(ctx, desired, patch), "cannot patch object")
}

// excludeLastAppliedConfigurationForSpecialResources will filter special object that can reduce the record for "bdc.kdp.io/last-applied-configuration" annotation.
func excludeLastAppliedConfigurationForSpecialResources(desired client.Object) bool {
	if desired == nil {
		return false
	}
	gvk := desired.GetObjectKind().GroupVersionKind()
	gp, kd := gvk.Group, gvk.Kind
	if gp == "" {
		// group is empty means it's Kubernetes core API, we won't record annotation for Secret and Configmap
		if kd == "Secret" || kd == "ConfigMap" || kd == "CustomResourceDefinition" {
			return false
		}
		if _, ok := desired.(*corev1.ConfigMap); ok {
			return false
		}
		if _, ok := desired.(*corev1.Secret); ok {
			return false
		}
		if _, ok := desired.(*v1.CustomResourceDefinition); ok {
			return false
		}
	}

	if gp == "core.oam.dev" {
		if kd == "Application" {
			return false
		}
	}

	ann := desired.GetAnnotations()
	if ann != nil {
		lac := ann[constants.AnnotationLastAppliedConfig]
		if lac == "-" || lac == "skip" {
			return false
		}
	}
	return true
}

// createOrGetExisting will create the object if it does not exist
// or get and return the existing object
func createOrGetExisting(ctx context.Context, act *applyAction, c client.Client, desired client.Object, ao ...ApplyOption) (client.Object, error) {
	var create = func() (client.Object, error) {
		// execute ApplyOptions even the object doesn't exist
		if err := executeApplyOptions(act, nil, desired, ao); err != nil {
			return nil, err
		}
		loggingApply("creating object", desired)
		return nil, errors.Wrap(c.Create(ctx, desired), "cannot create object")
	}

	if act.updateAnnotation {
		if err := addLastAppliedConfigAnnotation(desired); err != nil {
			return nil, err
		}
	}

	// allow to create object with only generateName
	if desired.GetName() == "" && desired.GetGenerateName() != "" {
		return create()
	}

	existing := &unstructured.Unstructured{}
	existing.GetObjectKind().SetGroupVersionKind(desired.GetObjectKind().GroupVersionKind())
	err := c.Get(ctx, types.NamespacedName{Name: desired.GetName(), Namespace: desired.GetNamespace()}, existing)
	if kerrors.IsNotFound(err) {
		return create()
	}
	if err != nil {
		return nil, errors.Wrap(err, "cannot get object")
	}
	return existing, nil
}

func executeApplyOptions(act *applyAction, existing, desired client.Object, aos []ApplyOption) error {
	// if existing is nil, it means the object is going to be created.
	// ApplyOption function should handle this situation carefully by itself.
	for _, fn := range aos {
		if err := fn(act, existing, desired); err != nil {
			return errors.Wrap(err, "cannot apply ApplyOption")
		}
	}
	return nil
}
