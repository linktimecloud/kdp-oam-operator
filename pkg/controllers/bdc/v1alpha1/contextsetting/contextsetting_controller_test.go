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

package contextsetting

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/pkg/utils"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"time"
)

var _ = Describe("Test ContextSetting Controller", func() {
	ctx := context.Background()

	BeforeEach(func() {
		var validDef = utils.ReadContent(filepath.Join("../../../../..", "charts", "kdp-oam-operator", "templates", "deftemplate", "contextsetting-def.yaml"))
		// Create namespace
		ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kdp-system"}}

		Eventually(func() error {
			return k8sClient.Create(ctx, &ns)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil()))

		// Create bdc definition map into configmap
		bdcDefinitionMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bdc-definition-map",
				Namespace: "kdp-system",
			},
			Data: map[string]string{
				"kerberos-ContextSetting": "ctx-setting-kerberos",
			},
		}
		Eventually(func() error {
			return k8sClient.Create(ctx, &bdcDefinitionMap)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil()))

		// Create xDefinition
		var def bdcv1alpha1.XDefinition
		err := yaml.Unmarshal([]byte(validDef), &def)
		if err != nil {
			return
		}

		Eventually(func() error {
			return k8sClient.Create(ctx, &def)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil()))
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "contextsetting-def"}, &def)).Should(Succeed())
	})

	Context("When the ContextSetting dependent xDefinition doesn't exist, should occur error", func() {
		It("Applying ContextSetting", func() {
			By("Apply ContextSetting")

			var validInstance = `
apiVersion: bdc.kdp.io/v1alpha1
kind: ContextSetting
metadata:
  name: contextsetting-sample
  annotations:
    # "setting.ctx.bdc.kdp.io/adopt": "true"
    "bdc.kdp.io/org": "bdctestorg"
    "bdc.kdp.io/name": "bdc-sample2"
    "setting.ctx.bdc.kdp.io/type": "HDFS/KDC/..."
    "setting.ctx.bdc.kdp.io/origin": "system/manual"
spec:
  type: ctx-setting-common
  name: bdc-test-ctx-common-auth-cfg
  properties:
    common_key: value
    commmon_key2: value2

`
			var contextSetting bdcv1alpha1.ContextSetting
			Expect(yaml.Unmarshal([]byte(validInstance), &contextSetting)).Should(BeNil())
			Expect(k8sClient.Create(ctx, &contextSetting)).Should(Succeed())
		})
	})
})
