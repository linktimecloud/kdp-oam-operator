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
	"encoding/json"
	"kdp-oam-operator/pkg/utils"
	"reflect"
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
