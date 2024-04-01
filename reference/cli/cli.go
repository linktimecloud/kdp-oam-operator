package cli

import (
	"flag"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"kdp-oam-operator/reference/pkg/utils/common"
	"kdp-oam-operator/reference/pkg/utils/util"
)

var assumeYes bool

// NewCommand will contain all commands
func NewCommand() *cobra.Command {
	return NewCommandWithIOStreams(util.NewDefaultIOStreams())
}

// NewCommandWithIOStreams will contain all commands and initialize them with given ioStream
func NewCommandWithIOStreams(ioStream util.IOStreams) *cobra.Command {
	cmds := &cobra.Command{
		Use:                "bdcctl",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			runHelp(cmd, cmd.Commands(), nil)
		},
		SilenceUsage: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
	}

	scheme := common.Scheme

	commandArgs := common.Args{
		Schema: scheme,
	}

	cmds.AddCommand(
		DefinitionCommandGroup(commandArgs, "2", ioStream),
		ApplicationCommandGroup(commandArgs, "3", ioStream),
		NewHelpCommand("1"),
	)

	fset := flag.NewFlagSet("logs", flag.ContinueOnError)
	klog.InitFlags(fset)
	pfset := pflag.NewFlagSet("logs", pflag.ContinueOnError)
	pfset.AddGoFlagSet(fset)
	pflg := pfset.Lookup("v")
	pflg.Name = "verbosity"
	pflg.Shorthand = "V"

	// init global flags
	cmds.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "Assume yes for all user prompts")
	cmds.PersistentFlags().AddFlag(pflg)
	return cmds
}
