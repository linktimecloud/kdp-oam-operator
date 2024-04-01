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

// the names for different type of definition
const (
	applicationXDefPrefix    = "application-"
	contextSettingXDefPrefix = "ctx-setting-"
)

var (
	ApiResourceTypePrefix = map[string]string{
		"Application":    applicationXDefPrefix,
		"ContextSetting": contextSettingXDefPrefix,
	}
)

// DefaultNS defines the default kdp namespace in Kubernetes
var DefaultNS = "admin"

// DefaultBdc defines the default big data cluster in Kubernetes
var DefatultBdc = "admin-admin"

// DefaultOrg defines the default organization in Kubernetes
var DefaultOrg = "admin"

// Config contains key/value pairs
type Config map[string]string

// EnvMeta stores the namespace for app environment
type EnvMeta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Labels    string `json:"labels"`
	Current   string `json:"current"`
}

const (
	// TagCommandType used for tag cli category
	TagCommandType = "commandType"

	// TagCommandOrder defines the order
	TagCommandOrder = "commandOrder"

	// TypeStart defines one category
	TypeStart = "Getting Started"

	// TypeApp defines one category
	TypeApp = "Managing Applications"

	// TypeCD defines workflow Management operations
	TypeCD = "Continuous Delivery"

	// TypeExtension defines one category
	TypeExtension = "Managing Extensions"

	// TypeSystem defines one category
	TypeSystem = "System Tools"

	// TypeAuxiliary defines auxiliary commands
	TypeAuxiliary = "Auxiliary Tools"

	// TypePlatform defines platform management commands
	TypePlatform = "Managing Platform"
)
const (
	BdcKey = "bdc.kdp.io/name"
	BdcOrg = "bdc.kdp.io/org"
)
