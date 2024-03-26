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

package types

import (
	"cuelang.org/go/cue"
)

// Source record the source of Capability
type Source struct {
	RepoName  string `json:"repoName"`
	ChartName string `json:"chartName,omitempty"`
}

// CRDInfo record the CRD info of the Capability
type CRDInfo struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

// CapType defines the type of capability
type CapType string

// CapabilityCategory defines the category of a capability
type CapabilityCategory string

// Parameter defines a parameter for cli from capability template
type Parameter struct {
	Name     string      `json:"name"`
	Short    string      `json:"short,omitempty"`
	Required bool        `json:"required,omitempty"`
	Default  interface{} `json:"default,omitempty"`
	Usage    string      `json:"usage,omitempty"`
	Ignore   bool        `json:"ignore,omitempty"`
	Type     cue.Kind    `json:"type,omitempty"`
	Alias    string      `json:"alias,omitempty"`
	JSONType string      `json:"jsonType,omitempty"`
}

// Capability defines the content of a capability
type Capability struct {
	Name           string             `json:"name"`
	Type           CapType            `json:"type"`
	CueTemplate    string             `json:"template,omitempty"`
	CueTemplateURI string             `json:"templateURI,omitempty"`
	Parameters     []Parameter        `json:"parameters,omitempty"`
	CrdName        string             `json:"crdName,omitempty"`
	Status         string             `json:"status,omitempty"`
	Description    string             `json:"description,omitempty"`
	Example        string             `json:"example,omitempty"`
	Labels         map[string]string  `json:"labels,omitempty"`
	Category       CapabilityCategory `json:"category,omitempty"`

	// trait only
	AppliesTo []string `json:"appliesTo,omitempty"`

	// Namespace represents it's a system-level or user-level capability.
	Namespace string `json:"namespace,omitempty"`

	// Plugin Source
	Source *Source `json:"source,omitempty"`

	// Terraform
	TerraformConfiguration string `json:"terraformConfiguration,omitempty"`
	ConfigurationType      string `json:"configurationType,omitempty"`
	Path                   string `json:"path,omitempty"`
}
