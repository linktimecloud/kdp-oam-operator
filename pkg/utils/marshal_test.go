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
	"bytes"
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func TestObject2Map(t *testing.T) {
	type args struct {
		obj interface{}
	}
	case1Obj := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}

	case1Expected := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}

	type TestPerson struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	case2Obj := TestPerson{
		Name:  "Alice",
		Age:   25,
		Email: "alice@example.com",
	}

	case2Expected := map[string]interface{}{
		"name":  "Alice",
		"age":   25,
		"email": "alice@example.com",
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "Test Case 1",
			args:    args{case1Obj},
			want:    case1Expected,
			wantErr: false,
		},
		{
			name:    "Test Case 2",
			args:    args{case2Obj},
			want:    case2Expected,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Object2Map(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Object2Map() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//    t.Errorf("Object2Map() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

// TestRawExtension2Map tests the functionality of RawExtension2Map function.
func TestRawExtension2Map(t *testing.T) {
	tests := []struct {
		name       string
		raw        *runtime.RawExtension
		wantResult map[string]interface{}
		wantErr    error
	}{
		// Test cases
		{
			name:       "Nil RawExtension",
			raw:        nil,
			wantResult: nil,
			wantErr:    nil,
		},
		{
			name:       "Empty RawExtension",
			raw:        &runtime.RawExtension{},
			wantResult: nil,
			wantErr:    nil,
		},
		{
			name: "RawExtension with StringData",
			raw: &runtime.RawExtension{
				Raw:    []byte(`{"a":{"c":"d"},"b":1}`),
				Object: nil,
			},
			wantResult: map[string]interface{}{
				"a": map[string]interface{}{
					"c": "d",
				},
				"b": float64(1),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := RawExtension2Map(tt.raw)
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("RawExtension2Map() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("RawExtension2Map() gotErr = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestObject2RawExtension(t *testing.T) {
	obj := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "John",
		Age:  30,
	}

	rawExtension := Object2RawExtension(obj)
	temp, _ := json.Marshal(obj)
	expected := &runtime.RawExtension{
		Raw: temp,
	}

	if !reflect.DeepEqual(rawExtension, expected) {
		t.Errorf("Object2RawExtension() = %v, expected %v", rawExtension, expected)
	}
}

// TestObject2Unstructured tests the Object2Unstructured function.
func TestObject2Unstructured(t *testing.T) {
	obj := runtime.Object(&corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
	})

	unstructuredObj, err := Object2Unstructured(obj)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Assert unstructuredObj has the correct type
	if unstructuredObj.GetAPIVersion() != "v1" || unstructuredObj.GetKind() != "Pod" {
		t.Errorf("Unexpected unstructuredObj type. Expected: v1 Pod, Got: %s %s", unstructuredObj.GetAPIVersion(), unstructuredObj.GetKind())
	}

	// Assert unstructuredObj has the correct fields
	if unstructuredObj.GetName() != "test-pod" {
		t.Errorf("Unexpected name field. Expected: test-pod, Got: %s", unstructuredObj.GetName())
	}

}

func TestMustJSONMarshal(t *testing.T) {
	// Test cases
	tests := []struct {
		input    interface{}
		expected string
	}{
		{
			input:    "test",
			expected: `"` + "test" + `"`,
		},
		{
			input:    123,
			expected: `123`,
		},
		{
			input:    true,
			expected: `true`,
		},
		{
			input:    false,
			expected: `false`,
		},
		{
			input:    nil,
			expected: `null`,
		},
		{
			input: struct {
				Name string
				Age  int
			}{
				Name: "John",
				Age:  30,
			},
			expected: `{"Name":"John","Age":30}`,
		},
	}

	// Run tests
	for _, tc := range tests {
		result := MustJSONMarshal(tc.input)

		if !bytes.Equal(result, []byte(tc.expected)) {
			t.Errorf("expected %q, got %q", tc.expected, string(result))
		}
	}
}

func TestRawExtension2Unstructured(t *testing.T) {
	raw := &runtime.RawExtension{
		Raw: []byte(`{"name":"example"}`),
	}

	expected := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"name": "example",
		},
	}

	result, err := RawExtension2Unstructured(raw)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Errorf("Expected non-nil result, got nil")
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Unexpected result: %v, expected: %v", result, expected)
	}
}
