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

package cli

import (
	"github.com/spf13/cobra"
)

// constants used in `svc` command
const (
	App       = "app"
	Namespace = "namespace"
	// FlagDryRun command flag to disable actual changes and only display intend changes
	FlagDryRun = "dry-run"
	// FlagName command flag to specify the name of the resource
	FlagName = "name"
	// FlagNamespace command flag to specify which namespace to use
	FlagNamespace = "namespace"

	// FlagBdcName command flag to specify which big data cluster to use
	FlagBdcName = "bdc"
	// FlagOrgName command flag to specify which organization to use
	FlagOrgName = "org"
)

func addNamespaceAndEnvArg(cmd *cobra.Command) {
	cmd.Flags().StringP(Namespace, "n", "", "specify the Kubernetes namespace to use")

	cmd.PersistentFlags().StringP("env", "e", "", "specify environment name for application")
}
