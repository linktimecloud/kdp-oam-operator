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

package e2etest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/cmd/apiserver/options"
	"kdp-oam-operator/pkg/apiserver"
	"kdp-oam-operator/pkg/apiserver/config"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"net/http"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
	"testing"
	"time"
)

var kubeConfig *rest.Config
var kubeClient client.Client
var testEnv *envtest.Environment
var ctx context.Context

const (
	baseDomain = "http://127.0.0.1:8888"
	baseURL    = "http://127.0.0.1:8888/api/v1"
)

func TestE2EAPIServerTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E APIServerTest Suite")
}

// Suite test in e2e-apiServer-test relies on the pre-setup kubernetes environment
var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		ControlPlaneStartTimeout: time.Minute,
		ControlPlaneStopTimeout:  time.Minute,
		CRDDirectoryPaths:        []string{filepath.Join("../../..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing:    true,
	}
	By(fmt.Sprintf("%s", []string{filepath.Join("../../..", "config", "crd", "bases")}))

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

	cfg := config.APIServerConfig{
		BindAddr:          "127.0.0.1:8888",
		SwaggerDocEnabled: false,
		KubeQPS:           100,
		KubeBurst:         300,
		GenericOptions: options.GenericOptions{
			LogLevel: "info",
		},
	}
	server, err := apiserver.New(cfg)
	Expect(server).ShouldNot(BeNil())

	go func() {
		defer GinkgoRecover()
		err := server.Run(ctx)
		fmt.Print(server)
		Expect(err).ShouldNot(HaveOccurred())
	}()
	By("wait for api server to start")

	Eventually(
		func() error {
			readyzReq, err := http.NewRequest("GET", baseDomain+"/readyz", nil)
			Expect(err).Should(BeNil())
			readyzReq.Header.Set("Content-Type", "application/json")
			res, err := http.DefaultClient.Do(readyzReq)
			if err != nil {
				return err
			}
			if res.StatusCode != 200 {
				return fmt.Errorf("check readyz failed: %v", res)
			}
			return nil
		}).WithTimeout(time.Second * 60).WithPolling(time.Millisecond * 200).Should(Succeed())
	Expect(err).ShouldNot(HaveOccurred())
	By("api server started")

	// Init BigDataCluster Instance
	By("Init bigdata-cluster")
	prepareBigDataCluster()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func postRequest(path string, body interface{}) *http.Response {
	b, err := json.Marshal(body)
	Expect(err).Should(BeNil())
	httpClient := &http.Client{}
	if !strings.HasPrefix(path, "/v1") {
		path = baseURL + path
	} else {
		path = baseDomain + path
	}
	req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	Expect(err).Should(BeNil())
	req.Header.Add("Content-Type", "application/json")

	response, err := httpClient.Do(req)
	Expect(err).Should(BeNil())
	return response
}

func putRequest(path string, body interface{}) *http.Response {
	b, err := json.Marshal(body)
	Expect(err).Should(BeNil())
	httpClient := &http.Client{}
	if !strings.HasPrefix(path, "/v1") {
		path = baseURL + path
	} else {
		path = baseDomain + path
	}
	req, err := http.NewRequest(http.MethodPut, path, bytes.NewBuffer(b))
	Expect(err).Should(BeNil())
	req.Header.Set("Content-Type", "application/json")

	response, err := httpClient.Do(req)
	Expect(err).Should(BeNil())
	return response
}

func getRequest(path string) *http.Response {
	httpClient := &http.Client{}
	if !strings.HasPrefix(path, "http") {
		if !strings.HasPrefix(path, "/v1") {
			path = baseURL + path
		} else {
			path = baseDomain + path
		}
	}
	req, err := http.NewRequest(http.MethodGet, path, nil)
	Expect(err).Should(BeNil())

	response, err := httpClient.Do(req)
	Expect(err).Should(BeNil())
	return response
}

func getWithQueryRequest(path string, params map[string]string) *http.Response {
	httpClient := &http.Client{}
	if !strings.HasPrefix(path, "/v1") {
		path = baseURL + path
	} else {
		path = baseDomain + path
	}
	req, err := http.NewRequest(http.MethodGet, path, nil)
	Expect(err).Should(BeNil())
	query := req.URL.Query()
	for k, v := range params {
		query.Set(k, v)
	}
	req.URL.RawQuery = query.Encode()
	response, err := httpClient.Do(req)
	Expect(err).Should(BeNil())
	return response
}

func deleteRequest(path string) *http.Response {
	httpClient := &http.Client{}
	if !strings.HasPrefix(path, "/v1") {
		path = baseURL + path
	} else {
		path = baseDomain + path
	}
	req, err := http.NewRequest(http.MethodDelete, path, nil)
	Expect(err).Should(BeNil())
	response, err := httpClient.Do(req)
	Expect(err).Should(BeNil())
	return response
}

func decodeResponseBody(resp *http.Response, dst interface{}) error {
	if resp.Body == nil {
		return fmt.Errorf("response body is nil")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if dst != nil {
		err = json.Unmarshal(body, dst)
		if err != nil {
			return err
		}
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("response code is not 200: %d body: %s", resp.StatusCode, string(body))
	}
	return nil
}
