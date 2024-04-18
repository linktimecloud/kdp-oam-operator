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
	"context"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/utils"
)

var testBDCName = "test-bdc"
var testBDCNs = "test-bdc-ns"
var testBDCOrg = "test-bdc-org"

var testAppFormName = "test-app"
var testAppDefType = "test"
var testAppName = testBDCName + "-" + testAppFormName

func prepareBigDataCluster() {
	defer GinkgoRecover()
	By("init bigdata cluster")

	var validBigDataClusterInstance = bdcv1alpha1.BigDataCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BigDataCluster",
			APIVersion: bdcv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				constants.AnnotationOrgName:        testBDCOrg,
				constants.AnnotationBDCAlias:       "test bdc alias",
				constants.AnnotationBDCName:        testBDCName,
				constants.AnnotationBDCDescription: "test bdc description",
			},
			Labels: map[string]string{
				constants.LabelBDCName:      testBDCName,
				constants.AnnotationOrgName: testBDCOrg,
			},
			Name: testBDCName,
		},
		Spec: bdcv1alpha1.BigDataClusterSpec{
			Frozen:   false,
			Disabled: false,
			Namespaces: []bdcv1alpha1.Namespace{
				{
					Name:      testBDCNs,
					IsDefault: true,
				},
			},
		},
	}
	Expect(kubeClient.Create(context.TODO(), &validBigDataClusterInstance)).Should(Succeed())
}

var _ = Describe("Test application rest api", func() {
	It("Test listing applications with error(test bdc instance not found)", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/applications", map[string]string{
			"bdcName": testBDCName + "-1",
		})
		Expect(res.StatusCode).Should(Equal(404))
	})

	It("Test listing applications is empty(Do not specify bdc)", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/applications", map[string]string{})
		var apps v1dto.ListApplicationsResponse
		Expect(decodeResponseBody(res, &apps)).Should(Succeed())
		Expect(cmp.Diff(len(apps.Data), 0)).Should(BeEmpty())
	})

	It("Test listing applications is empty", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/applications", map[string]string{
			"bdcName": testBDCName,
		})
		var apps v1dto.ListApplicationsResponse
		Expect(decodeResponseBody(res, &apps)).Should(Succeed())
		Expect(cmp.Diff(len(apps.Data), 0)).Should(BeEmpty())
	})

	It("Test create application", func() {
		defer GinkgoRecover()
		req := v1dto.CreateApplicationRequest{
			CreateApplicationRequestBody: v1dto.CreateApplicationRequestBody{
				AppTemplateType: testAppDefType,
				AppFormName:     testAppFormName,
			},
		}
		req.Properties = utils.Object2RawExtension(map[string]interface{}{
			"url":     "https://nx.test.com/repository/helm-hosted/",
			"version": "10.12.0",
		})
		res := postRequest("/bigdataclusters/"+testBDCName+"/applications", req)
		var appBase v1dto.ApplicationBase

		Expect(decodeResponseBody(res, &appBase)).Should(Succeed())
	})

	It("Test listing applications", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/applications", map[string]string{
			"bdcName": testBDCName,
		})
		var apps v1dto.ListApplicationsResponse

		Expect(decodeResponseBody(res, &apps)).Should(Succeed())
		Expect(cmp.Diff(len(apps.Data), 1)).Should(BeEmpty())
	})

	It("Test listing applications is empty(Do not specify bdc)", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/applications", map[string]string{})
		var apps v1dto.ListApplicationsResponse
		Expect(decodeResponseBody(res, &apps)).Should(Succeed())
		Expect(cmp.Diff(len(apps.Data), 1)).Should(BeEmpty())
	})

	It("Test get application", func() {
		defer GinkgoRecover()
		res := getRequest("/applications/" + testAppName)
		var appBase v1dto.GetApplicationsResponse
		Expect(decodeResponseBody(res, &appBase)).Should(Succeed())
		Expect(cmp.Diff(appBase.Data.Name, testAppName)).Should(BeEmpty())
	})

	It("Test update app", func() {
		defer GinkgoRecover()
		var req = v1dto.UpdateApplicationRequest{
			UpdateApplicationRequestBody: v1dto.UpdateApplicationRequestBody{
				ApplicationSpecProperties: v1dto.ApplicationSpecProperties{
					Properties: utils.Object2RawExtension(map[string]interface{}{
						"url":     "https://nx.test.com/repository/helm-hosted/",
						"version": "10.12.1",
					}),
				},
			},
		}
		res := putRequest("/applications/"+testAppName, req)
		var appBase v1dto.GetApplicationsResponse
		Expect(decodeResponseBody(res, &appBase)).Should(Succeed())
	})

	It("Test detail application", func() {
		defer GinkgoRecover()
		res := getRequest("/applications/" + testAppName + "/detail")
		var detail v1dto.ApplicationRawResponse
		Expect(decodeResponseBody(res, &detail)).Should(Succeed())
	})

	It("Test delete application", func() {
		defer GinkgoRecover()
		res := deleteRequest("/applications/" + testAppName)
		Expect(decodeResponseBody(res, nil)).Should(Succeed())
	})
})
