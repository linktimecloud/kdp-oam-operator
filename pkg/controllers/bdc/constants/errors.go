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
	ErrUpdateCapabilityInConfigMap                    = "cannot create or update capability %s in ConfigMap: %v"
	ErrUpdateDefinitionAndAPIResourceMappingConfigMap = "cannot create or update definition and apiresource map %s in ConfigMap: %v"
	ErrCreateBDCResource                              = "cannot create %s of %s: %v"
	ErrGenerateOpenAPIV3JSONSchemaForCapability       = "cannot generate OpenAPI v3 JSON schema for capability %s: %v"
	ErrGetBdc                                         = "cannot get bdc %s: %v"
	ErrGetBdbFile                                     = "cannot generate XDefinition File: %v"
	ErrGenerateManifests                              = "cannot generate manifest: %v"
	ErrGetVelaApplication                             = "cannot get vela application: %v"
)
