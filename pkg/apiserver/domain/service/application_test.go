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
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test application service function", func() {
	var (
		testBDCName    = "test-bdc"
		testBDCNs      = "test-bdc-ns"
		AppFormName    = "test-app"
		testAppDefType = "test"
		testAppName    = testBDCName + "-" + AppFormName
	)

	BeforeEach(func() {
		InitTestEnv()
	})

	It("Test CreateApplication function", func() {

		By("prepare create application request: application definition")

		req := v1dto.CreateApplicationRequest{
			CreateApplicationRequestBody: v1dto.CreateApplicationRequestBody{
				AppTemplateType: testAppDefType,
				AppFormName:     AppFormName,
			},
		}
		req.Properties = utils.Object2RawExtension(map[string]interface{}{
			"url":     "https://nx.test.com/repository/helm-hosted/",
			"version": "10.12.0",
		})
		req.BDC = &v1dto.BigDataClusterBase{
			Name:      testBDCName,
			DefaultNS: testBDCNs,
		}
		By("test create application")
		_, err := appService.CreateApplication(context.TODO(), req)
		Expect(err).Should(BeNil())
		//Expect(cmp.Diff(base.BDC.Name, req.BDC.Name)).Should(BeEmpty())
	})

	It("Test ListApplications function", func() {
		options := v1dto.ListOptions{}
		_, err := appService.ListApplications(context.TODO(), options)
		Expect(err).Should(BeNil())
	})

	It("Test ListApplications function  by their labels as a selector to restrict the list of returned objects", func() {
		options := v1dto.ListOptions{Labels: map[string]string{constants.LabelBDCName: testBDCName}}
		_, err := appService.ListApplications(context.TODO(), options)
		Expect(err).Should(BeNil())
	})

	It("Test GetlApplication function", func() {
		_, err := appService.GetApplication(context.TODO(), testAppName)
		Expect(err).Should(BeNil())

	})

	It("Test DetailApplication function", func() {
		_, err := appService.DetailApplication(context.TODO(), testAppName)
		Expect(err).Should(BeNil())

	})

	It("Test UpdateApplication function", func() {
		By("prepare update application request")
		req := v1dto.UpdateApplicationRequest{
			AppName: testAppName,
		}
		req.BDC = &v1dto.BigDataClusterBase{
			Name:      testBDCName,
			DefaultNS: testBDCNs,
		}
		req.Properties = utils.Object2RawExtension(map[string]interface{}{
			"url":     "https://nx.test.com/repository/helm-hosted/",
			"version": "10.12.0",
		})
		By("test update application")
		_, err := appService.UpdateApplication(context.TODO(), req)
		Expect(err).Should(BeNil())

	})

	It("Test DeleteApplication function", func() {
		err := appService.DeleteApplication(context.TODO(), testAppName)
		Expect(err).Should(BeNil())

	})
})
