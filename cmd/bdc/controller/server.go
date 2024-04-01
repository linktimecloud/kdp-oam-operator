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

package controller

import (
	"context"
	"fmt"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/cmd/bdc/controller/options"
	"kdp-oam-operator/pkg/common"
	"kdp-oam-operator/pkg/controllers/bdc/v1alpha1"
	"kdp-oam-operator/pkg/controllers/configmap"
	"kdp-oam-operator/pkg/webhook/bdc"
	"kdp-oam-operator/version"
	"strconv"
	"strings"

	velav1beta1 "github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(bdcv1alpha1.AddToScheme(scheme))

	utilruntime.Must(velav1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func NewCoreCommand() *cobra.Command {
	s := options.NewCoreOptions()
	cmd := &cobra.Command{
		Use:  "bdc-core",
		Long: `The KDP controller manager is a daemon that embeds the bdc control loops`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(signals.SetupSignalHandler(), s)
		},
		SilenceUsage: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()
	for _, set := range namedFlagSets.FlagSets {
		fs.AddFlagSet(set)
	}

	klog.InfoS("KDP information", "version", version.CoreVersion, "revision", version.GitRevision)
	klog.InfoS("KDP-Core init", "namespace", common.SystemDefaultNamespace)

	return cmd
}

func run(ctx context.Context, s *options.CoreOptions) error {
	restConfig := ctrl.GetConfigOrDie()
	restConfig.UserAgent = common.SystemName + "/" + version.GitRevision
	restConfig.QPS = float32(s.QPS)
	restConfig.Burst = s.Burst
	klog.InfoS("Kubernetes Config Loaded",
		"UserAgent", restConfig.UserAgent,
		"QPS", restConfig.QPS,
		"Burst", restConfig.Burst,
	)
	// go profiling.StartProfilingServer(nil)

	ctrl.SetLogger(klogr.New())

	options.BigDataClusterReSyncPeriod = s.InformerSyncPeriod

	leaderElectionID := GenerateLeaderElectionID(common.SystemName, s.ControllerArgs.IgnoreAppWithoutControllerRequirement)
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:                     scheme,
		MetricsBindAddress:         s.MetricsAddr,
		LeaderElection:             s.EnableLeaderElection,
		LeaderElectionNamespace:    s.LeaderElectionNamespace,
		LeaderElectionID:           leaderElectionID,
		Port:                       s.WebhookPort,
		CertDir:                    s.CertDir,
		HealthProbeBindAddress:     s.HealthAddr,
		LeaderElectionResourceLock: s.LeaderElectionResourceLock,
		LeaseDuration:              &s.LeaseDuration,
		RenewDeadline:              &s.RenewDeadLine,
		RetryPeriod:                &s.RetryPeriod,
		SyncPeriod:                 &s.InformerSyncPeriod,
		// SyncPeriod is configured with default value, aka. 10h. First, controller-runtime does not
		// recommend use it as a time trigger, instead, it is expected to work for failure tolerance
		// of controller-runtime. Additionally, set this value will affect not only application
		// controller but also all other controllers like definition controller. Therefore, for
		// functionalities like state-keep, they should be invented in other ways.
	})
	if err != nil {
		klog.ErrorS(err, "Unable to create a controller manager")
		return err
	}

	if err := registerHealthChecks(mgr); err != nil {
		klog.ErrorS(err, "Unable to register ready/health checks")
		return err
	}

	if s.WebhookEnable {
		klog.InfoS("Enable webhook", "server port", strconv.Itoa(s.WebhookPort))
		bdc.Register(mgr)
	}

	if err := v1alpha1.Setup(mgr, *s.ControllerArgs); err != nil {
		klog.ErrorS(err, fmt.Sprintf("Unable to setup the %s controller", common.BDCControllerName))
		return err
	}

	klog.InfoS("starting configmap syncer to ContextSetting process")
	go func() {
		err = configmap.RunCMSyncer(mgr, s)
	}()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		klog.ErrorS(err, "Failed to run manager")
		return err
	}
	if s.LogFilePath != "" {
		klog.Flush()
	}
	klog.Info("Safely stops Program...")
	return nil
}

// registerHealthChecks is used to create readiness&liveness probes
func registerHealthChecks(mgr ctrl.Manager) error {
	klog.Info("Create readiness/health check")
	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		return err
	}
	// TODO: change the health check to be different from readiness check
	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return err
	}
	return nil
}

func GenerateLeaderElectionID(name string, versionedDeploy bool) string {
	if versionedDeploy {
		return name + "-" + strings.ToLower(strings.ReplaceAll(version.CoreVersion, ".", "-"))
	}
	return name
}
