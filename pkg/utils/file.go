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
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
)

func IsEmptyDir(path string) (bool, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return false, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	// Read just one file in the dir (just read names, which is faster)
	_, err = f.Readdirnames(1)
	// If the error is EOF, the dir is empty
	if errors.Is(err, io.EOF) {
		return true, nil
	}

	return false, err
}

func ReadContent(path string) string {
	oBytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		fmt.Printf(err.Error())
		return ""
	}
	return string(oBytes)
}

func PrettyYAMLMarshal(obj map[string]interface{}) (string, error) {
	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	encoder.SetIndent(2)
	err := encoder.Encode(&obj)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
