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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.Handler = &MutatingHandler{}

// MutatingHandler handles BigDataCluster
type MutatingHandler struct {
	Client client.Client
	// Decoder decodes objects
	Decoder *admission.Decoder
}

var _ inject.Client = &MutatingHandler{}

// InjectClient injects the client into the BigDataClusterValidateHandler
func (h *MutatingHandler) InjectClient(c client.Client) error {
	if h.Client != nil {
		return nil
	}
	h.Client = c
	return nil
}

var _ admission.DecoderInjector = &MutatingHandler{}

// InjectDecoder injects the decoder into the BigDataClusterValidateHandler
func (h *MutatingHandler) InjectDecoder(d *admission.Decoder) error {
	if h.Decoder != nil {
		return nil
	}
	h.Decoder = d
	return nil
}

// Handle validate BigDataCluster Spec here
func (h *MutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.ValidationResponse(true, "")
}

// RegisterMutatingHandler will register BigDataCluster validate handler to the webhook
func RegisterMutatingHandler(mgr manager.Manager) {
	server := mgr.GetWebhookServer()
	server.Register("/mutate-bdc-kdp-io-v1alpha1-bigdatacluster", &webhook.Admission{Handler: &MutatingHandler{}})
}
