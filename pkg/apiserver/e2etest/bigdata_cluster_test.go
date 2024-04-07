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
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
)

var _ = Describe("Test bigdata cluster rest api", func() {

	It("Test listing bigdata-clusters is empty", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/bigdataclusters", map[string]string{
			"labelSelector": "bdc.kdp.io/org=" + testBDCOrg + "-1",
		})
		var apps v1dto.ListBigDataClustersResponse
		Expect(decodeResponseBody(res, &apps)).Should(Succeed())
		Expect(cmp.Diff(len(apps.Data), 0)).Should(BeEmpty())
	})

	It("Test listing bigdata-clusters", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/bigdataclusters", map[string]string{
			"labelSelector": "bdc.kdp.io/org=" + testBDCOrg,
		})
		var bdcs v1dto.ListBigDataClustersResponse
		Expect(decodeResponseBody(res, &bdcs)).Should(Succeed())
		Expect(cmp.Diff(len(bdcs.Data), 1)).Should(BeEmpty())
	})

	It("Test get bigdata-cluster", func() {
		defer GinkgoRecover()
		res := getRequest("/bigdataclusters/" + testBDCName)
		var bdcBase v1dto.GetBigDataClusterResponse
		Expect(decodeResponseBody(res, &bdcBase)).Should(Succeed())
		Expect(cmp.Diff(bdcBase.Data.Name, testBDCName)).Should(BeEmpty())
	})
})
