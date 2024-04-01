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
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/pkg/utils"
	"os"
	"reflect"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestRenderXDefinitionJSONSchem(t *testing.T) {
	dynamicParameterValues := map[string][]interface{}{
		"dependencies.zookeeper.quorum": {"zookeeper.admin.svc.cluster.local"},
		"kerberos.enable":               {true},
	}

	jsonSchemaStr := `{
		"properties": {
			"chart": {
				"default": "1.1.3",
				"type": "string"
			},
			"dependencies": {
				"properties": {
					"hdfs": {
						"properties": {
							"config": {
								"type": "string",
								"enum": []
							}
						},
						"required": ["config"],
						"type": "object"
					},
					"zookeeper": {
						"properties": {
							"port": {
								"type": "string",
								"enum": ["2181"]
							},
							"quorum": {
								"type": "string",
								"enum": ["zookeeper.admin.svc.cluster.local"]
							}
						},
						"required": ["quorum", "port"],
						"type": "object"
					}
				},
				"required": ["hdfs", "zookeeper"],
				"type": "object"
			},
			"image": {
				"default": "2.4.18-SNAPSHOT_kdp-1.1.1",
				"type": "string"
			},
			"kerberos": {
				"properties": {
					"enable": {
						"type": "boolean",
						"enum": [true]
					}
				},
				"required": ["enable"],
				"type": "object"
			}
		},
		"required": ["dependencies", "chart", "image", "kerberos"],
		"type": "object"
	}`

	jsonSchema := map[string]interface{}{
		"properties": map[string]interface{}{
			"chart": map[string]interface{}{
				"default": "1.1.3",
				"type":    "string",
			},
			"dependencies": map[string]interface{}{
				"properties": map[string]interface{}{
					"hdfs": map[string]interface{}{
						"properties": map[string]interface{}{
							"config": map[string]interface{}{
								"type": "string",
								"enum": []interface{}{},
							},
						},
						"required": []interface{}{"config"},
						"type":     "object",
					},
					"zookeeper": map[string]interface{}{
						"properties": map[string]interface{}{
							"port": map[string]interface{}{
								"type": "string",
								"enum": []interface{}{"2181"},
							},
							"quorum": map[string]interface{}{
								"type": "string",
								"enum": []interface{}{"zookeeper.admin.svc.cluster.local"},
							},
						},
						"required": []interface{}{"quorum", "port"},
						"type":     "object",
					},
				},
				"required": []interface{}{"hdfs", "zookeeper"},
				"type":     "object",
			},
			"image": map[string]interface{}{
				"default": "2.4.18-SNAPSHOT_kdp-1.1.1",
				"type":    "string",
			},
			"kerberos": map[string]interface{}{
				"properties": map[string]interface{}{
					"enable": map[string]interface{}{
						"type": "boolean",
						"enum": []interface{}{true},
					},
				},
				"required": []interface{}{"enable"},
				"type":     "object",
			},
		},
		"required": []interface{}{"dependencies", "chart", "image", "kerberos"},
		"type":     "object",
	}

	if !reflect.DeepEqual(jsonSchema, utils.StringToMap(jsonSchemaStr)) {
		t.Errorf("string marshal to map failed, expected: %v, actual: %v", jsonSchema, jsonSchemaStr)
	}

	expectedSchema := map[string]interface{}{
		"properties": map[string]interface{}{
			"chart": map[string]interface{}{
				"default": "1.1.3",
				"type":    "string",
			},
			"dependencies": map[string]interface{}{
				"properties": map[string]interface{}{
					"hdfs": map[string]interface{}{
						"properties": map[string]interface{}{
							"config": map[string]interface{}{
								"type": "string",
								"enum": []interface{}{},
							},
						},
						"required": []interface{}{"config"},
						"type":     "object",
					},
					"zookeeper": map[string]interface{}{
						"properties": map[string]interface{}{
							"port": map[string]interface{}{
								"type": "string",
								"enum": []interface{}{"2181"},
							},
							"quorum": map[string]interface{}{
								"type": "string",
								"enum": []interface{}{"zookeeper.admin.svc.cluster.local"},
							},
						},
						"required": []interface{}{"quorum", "port"},
						"type":     "object",
					},
				},
				"required": []interface{}{"hdfs", "zookeeper"},
				"type":     "object",
			},
			"image": map[string]interface{}{
				"default": "2.4.18-SNAPSHOT_kdp-1.1.1",
				"type":    "string",
			},
			"kerberos": map[string]interface{}{
				"properties": map[string]interface{}{
					"enable": map[string]interface{}{
						"type": "boolean",
						"enum": []interface{}{true},
					},
				},
				"required": []interface{}{"enable"},
				"type":     "object",
			},
		},
		"required": []interface{}{"dependencies", "chart", "image", "kerberos"},
		"type":     "object",
	}
	expectedSchemaJSON, _ := json.Marshal(expectedSchema)

	updatedSchemaJSON, _ := renderXDefinitionJSONSchem(jsonSchemaStr, dynamicParameterValues)
	if !reflect.DeepEqual(*updatedSchemaJSON, string(expectedSchemaJSON)) {

		t.Errorf("Updated schema does not match expected schema.\nGot:\n%v\nExpected:\n%s", *updatedSchemaJSON, string(expectedSchemaJSON))
	}
}

var _ = Describe("Test definition service function", func() {
	var (
		// testBDCName = "test-bdc"
		testDefType = "test"
	)

	BeforeEach(func() {
		InitTestEnv()
	})

	It("Test GetlDefinition function", func() {
		By("prepare create application request: application definition")
		// initDefinitions(kubeClient)
		applicationDef, err := os.ReadFile("./testdata/application-def.yaml")
		fmt.Printf("applicationDef: %+v", string(applicationDef))
		Expect(err).Should(BeNil())
		var def bdcv1alpha1.XDefinition
		err = yaml.Unmarshal(applicationDef, &def)
		Expect(err).Should(BeNil())
		fmt.Printf("def: %+v", def)
		Expect(kubeClient.Create(context.TODO(), &def))

		By("list definition")
		defList := new(bdcv1alpha1.XDefinitionList)
		err = kubeClient.List(context.TODO(), defList)
		Expect(err).Should(BeNil())
		Expect(defList).ShouldNot(BeNil())
		fmt.Printf("defList: %+v", defList.Items)

		By("get application definition")
		var selectDefinition *bdcv1alpha1.XDefinition
		for _, item := range defList.Items {
			if item.Spec.APIResource.Definition.Kind == "Application" && item.Spec.APIResource.Definition.Type == testDefType {
				selectDefinition = &item
			}
		}
		Expect(selectDefinition).ShouldNot(BeNil())
		Expect(cmp.Diff(selectDefinition.Name, "application-test")).Should(BeEmpty())
		Expect(selectDefinition.Spec.APIResource.Definition.Type).Should(Equal(testDefType))

	})
})
