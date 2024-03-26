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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
)

var _ = Describe("Test bigdata cluster service function", func() {
	var (
		testBigDataClusterName = "test-bigdata-cluster"
		testBigDataClusterOrg  = "test-bigdata-cluster-org"
	)

	BeforeEach(func() {
		InitTestEnv()
	})

	It("Test CreateBigDataCluster function", func() {

		By("init bigdata cluster")

		var validBigDataClusterInstance = bdcv1alpha1.BigDataCluster{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					constants.AnnotationOrgName:        testBigDataClusterOrg,
					constants.AnnotationBDCAlias:       "test alias",
					constants.AnnotationBDCName:        testBigDataClusterName,
					constants.AnnotationBDCDescription: "test description",
				},
				Labels: map[string]string{
					constants.LabelBDCName: testBigDataClusterName,
				},
				Name: testBigDataClusterName,
			},
			Spec: bdcv1alpha1.BigDataClusterSpec{
				Frozen:   false,
				Disabled: false,
				Namespaces: []bdcv1alpha1.Namespace{
					{
						Name:      "ns-bdc-sample1",
						IsDefault: true,
					},
				},
			},
		}
		Expect(kubeClient.Create(ctx, &validBigDataClusterInstance)).Should(Succeed())
	})

	It("Test ListBigDataClusters function", func() {
		options := v1dto.ListOptions{}
		_, err := bigDataClusterService.ListBigDataClusters(context.TODO(), options)
		Expect(err).Should(BeNil())
	})

	It("Test ListBigDataClusters function  by their labels as a selector to restrict the list of returned objects", func() {
		options := v1dto.ListOptions{Labels: map[string]string{constants.AnnotationOrgName: testBigDataClusterOrg}}
		_, err := bigDataClusterService.ListBigDataClusters(context.TODO(), options)
		Expect(err).Should(BeNil())
	})

	It("Test DetailBigDataCluster function", func() {
		_, err := bigDataClusterService.GetBigDataCluster(context.TODO(), testBigDataClusterName)
		Expect(err).Should(BeNil())

	})
})
