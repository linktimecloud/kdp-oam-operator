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

package service

import (
	"context"
	"fmt"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"k8s.io/client-go/kubernetes/scheme"
	//+kubebuilder:scaffold:imports
)

var kubeConfig *rest.Config
var kubeClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Suite")
}

// claim all services, ds, kubeClient
var (
	ctx                   context.Context
	bigDataClusterService *bigDataClusterServiceImpl
	appService            *applicationServiceImpl
	contextSecretService  *contextSecretServiceImpl
	contextSettingService *contextSettingServiceImpl
	defService            *xDefinitionServiceImpl
)

func InitAllServices() {

	bigDataClusterService = &bigDataClusterServiceImpl{kubeClient, kubeConfig}
	appService = &applicationServiceImpl{kubeClient, kubeConfig}
	contextSecretService = &contextSecretServiceImpl{kubeClient, kubeConfig}
	contextSettingService = &contextSettingServiceImpl{kubeClient, kubeConfig}
	defService = &xDefinitionServiceImpl{kubeClient, kubeConfig}

}

func InitTestEnv() {
	InitAllServices()
	ctx = context.Background()
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		ControlPlaneStartTimeout: time.Minute,
		ControlPlaneStopTimeout:  time.Minute,
		CRDDirectoryPaths:        []string{filepath.Join("../../../..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing:    true,
	}
	By(fmt.Sprintf("%s", []string{filepath.Join("../../../..", "config", "crd", "bases")}))

	var err error
	// cfg is defined in this file globally.
	kubeConfig, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(kubeConfig).NotTo(BeNil())

	err = bdcv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	kubeClient, err = client.New(kubeConfig, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(kubeClient).NotTo(BeNil())
	By("new kube client success")
	clients.SetKubeClient(kubeClient)
	Expect(err).Should(BeNil())

	initDefinitions(kubeClient)

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func initDefinitions(kubeClient client.Client) {
	By("init definitions")
	applicationDef, err := os.ReadFile("./testdata/application-def.yaml")
	fmt.Printf("applicationDef: %+v", string(applicationDef))
	Expect(err).Should(BeNil())
	var def bdcv1alpha1.XDefinition
	err = yaml.Unmarshal(applicationDef, &def)
	Expect(err).Should(BeNil())
	fmt.Printf("def: %+v", def)
	Expect(kubeClient.Create(context.TODO(), &def))

}
