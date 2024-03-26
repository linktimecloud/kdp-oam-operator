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

package cue

import (
	"cuelang.org/go/cue"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

func ValidateAndCompile(v cue.Value) ([]byte, error) {
	if err := v.Err(); err != nil {
		return nil, err
	}
	// compiled object should be final and concrete value
	if err := v.Validate(cue.Concrete(true), cue.Final()); err != nil {
		return nil, err
	}
	return v.MarshalJSON()
}

// Unstructured convert cue values to unstructured.Unstructured
func Unstructured(v cue.Value) (*unstructured.Unstructured, error) {
	jsonv, err := ValidateAndCompile(v)
	if err != nil {
		klog.ErrorS(err, "failed to validate and compile cue value", "Definition", v)
		return nil, errors.Wrap(err, "failed to have the unstructured")
	}
	o := &unstructured.Unstructured{}
	if err := o.UnmarshalJSON(jsonv); err != nil {
		return nil, err
	}
	return o, nil
}
