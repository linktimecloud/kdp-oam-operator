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
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
	"strconv"
)

type labelAnnotationObject interface {
	GetLabels() map[string]string
	SetLabels(labels map[string]string)
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
}

func AddLabels(o labelAnnotationObject, labels map[string]string) {
	o.SetLabels(MergeMapOverrideWithDst(o.GetLabels(), labels))
}

// AddAnnotations will merge annotations with existing ones. If any conflict keys, use new value to override existing value.
func AddAnnotations(o labelAnnotationObject, annos map[string]string) {
	o.SetAnnotations(MergeMapOverrideWithDst(o.GetAnnotations(), annos))
}

// MergeMapOverrideWithDst merges two could be nil maps. Keep the dst for any conflicts,
func MergeMapOverrideWithDst(src, dst map[string]string) map[string]string {
	if src == nil && dst == nil {
		return nil
	}
	r := make(map[string]string)
	for k, v := range src {
		r[k] = v
	}
	// override the src for the same key
	for k, v := range dst {
		r[k] = v
	}
	return r
}

// MergeMapOverrideWithFilters merges two could be nil maps. Keep the dst for any conflicts and remove keys in filterKeyList
func MergeMapOverrideWithFilters(src, dst map[string]string, filterKeyList []string) map[string]string {
	if src == nil && dst == nil {
		return nil
	}
	r := make(map[string]string)
	for k, v := range src {
		if ListContains(filterKeyList, k) {
			continue
		}
		r[k] = v
	}
	// override the src for the same key
	for k, v := range dst {
		if ListContains(filterKeyList, k) {
			continue
		}
		r[k] = v
	}
	return r
}

func ListContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func StringToMap(content string) map[string]interface{} {
	var resMap map[string]interface{}
	klog.V(1).Infof("source content: %s", content)
	err := json.Unmarshal([]byte(content), &resMap)
	if err != nil {
		klog.Error(err, "convert string to map failed")
	}
	return resMap
}

// FinalizerExists checks whether given finalizer is already set.
func FinalizerExists(o metav1.Object, finalizer string) bool {
	f := o.GetFinalizers()
	for _, e := range f {
		if e == finalizer {
			return true
		}
	}
	return false
}

// RemoveFinalizer from the supplied Kubernetes object's metadata.
func RemoveFinalizer(o metav1.Object, finalizer string) {
	f := o.GetFinalizers()
	for i, e := range f {
		if e == finalizer {
			f = append(f[:i], f[i+1:]...)
		}
	}
	o.SetFinalizers(f)
}

// GetEnv from env get data
func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// StringToInt32 Convert the string value to an int32
func StringToInt32(strValue string, fallback int32) int32 {
	intValue, err := strconv.ParseInt(strValue, 10, 32)
	if err != nil {
		fmt.Println("Error:", err)
		return fallback
	}
	int32Value := int32(intValue)
	return int32Value
}
