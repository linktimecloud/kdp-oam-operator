/*
Copyright 2023 KDP(Kubernetes Data Platform).

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

package xdefinitions

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var _ = Describe("Test XDefinition Controller", func() {
	ctx := context.Background()

	BeforeEach(func() {
		// Create namespace
		ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kdp-system"}}

		Eventually(func() error {
			return k8sClient.Create(ctx, &ns)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil()))

	})

	Context("When the XDefinition and other CR mapping configmap doesn't exist, should occur error", func() {
		It("Applying XDefinition", func() {
			By("Apply XDefinition")

			bdcDefinitionMap := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bdc-definition-map",
					Namespace: "kdp-system",
				},
				Data: map[string]string{
					"default-BigDataCluster": "bigdatacluster-def",
				},
			}
			Expect(k8sClient.Create(ctx, &bdcDefinitionMap)).Should(Succeed())
		})
	})
})
