/*
Copyright 2024 KDP(Kubernetes Data Platform).

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

package config

import (
	"kdp-oam-operator/cmd/apiserver/options"
	"time"
)

// APIServerConfig config for server
type APIServerConfig struct {
	// api server bind address
	BindAddr string
	// monitor metric path
	MetricPath string
	// swagger doc enabled
	SwaggerDocEnabled bool
	// generic options
	GenericOptions options.GenericOptions
	// KubeBurst the burst of kube client
	KubeBurst int
	// KubeQPS the QPS of kube client
	KubeQPS float64
	// LeaderConfig for leader election
	LeaderConfig leaderConfig
	// DefaultSystemNS
	DefaultSystemNS string
}

type leaderConfig struct {
	ID       string
	LockName string
	Duration time.Duration
}
