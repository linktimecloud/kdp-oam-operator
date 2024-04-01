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

package options

import (
	pkgcommon "kdp-oam-operator/pkg/common"
	bdcctrl "kdp-oam-operator/pkg/controllers/bdc"
	"strconv"
	"time"

	cliflag "k8s.io/component-base/cli/flag"
)

var (
	// PerfEnabled identify whether to add performance log for controllers
	PerfEnabled = false
	// BigDataClusterReSyncPeriod re-sync period to reconcile application
	BigDataClusterReSyncPeriod = time.Minute * 5
)

const (
	LogDebug = 1
)

type CoreOptions struct {
	WebhookEnable              bool
	CertDir                    string
	WebhookPort                int
	MetricsAddr                string
	EnableLeaderElection       bool
	LeaderElectionNamespace    string
	LogFilePath                string
	LogFileMaxSize             uint64
	LogDebug                   bool
	ControllerArgs             *bdcctrl.Args
	HealthAddr                 string
	StorageDriver              string
	InformerSyncPeriod         time.Duration
	QPS                        float64
	Burst                      int
	LeaderElectionResourceLock string
	LeaseDuration              time.Duration
	RenewDeadLine              time.Duration
	RetryPeriod                time.Duration
}

// NewCoreOptions creates a new NewCoreOptions object with default parameters
func NewCoreOptions() *CoreOptions {
	s := &CoreOptions{
		WebhookEnable:           false,
		CertDir:                 "/k8s-webhook-server/serving-certs",
		WebhookPort:             9443,
		MetricsAddr:             ":8080",
		EnableLeaderElection:    false,
		LeaderElectionNamespace: "",
		LogFilePath:             "",
		LogFileMaxSize:          1024,
		LogDebug:                false,
		ControllerArgs: &bdcctrl.Args{
			RevisionLimit:                                50,
			AppRevisionLimit:                             10,
			DefRevisionLimit:                             20,
			AutoGenWorkloadDefinition:                    true,
			ConcurrentReconciles:                         4,
			IgnoreAppWithoutControllerRequirement:        false,
			IgnoreDefinitionWithoutControllerRequirement: false,
		},
		HealthAddr:                 ":9440",
		StorageDriver:              "Local",
		InformerSyncPeriod:         1 * time.Minute,
		QPS:                        50,
		Burst:                      100,
		LeaderElectionResourceLock: "configmapsleases",
		LeaseDuration:              15 * time.Second,
		RenewDeadLine:              10 * time.Second,
		RetryPeriod:                2 * time.Second,
	}
	return s
}

// Flags returns the complete NamedFlagSets
func (s *CoreOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}

	gfs := fss.FlagSet("generic")
	gfs.BoolVar(&s.WebhookEnable, "webhook-enable", s.WebhookEnable, "Enable Admission Webhook")
	gfs.StringVar(&s.CertDir, "webhook-cert-dir", s.CertDir, "Admission webhook cert/key dir.")
	gfs.IntVar(&s.WebhookPort, "webhook-port", s.WebhookPort, "admission webhook listen address")
	gfs.StringVar(&s.MetricsAddr, "metrics-addr", s.MetricsAddr, "The address the metric endpoint binds to.")
	gfs.BoolVar(&s.EnableLeaderElection, "enable-leader-election", s.EnableLeaderElection, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	gfs.StringVar(&s.LeaderElectionNamespace, "leader-election-namespace", s.LeaderElectionNamespace, "Determines the namespace in which the leader election configmap will be created.")
	gfs.StringVar(&s.LogFilePath, "log-file-path", s.LogFilePath, "The file to write logs to.")
	gfs.Uint64Var(&s.LogFileMaxSize, "log-file-max-size", s.LogFileMaxSize, "Defines the maximum size a log file can grow to, Unit is megabytes.")
	gfs.BoolVar(&s.LogDebug, "log-debug", s.LogDebug, "Enable debug logs for development purpose")
	gfs.StringVar(&s.HealthAddr, "health-addr", s.HealthAddr, "The address the health endpoint binds to.")
	gfs.DurationVar(&s.InformerSyncPeriod, "informer-sync-period", s.InformerSyncPeriod, "The re-sync period for informer in controller-runtime. This is a system-level configuration.")
	gfs.Float64Var(&s.QPS, "kube-api-qps", s.QPS, "the qps for reconcile clients. Low qps may lead to low throughput. High qps may give stress to api-server. Raise this value if concurrent-reconciles is set to be high.")
	gfs.IntVar(&s.Burst, "kube-api-burst", s.Burst, "the burst for reconcile clients. Recommend setting it qps*2.")
	gfs.StringVar(&s.LeaderElectionResourceLock, "leader-election-resource-lock", s.LeaderElectionResourceLock, "The resource lock to use for leader election")
	gfs.DurationVar(&s.LeaseDuration, "leader-election-lease-duration", s.LeaseDuration, "The duration that non-leader candidates will wait to force acquire leadership")
	gfs.DurationVar(&s.RenewDeadLine, "leader-election-renew-deadline", s.RenewDeadLine, "The duration that the acting controlplane will retry refreshing leadership before giving up")
	gfs.DurationVar(&s.RetryPeriod, "leader-election-retry-period", s.RetryPeriod, "The duration the LeaderElector clients should wait between tries of actions")

	s.ControllerArgs.AddFlags(fss.FlagSet("controllerArgs"), s.ControllerArgs)

	cfs := fss.FlagSet("common_config")
	cfs.DurationVar(&BigDataClusterReSyncPeriod, "bigdata-cluster-re-sync-period", BigDataClusterReSyncPeriod,
		"Re-sync period for bigdata-cluster to re-sync, also known as the state-keep interval.")
	cfs.BoolVar(&PerfEnabled, "perf-enabled", PerfEnabled, "Enable performance logging for controllers, disabled by default.")
	cfs.StringVar(&pkgcommon.SystemDefaultNamespace, "system-default-namespace", "kdp-system", "define the namespace of the system-level definition")
	cfs.StringVar(&pkgcommon.KdpContextLabelKey, "kdp-context-label-key", "kdp-operator-context", "define the label key of context cm")
	cfs.StringVar(&pkgcommon.KdpContextLabelValue, "kdp-context-label-value", "KDP", "define the label value of context cm")

	kfs := fss.FlagSet("klog")
	if s.LogDebug {
		_ = kfs.Set("v", strconv.Itoa(int(LogDebug)))
	}

	if s.LogFilePath != "" {
		_ = kfs.Set("log_to_stderr", "false")
		_ = kfs.Set("log_file", s.LogFilePath)
		_ = kfs.Set("log_file_max_size", strconv.FormatUint(s.LogFileMaxSize, 10))
	}

	return fss
}
