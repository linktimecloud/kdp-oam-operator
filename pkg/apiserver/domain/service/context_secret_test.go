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
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/controllers/bdc/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test context secret service function", func() {
	var (
		testContextSecretName = "test-context-secret"
		testBigDataClusterOrg = "test-bigdata-cluster-org"
	)

	BeforeEach(func() {
		InitTestEnv()
	})

	It("Test CreateBigDataCluster function", func() {

		By("init bigdata cluster")

		var validContextSecretInstance = bdcv1alpha1.ContextSecret{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					constants.AnnotationBDCDefaultNamespace: "",
				},
				Labels: map[string]string{
					constants.LabelBDCName:      "",
					constants.AnnotationOrgName: "",
				},
				Name: testContextSecretName,
			},
			Spec: bdcv1alpha1.ContextSecretSpec{
				Name: testContextSecretName,
				Properties: &runtime.RawExtension{
					Raw: []byte(`{"test": "test"}`),
				},
				Type: "test",
			},
		}
		Expect(kubeClient.Create(ctx, &validContextSecretInstance)).Should(Succeed())
	})

	It("Test ListContextSecrets function", func() {
		options := v1dto.ListOptions{}
		_, err := contextSecretService.ListContextSecrets(context.TODO(), options)
		Expect(err).Should(BeNil())
	})

	It("Test ListContextSecrets function  by their labels as a selector to restrict the list of returned objects", func() {
		options := v1dto.ListOptions{Labels: map[string]string{constants.AnnotationOrgName: testBigDataClusterOrg}}
		_, err := contextSecretService.ListContextSecrets(context.TODO(), options)
		Expect(err).Should(BeNil())
	})

	It("Test DetailBigDataCluster function", func() {
		_, err := contextSecretService.GetContextSecret(context.TODO(), testContextSecretName)
		Expect(err).Should(BeNil())

	})
})
