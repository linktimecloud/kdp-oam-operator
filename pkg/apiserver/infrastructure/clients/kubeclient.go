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

package clients

import (
	apiserverCfg "kdp-oam-operator/pkg/apiserver/config"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var kubeClient client.Client
var kubeConfig *rest.Config
var Scheme = k8sruntime.NewScheme()

// SetKubeClient for test
func SetKubeClient(c client.Client) {
	kubeClient = c
}

func setKubeConfig(conf *rest.Config) (err error) {
	if conf == nil {
		conf, err = config.GetConfig()
		if err != nil {
			return err
		}
	}
	kubeConfig = conf
	return nil
}

// SetKubeConfig generate the kube config from the config of apiserver
func SetKubeConfig(c apiserverCfg.APIServerConfig) error {
	conf, err := config.GetConfig()
	if err != nil {
		return err
	}
	kubeConfig = conf
	kubeConfig.Burst = c.KubeBurst
	kubeConfig.QPS = float32(c.KubeQPS)
	return setKubeConfig(kubeConfig)
}

// GetKubeClient create and return kube runtime client
func GetKubeClient() (client.Client, error) {
	if kubeClient != nil {
		return kubeClient, nil
	}
	// create single cluster client
	conf, err := GetKubeConfig()
	if err != nil {
		return nil, err
	}
	kubeClient, err = client.New(conf, client.Options{Scheme: Scheme})
	if err != nil {
		return nil, err
	}
	return kubeClient, nil
}

// GetKubeConfig create/get kube runtime config
func GetKubeConfig() (*rest.Config, error) {
	var err error
	if kubeConfig == nil {
		kubeConfig, err = config.GetConfig()
		return kubeConfig, err
	}
	return kubeConfig, nil
}
