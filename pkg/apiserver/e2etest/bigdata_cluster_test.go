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
