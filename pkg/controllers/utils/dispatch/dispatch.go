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

package dispatch

import (
	"context"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"kdp-oam-operator/pkg/controllers/utils/apply"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ManifestsDispatcher dispatch manifests into K8s
type ManifestsDispatcher struct {
	C          client.Client
	Applicator apply.Applicator
}

// NewManifestsDispatcher creates an ManifestsDispatcher.
func NewManifestsDispatcher(c client.Client) *ManifestsDispatcher {
	return &ManifestsDispatcher{
		C:          c,
		Applicator: apply.NewAPIApplicator(c),
	}
}

// Dispatch apply manifests into k8s
func (a *ManifestsDispatcher) Dispatch(ctx context.Context, manifests ...*unstructured.Unstructured) error {
	var applyOpts []apply.ApplyOption
	for _, rsc := range manifests {
		if rsc == nil {
			continue
		}
		// each resource applied by dispatcher MUST be controlled by resource tracker
		if err := a.Applicator.Apply(ctx, rsc, applyOpts...); err != nil {
			klog.ErrorS(err, "Failed to apply a resource", "object",
				klog.KObj(rsc), "apiVersion", rsc.GetAPIVersion(), "kind", rsc.GetKind())
			return errors.Wrapf(err, "cannot apply manifest, name: %q apiVersion: %q kind: %q",
				rsc.GetName(), rsc.GetAPIVersion(), rsc.GetKind())
		}
		klog.InfoS("Successfully apply a resource", "object",
			klog.KObj(rsc), "apiVersion", rsc.GetAPIVersion(), "kind", rsc.GetKind())
	}
	return nil
}
