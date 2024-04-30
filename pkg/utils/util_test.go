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

package utils

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

// TestAddLabels tests the AddLabels function.
func TestAddLabels(t *testing.T) {
	// Create a new instance of the struct.
	s := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deployment",
				"namespace": "test-namespace",
			},
		},
	}

	// Define the expected labels.
	expectedLabels := map[string]string{
		"label1": "value1",
		"label2": "value2",
		"label3": "value3",
	}

	// Call the AddLabels function.
	AddLabels(&s, expectedLabels)

	// Check if the labels are added correctly.
	if len(s.GetLabels()) != len(expectedLabels) {
		t.Errorf("Expected labels length: %d, got: %d", len(expectedLabels), len(s.GetLabels()))
	}

	for key, value := range expectedLabels {
		if s.GetLabels()[key] != value {
			t.Errorf("Expected label %s value: %s, got: %s", key, value, s.GetLabels()[key])
		}
	}
}

func TestAddAnnotations(t *testing.T) {
	// Create a new instance of the struct.
	s := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deployment",
				"namespace": "test-namespace",
			},
		},
	}
	annotations := map[string]string{
		"anno1": "value1",
		"anno2": "value2",
	}

	AddAnnotations(&s, annotations)

	if len(s.GetAnnotations()) != len(annotations) {
		t.Errorf("Expected %d annotations, but got %d", len(annotations), len(s.GetAnnotations()))
	}

	for key, value := range annotations {
		if s.GetAnnotations()[key] != value {
			t.Errorf("Expected annotation key '%s' to have value '%s', but got '%s'", key, value, s.GetAnnotations()[key])
		}
	}
}

func TestMergeMapOverrideWithDst(t *testing.T) {
	// Test case 1
	srcMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	dstMap := map[string]string{
		"key2": "newvalue2",
		"key3": "value3",
	}
	expectedMap := map[string]string{
		"key1": "value1",
		"key2": "newvalue2",
		"key3": "value3",
	}

	resultMap := MergeMapOverrideWithDst(srcMap, dstMap)
	if len(resultMap) != len(expectedMap) {
		t.Errorf("Test case 1 failed: Expected map length %d, but got %d", len(expectedMap), len(resultMap))
	}
	for key, value := range expectedMap {
		if resultMap[key] != value {
			t.Errorf("Test case 1 failed: Expected value %s for key %s, but got %s", value, key, resultMap[key])
		}
	}

	// Test case 2
	srcMap = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	dstMap = map[string]string{
		"key2": "newvalue2",
	}
	expectedMap = map[string]string{
		"key1": "value1",
		"key2": "newvalue2",
	}

	resultMap = MergeMapOverrideWithDst(srcMap, dstMap)
	if len(resultMap) != len(expectedMap) {
		t.Errorf("Test case 2 failed: Expected map length %d, but got %d", len(expectedMap), len(resultMap))
	}
	for key, value := range expectedMap {
		if resultMap[key] != value {
			t.Errorf("Test case 2 failed: Expected value %s for key %s, but got %s", value, key, resultMap[key])
		}
	}

	// Test case 3
	srcMap = map[string]string{}
	dstMap = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	expectedMap = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	resultMap = MergeMapOverrideWithDst(srcMap, dstMap)
	if len(resultMap) != len(expectedMap) {
		t.Errorf("Test case 3 failed: Expected map length %d, but got %d", len(expectedMap), len(resultMap))
	}
	for key, value := range expectedMap {
		if resultMap[key] != value {
			t.Errorf("Test case 3 failed: Expected value %s for key %s, but got %s", value, key, resultMap[key])
		}
	}

	// Test case 4
	srcMap = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	dstMap = map[string]string{}
	expectedMap = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	resultMap = MergeMapOverrideWithDst(srcMap, dstMap)
	if len(resultMap) != len(expectedMap) {
		t.Errorf("Test case 4 failed: Expected map length %d, but got %d", len(expectedMap), len(resultMap))
	}
	for key, value := range expectedMap {
		if resultMap[key] != value {
			t.Errorf("Test case 4 failed: Expected value %s for key %s, but got %s", value, key, resultMap[key])
		}
	}

	// Test case 5
	srcMap = map[string]string{}
	dstMap = map[string]string{}

	MergeMapOverrideWithDst(srcMap, dstMap)

	if len(srcMap) != 0 {
		t.Errorf("Expected srcMap to be empty, but got %d elements", len(srcMap))
	}

	if len(dstMap) != 0 {
		t.Errorf("Expected dstMap to be empty, but got %d elements", len(dstMap))
	}
}

