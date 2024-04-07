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

var testContextSettingName = "test-context-setting"

func prepareContextSetting() {
	defer GinkgoRecover()
	By("init context setting")

	var validContextSettingInstance = bdcv1alpha1.ContextSetting{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContextSetting",
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
			Name: testContextSettingName,
		},
		Spec: bdcv1alpha1.ContextSettingSpec{
			Name: testContextSettingName,
			Properties: &runtime.RawExtension{
				Raw: []byte(`{"test": "test"}`),
			},
			Type: "test",
		},
	}
	Expect(kubeClient.Create(context.TODO(), &validContextSettingInstance)).Should(Succeed())

}

var _ = Describe("Test context setting rest api", func() {
	It("Test listing context settings with error(test bdc instance not found)", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/contextsettings", map[string]string{
			"bdcName": testBDCName + "-1",
		})
		Expect(res.StatusCode).Should(Equal(404))
	})

	It("Test listing context settings is empty", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/contextsettings", map[string]string{
			"bdcName": testBDCName,
		})
		var ctxSettings v1dto.ListContextSettingsResponse
		Expect(decodeResponseBody(res, &ctxSettings)).Should(Succeed())
		Expect(cmp.Diff(len(ctxSettings.Data), 0)).Should(BeEmpty())
	})

	It("Test create context setting", func() {
		prepareContextSetting()
	})

	It("Test listing context settings", func() {
		defer GinkgoRecover()
		res := getWithQueryRequest("/contextsettings", map[string]string{
			"bdcName": testBDCName,
		})
		var ctxSettings v1dto.ListContextSettingsResponse

		Expect(decodeResponseBody(res, &ctxSettings)).Should(Succeed())
		Expect(cmp.Diff(len(ctxSettings.Data), 1)).Should(BeEmpty())
	})

	It("Test get context setting", func() {
		defer GinkgoRecover()
		res := getRequest("/contextsettings/" + testContextSettingName)
		var ctxSettingBase v1dto.GetContextSettingResponse
		Expect(decodeResponseBody(res, &ctxSettingBase)).Should(Succeed())
		Expect(cmp.Diff(ctxSettingBase.Data.Name, testContextSettingName)).Should(BeEmpty())
	})
})
