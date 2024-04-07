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
	"k8s.io/apimachinery/pkg/runtime"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
)

var testContextSecretName = "test-context-secret"

func prepareContextSecret() {
	defer GinkgoRecover()
	By("init context secret")

	var validContextSecretInstance = bdcv1alpha1.ContextSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContextSecret",
			APIVersion: bdcv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				constants.AnnotationBDCName:             testBDCName,
				constants.AnnotationBDCDefaultNamespace: testBDCNs,
				constants.AnnotationCtxSettingOrigin:    "system",
			},
			Labels: map[string]string{
				constants.LabelBDCName:      testBDCName,
				constants.AnnotationOrgName: testBDCOrg,
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
	Expect(kubeClient.Create(context.TODO(), &validContextSecretInstance)).Should(Succeed())

}

var _ = Describe("Test context secret rest api", func() {
	It("Test listing context secrets with error(test bdc instance not found)", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/contextsecrets", map[string]string{
			"bdcName": testBDCName + "-1",
		})
		Expect(res.StatusCode).Should(Equal(404))
	})

	It("Test listing context secrets is empty", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/contextsecrets", map[string]string{
			"bdcName": testBDCName,
		})
		var ctxSecrets v1dto.ListContextSecretsResponse
		Expect(decodeResponseBody(res, &ctxSecrets)).Should(Succeed())
		Expect(cmp.Diff(len(ctxSecrets.Data), 0)).Should(BeEmpty())
	})

	It("Test create context secret", func() {
		prepareContextSecret()
	})

	It("Test listing context secret", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/contextsecrets", map[string]string{
			"bdcName": testBDCName,
		})
		var ctxSecrets v1dto.ListContextSecretsResponse

		Expect(decodeResponseBody(res, &ctxSecrets)).Should(Succeed())
		Expect(cmp.Diff(len(ctxSecrets.Data), 1)).Should(BeEmpty())
	})

	It("Test get context secret", func() {
		defer GinkgoRecover()
		res := getRequest("/contextsecrets/" + testContextSecretName)
		var ctxSecretBase v1dto.GetContextSecretResponse
		Expect(decodeResponseBody(res, &ctxSecretBase)).Should(Succeed())
		Expect(cmp.Diff(ctxSecretBase.Data.Name, testContextSecretName)).Should(BeEmpty())
	})
})
