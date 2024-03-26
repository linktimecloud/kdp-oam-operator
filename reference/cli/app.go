package cli

import (
	"context"
	"fmt"
	"kdp-oam-operator/api/bdc/v1alpha1"
	pkgdef "kdp-oam-operator/reference/pkg/definition"
	"kdp-oam-operator/reference/pkg/types"
	"kdp-oam-operator/reference/pkg/utils"
	"kdp-oam-operator/reference/pkg/utils/common"
	"kdp-oam-operator/reference/pkg/utils/util"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

// ApplicationCommandGroup create the command group for `bdcctl app` command to manage applications
func ApplicationCommandGroup(c common.Args, order string, ioStreams util.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage application.",
		Long:  "Manage application.",
		Annotations: map[string]string{
			types.TagCommandOrder: order,
			types.TagCommandType:  types.TypeApp,
		},
	}
	cmd.SetOut(ioStreams.Out)
	cmd.AddCommand(
		NewApplicationApplyCommand(c, ioStreams),
		NewApplicationDeleteCommand(c, ioStreams),
		NewApplicationListCommand(c, ioStreams),
	)
	return cmd
}

// NewApplicationApplyCommand create the `bdcctl app apply` command to help user apply local application to k8s
func NewApplicationApplyCommand(c common.Args, streams util.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply application.yaml",
		Short: "Apply Application.",
		Long:  "Apply Application from local storage to kubernetes cluster. It will apply file to bdc admin-admin and org admin by default.",
		Example: "# Command below will apply the local my-webservice.yaml file to kubernetes, bdc admin-admin\n" +
			"> bdcctl app apply -n webservice my-webservice.yaml\n" +
			"# Command below will apply the ./defs/my-webservice.yaml file to kubernetes bdc test-test\n" +
			"> bdcctl app apply -n webservice -b test-test ./defs/my-webservice.yaml",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			dryRun, err := cmd.Flags().GetBool(FlagDryRun)
			if err != nil {
				return errors.Wrapf(err, "failed to get `%s`", FlagDryRun)
			}
			name, err := cmd.Flags().GetString(FlagName)
			if err != nil {
				return errors.Wrapf(err, "failed to get `%s`", FlagName)
			}
			org, err := cmd.Flags().GetString(FlagOrgName)
			if err != nil {
				return errors.Wrapf(err, "failed to get `%s`", FlagOrgName)
			}
			bdc, err := cmd.Flags().GetString(FlagBdcName)
			if err != nil {
				return errors.Wrapf(err, "failed to get `%s`", FlagBdcName)
			}

			if len(args) < 1 {
				return errors.New("you must specify the application path, directory or URL")
			}
			return applicationApplyAll(ctx, c, streams, name, org, bdc, args[0], dryRun)
		},
	}

	cmd.Flags().StringP(FlagBdcName, "b", types.DefatultBdc, "Specify which bdc the application to deploy.")
	cmd.Flags().StringP(FlagOrgName, "g", types.DefaultOrg, "Specify which org the application belongs to.")
	cmd.Flags().StringP(FlagName, "n", "", "Specify the application resource name.")

	return cmd
}

func applicationApplyAll(ctx context.Context, c common.Args, io util.IOStreams, name string, org string, bdc string, file string, run bool) error {
	stat, err := os.Stat(file)
	if err != nil {
		return errors.Wrapf(err, "failed to get stat for %s, is it exsits?", file)
	}
	if stat.IsDir() || !utils.IsYamlFile(file) {
		return errors.New("only support apply single yaml file")
	}

	files, err := utils.LoadDataFromPath(ctx, file, utils.IsYamlFile)
	if err != nil {
		return err
	}
	result, err := applicationApply(ctx, c, io, name, org, bdc, files[0].Path, files[0].Data, run)
	if err != nil {
		return err
	}
	io.Infonln(result)

	return nil
}