func TestMergeMapOverrideWithFilters(t *testing.T) {
	type args struct {
		src           map[string]string
		dst           map[string]string
		filterKeyList []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "test1",
			args: args{
				src:           map[string]string{"apple": "apple", "banana": "banana", "cherry": "cherry"},
				dst:           map[string]string{"apple": "apple", "banana": "banana", "cherry": "cherry"},
				filterKeyList: nil,
			},
			want: map[string]string{"apple": "apple", "banana": "banana", "cherry": "cherry"},
		},
		{
			name: "test1",
			args: args{
				src:           map[string]string{"apple1": "apple", "banana": "banana", "cherry": "cherry"},
				dst:           map[string]string{"apple2": "apple", "banana": "banana", "cherry": "cherry"},
				filterKeyList: nil,
			},
			want: map[string]string{"apple1": "apple", "banana": "banana", "cherry": "cherry", "apple2": "apple"},
		},
		{
			name: "test1",
			args: args{
				src:           map[string]string{"apple1": "apple", "banana": "banana", "cherry": "cherry"},
				dst:           map[string]string{"apple2": "apple", "banana": "banana", "cherry": "cherry"},
				filterKeyList: []string{"apple1"},
			},
			want: map[string]string{"banana": "banana", "cherry": "cherry", "apple2": "apple"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeMapOverrideWithFilters(tt.args.src, tt.args.dst, tt.args.filterKeyList); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeMapOverrideWithFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListContains(t *testing.T) {
	type args struct {
		s   []string
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				s:   []string{"apple", "banana", "cherry"},
				str: "test",
			},
			want: false,
		},
		{
			name: "test1",
			args: args{
				s:   []string{"apple", "banana", "cherry"},
				str: "apple",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ListContains(tt.args.s, tt.args.str); got != tt.want {
				t.Errorf("ListContains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToMap(t *testing.T) {
	content := `{"name": "Alice", "age": 25}`

	expected := map[string]interface{}{
		"name": "Alice",
		"age":  float64(25),
	}

	res := StringToMap(content)

	if len(res) != len(expected) {
		t.Errorf("Length of result map is incorrect, got: %d, want: %d", len(res), len(expected))
	}

	for key, value := range expected {
		if resValue, ok := res[key]; !ok {
			t.Errorf("Result map is missing key: %s", key)
		} else if resValue != value {
			t.Errorf("Result map value is incorrect for key: %s, got: %v, want: %v", key, resValue, value)
		}
	}
}

func TestFinalizerExists(t *testing.T) {
	type args struct {
		o         v1.Object
		finalizer string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TestFinalizerExists test case1",
			args: args{
				o:         &v1.ObjectMeta{Finalizers: []string{"finalizer1", "finalizer2"}},
				finalizer: "finalizer1",
			},
			want: true,
		},
		{
			name: "TestFinalizerExists test case2",
			args: args{
				o:         &v1.ObjectMeta{Finalizers: []string{"finalizer1", "finalizer2"}},
				finalizer: "finalizer3",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FinalizerExists(tt.args.o, tt.args.finalizer); got != tt.want {
				t.Errorf("FinalizerExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveFinalizer(t *testing.T) {
	type args struct {
		o         v1.Object
		finalizer string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TestRemoveFinalizer test case1",
			args: args{
				o:         &v1.ObjectMeta{Finalizers: []string{"finalizer1", "finalizer2"}},
				finalizer: "finalizer1",
			},
		},
		{
			name: "TestRemoveFinalizer test case2",
			args: args{
				o:         &v1.ObjectMeta{Finalizers: []string{"finalizer1", "finalizer2"}},
				finalizer: "finalizer3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RemoveFinalizer(tt.args.o, tt.args.finalizer)
		})
	}
}

func TestGetEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("EXISTING_KEY", "existing_value")

	// 测试存在的键
	existingValue := GetEnv("EXISTING_KEY", "fallback")
	if existingValue != "existing_value" {
		t.Errorf("GetEnv(\"EXISTING_KEY\", \"fallback\") = %s; want existing_value", existingValue)
	}

	// 测试不存在的键
	nonexistentValue := GetEnv("NONEXISTENT_KEY", "fallback")
	if nonexistentValue != "fallback" {
		t.Errorf("GetEnv(\"NONEXISTENT_KEY\", \"fallback\") = %s; want fallback", nonexistentValue)
	}
}

func TestStringToInt64(t *testing.T) {
	t.Run("Valid string to int64 conversion", func(t *testing.T) {
		str := "123"
		expected := int64(123)
		result := StringToInt64(str, 0)
		if result != expected {
			t.Errorf("Expected %d but got %d", expected, result)
		}
	})

	t.Run("Invalid string to int64 conversion", func(t *testing.T) {
		str := "invalid"
		expected := int64(0)
		result := StringToInt64(str, 0)
		if result != expected {
			t.Errorf("Expected %d but got %d", expected, result)
		}
	})
}

func TestStringToInt(t *testing.T) {
	t.Run("Valid string to int conversion", func(t *testing.T) {
		str := "123"
		expected := 123
		result := StringToInt(str, 0)
		if result != expected {
			t.Errorf("Expected %d but got %d", expected, result)
		}
	})

	t.Run("Invalid string to int conversion", func(t *testing.T) {
		str := "invalid"
		expected := 0
		result := StringToInt(str, 0)
		if result != expected {
			t.Errorf("Expected %d but got %d", expected, result)
		}
	})
}

func TestGenerateShortHashID(t *testing.T) {
	tests := []struct {
		length  int
		values  []string
		wantLen int
	}{
		{16, []string{"param1", "param2", "param3"}, 16},
		{8, []string{"param4", "param5", "param6"}, 8},
	}

	for _, tt := range tests {
		got, err := GenerateShortHashID(tt.length, tt.values...)
		if err != nil {
			t.Errorf("GenerateShortHashID(%d, %v) returned error: %v", tt.length, tt.values, err)
			continue
		}

		if len(got) != tt.wantLen {
			t.Errorf("GenerateShortHashID(%d, %v) = %s, want length %d", tt.length, tt.values, got, tt.wantLen)
		}
	}
}

func TestGetStatusCode(t *testing.T) {
	// 创建一个模拟的 HTTP 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 测试获取状态码
	statusCode, err := GetStatusCode(server.URL)
	if err != nil {
		t.Errorf("GetStatusCode() returned error: %v", err)
	}
	if statusCode != http.StatusOK {
		t.Errorf("GetStatusCode() returned unexpected status code: got %d, want %d", statusCode, http.StatusOK)
	}
}

// TestGetStringValue 测试从 map[string]interface{} 中获取字符串值功能
func TestGetStringValue(t *testing.T) {
	// 创建测试数据
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123, // 不是字符串类型
	}

	// 测试获取字符串值
	value := GetStringValue(data, "key1")
	if value != "value1" {
		t.Errorf("GetStringValue() returned unexpected value: got %s, want %s", value, "value1")
	}

	// 测试获取非字符串值的情况
	invalidValue := GetStringValue(data, "key2")
	if invalidValue != "" {
		t.Errorf("GetStringValue() returned unexpected value: got %s, want %s", invalidValue, "")
	}
}

// TestGetInt64Value 测试从 map[string]interface{} 中获取 int64 值功能
func TestGetInt64Value(t *testing.T) {
	// 创建测试数据
	data := map[string]interface{}{
		"key1": "value1", // 不是 int64 类型
		"key2": int64(123),
	}

	// 测试获取 int64 值
	value := GetInt64Value(data, "key2")
	if value != int64(123) {
		t.Errorf("GetInt64Value() returned unexpected value: got %d, want %d", value, int64(123))
	}

	// 测试获取非 int64 值的情况
	invalidValue := GetInt64Value(data, "key1")
	if invalidValue != 0 {
		t.Errorf("GetInt64Value() returned unexpected value: got %d, want %d", invalidValue, 0)
	}
}
