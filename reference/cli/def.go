/*
Copyright 2021 The Kubebdcctl Authors.

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
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8stypes "k8s.io/apimachinery/pkg/types"
	pkgdef "kdp-oam-operator/reference/pkg/definition"
	"kdp-oam-operator/reference/pkg/types"
	"kdp-oam-operator/reference/pkg/utils"
	"kdp-oam-operator/reference/pkg/utils/common"
	"kdp-oam-operator/reference/pkg/utils/util"
)

// DefinitionCommandGroup create the command group for `bdcctl def` command to manage definitions
func DefinitionCommandGroup(c common.Args, order string, ioStreams util.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "def",
		Short: "Manage definitions.",
		Long:  "Manage X-Definitions for extension.",
		Annotations: map[string]string{
			types.TagCommandOrder: order,
			types.TagCommandType:  types.TypeExtension,
		},
	}
	cmd.SetOut(ioStreams.Out)
	cmd.AddCommand(
		NewDefinitionApplyCommand(c, ioStreams),
	)
	return cmd
}

// NewDefinitionApplyCommand create the `bdcctl def apply` command to help user apply local definitions to k8s
func NewDefinitionApplyCommand(c common.Args, streams util.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply DEFINITION.cue",
		Short: "Apply X-Definition.",
		Long:  "Apply X-Definition from local storage to kubernetes cluster. ",
		Example: "# Command below will apply the local my-webservice.cue file to kubernetes\n" +
			"> bdcctl def apply my-webservice.cue\n" +
			"# Apply the local directory including all files(CUE definition) to kubernetes\n" +
			"> bdcctl def apply def/\n",
		// todo: implement dry-run function
		// +
		//"# Command below will convert the ./defs/my-trait.cue file to kubernetes CRD object and print it without applying it to kubernetes\n" +
		//"> bdcctl def apply ./defs/my-trait.cue --dry-run （to do）",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			//dryRun, err := cmd.Flags().GetBool(FlagDryRun)
			//if err != nil {
			//	return errors.Wrapf(err, "failed to get `%s`", FlagDryRun)
			//}
			if len(args) < 1 {
				return errors.New("you must specify the definition path, directory or URL")
			}
			return defApplyAll(ctx, c, streams, args[0], false)
		},
	}

	//cmd.Flags().BoolP(FlagDryRun, "", false, "only build definition from CUE into CRB object without applying it to kubernetes clusters")
	return cmd
}

func defApplyAll(ctx context.Context, c common.Args, io util.IOStreams, path string, dryRun bool) error {
	files, err := utils.LoadDataFromPath(ctx, path, utils.IsCUEFile)
	if err != nil {
		return errors.Wrapf(err, "failed to get from %s", path)
	}
	for _, f := range files {
		result, err := defApplyOne(ctx, c, f.Path, f.Data, dryRun)
		if err != nil {
			return err
		}
		io.Infonln(result)
	}
	return nil
}

func defApplyOne(ctx context.Context, c common.Args, defpath string, defBytes []byte, dryRun bool) (string, error) {
	_, err := c.GetConfig()
	if err != nil {
		return "", err
	}
	k8sClient, err := c.GetClient()
	if err != nil {
		return "", errors.Wrapf(err, "failed to get k8s client")
	}

	def := pkgdef.Definition{Unstructured: unstructured.Unstructured{}}

	if err := def.FromCUEString(string(defBytes)); err != nil {
		return "", errors.Wrapf(err, "failed to parse CUE for definition")
	}

	oldDef := pkgdef.Definition{Unstructured: unstructured.Unstructured{}}
	oldDef.SetGroupVersionKind(def.GroupVersionKind())
	err = k8sClient.Get(ctx, k8stypes.NamespacedName{
		Namespace: def.GetNamespace(),
		Name:      def.GetName(),
	}, &oldDef)
	if err != nil {
		if errors2.IsNotFound(err) {
			kind := def.GetKind()
			if err = k8sClient.Create(ctx, &def); err != nil {
				return "", errors.Wrapf(err, "failed to create new definition in kubernetes")
			}
			return fmt.Sprintf("%s %s created.\n", kind, def.GetName()), nil
		}
		return "", errors.Wrapf(err, "failed to check existence of target definition in kubernetes")
	}
	if err := oldDef.FromCUEString(string(defBytes)); err != nil {
		return "", errors.Wrapf(err, "failed to merge with existing definition")
	}
	if err = k8sClient.Update(ctx, &oldDef); err != nil {
		return "", errors.Wrapf(err, "failed to update existing definition in kubernetes")
	}
	return fmt.Sprintf("%s %s updated.\n", oldDef.GetKind(), oldDef.GetName()), nil
}