func applicationApply(ctx context.Context, c common.Args, io util.IOStreams, name string, org string, bdc string, defpath string, defBytes []byte, dryRun bool) (string, error) {
	k8sClient, err := c.GetClient()
	if err != nil {
		return "", errors.Wrapf(err, "failed to get k8s client")
	}

	application := pkgdef.Definition{Unstructured: unstructured.Unstructured{}}
	err = yaml.Unmarshal(defBytes, &application)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse yaml for application")
	}

	err = appMetadata(&application, name, org, bdc)
	if err != nil {
		return "", errors.Wrapf(err, "failed to set metadata for application")
	}

	if dryRun {
		data, err := yaml.Marshal(application)
		if err != nil {
			return "", errors.Wrapf(err, "failed to marshal application")
		}
		return string(data), nil
	}

	oldApplication := pkgdef.Definition{Unstructured: unstructured.Unstructured{}}
	oldApplication.SetGroupVersionKind(application.GroupVersionKind())
	err = k8sClient.Get(ctx, k8stypes.NamespacedName{
		Namespace: application.GetNamespace(), Name: application.GetName(),
	}, &oldApplication)
	if err != nil {
		if errors2.IsNotFound(err) {
			if err = k8sClient.Create(ctx, &application); err != nil {
				return "", errors.Wrapf(err, "failed to create new definition in kubernetes")
			}
			return fmt.Sprintf("%s %s created.\n", "Application", application.GetName()), nil
		}
		return "", errors.Wrapf(err, "failed to check existence of target definition in kubernetes")
	}

	mergeApplication(&application, &oldApplication)

	err = k8sClient.Update(ctx, &oldApplication)
	if err != nil {
		return "", errors.Wrapf(err, "failed to update application")
	}

	return fmt.Sprintf("%s %s updated.\n", "Application", application.GetName()), nil
}

func mergeApplication(newApplication *pkgdef.Definition, targetApplication *pkgdef.Definition) {
	targetApplication.SetGVK(newApplication.GetKind())
	targetApplication.SetAnnotations(newApplication.GetAnnotations())
	targetApplication.SetLabels(newApplication.GetLabels())
	targetApplication.Object["spec"] = newApplication.Object["spec"]
}

func appMetadata(application *pkgdef.Definition, name string, org string, bdc string) error {
	// Prefix the application name with bdc
	application.SetName(bdc + "-" + name)

	annotations := application.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// Set Bdc and Org annotations
	annotations[types.BdcKey] = bdc
	annotations[types.BdcOrg] = org
	application.SetAnnotations(annotations)

	labels := application.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	// Set Bdc and Org labels
	labels[types.BdcKey] = bdc
	labels[types.BdcOrg] = org
	application.SetLabels(labels)

	spec, ok := application.Object["spec"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unable to cast spec to map[string]interface{}")
	}

	// Set name in spec
	spec["name"] = name

	return nil
}

// NewApplicationDeleteCommand create the `bdcctl app delete` command to help user delete application in k8s
func NewApplicationDeleteCommand(c common.Args, streams util.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete webservice",
		Short: "delete Application.",
		Long:  "delete Application from kubernetes cluster. It will delete application on kubernetes.",
		Example: "# Command below will delete webservice in kubernetes\n" +
			"> bdcctl app delete webservice\n" +
			"# Command below will delete webservice1 webservice2 in kubernetes\n" +
			"> bdcctl app delete webservice1 webservice2\n",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) < 1 {
				return errors.New("you must specify the application name or name list")
			}
			return applicationDeleteAll(ctx, c, streams, args)
		},
	}

	return cmd
}

func applicationDeleteAll(ctx context.Context, c common.Args, streams util.IOStreams, args []string) error {
	for _, name := range args {
		err := applicationDeleteOne(ctx, c, streams, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func applicationDeleteOne(ctx context.Context, c common.Args, streams util.IOStreams, name string) error {
	k8sClient, err := c.GetClient()
	if err != nil {
		return errors.Wrapf(err, "failed to get k8s client")
	}

	application := pkgdef.Definition{Unstructured: unstructured.Unstructured{}}
	application.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   v1alpha1.GroupVersion.Group,
		Version: v1alpha1.GroupVersion.Version,
		Kind:    "Application",
	})
	application.SetName(name)
	err = k8sClient.Delete(ctx, &application)
	if err != nil {
		return errors.Wrapf(err, "failed to delete application %s", name)
	}

	streams.Info(fmt.Sprintf("Application %s deleted", name))
	return nil
}

// NewApplicationListCommand create the `bdcctl app list` command to help user delete application in k8s
func NewApplicationListCommand(c common.Args, streams util.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list Application.",
		Long:  "list Application from kubernetes cluster.",
		Example: "# Command below will list webservice in kubernetes\n" +
			"> bdcctl app list\n",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			return applicationListAll(ctx, c, streams, args)
		},
	}

	return cmd
}

func applicationListAll(ctx context.Context, c common.Args, streams util.IOStreams, args []string) error {
	k8sClient, err := c.GetClient()
	if err != nil {
		return errors.Wrapf(err, "failed to get k8s client")
	}
	applicationList := unstructured.UnstructuredList{}
	applicationList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   v1alpha1.GroupVersion.Group,
		Version: v1alpha1.GroupVersion.Version,
		Kind:    "ApplicationList",
	})
	if err := k8sClient.List(ctx, &applicationList); err != nil {
		if meta.IsNoMatchError(err) {

		} else {
			return errors.Wrapf(err, "failed to list application")
		}
	}

	for _, application := range applicationList.Items {
		streams.Info(application.GetName())
	}

	return nil
}
