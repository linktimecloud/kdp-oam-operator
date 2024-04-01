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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestIsEmptyDir tests the IsEmptyDir function.
func TestIsEmptyDir(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	emptyDir := filepath.Join(tmpDir, "emptydir")
	err = os.Mkdir(emptyDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := IsEmptyDir(emptyDir); err != nil {
		t.Errorf("Expected true, got false")
	} else {
		t.Logf("Empty directory: %s", emptyDir)
	}

	nonEmptyDir := filepath.Join(tmpDir, "nonemptydir")
	err = os.Mkdir(nonEmptyDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(nonEmptyDir, "file.txt")
	err = os.WriteFile(file, []byte("test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := IsEmptyDir(emptyDir); err != nil {
		t.Errorf("Expected true, got false")
	} else {
		t.Logf("Empty directory: %s", emptyDir)
	}

	nonDir := filepath.Join(tmpDir, "file.txt")
	err = os.WriteFile(nonDir, []byte("test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := IsEmptyDir(emptyDir); err != nil {
		t.Errorf("Expected true, got false")
	} else {
		t.Logf("Empty directory: %s", emptyDir)
	}

	mixedDir := filepath.Join(tmpDir, "mixeddir")
	err = os.Mkdir(mixedDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat(mixedDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := IsEmptyDir(emptyDir); err != nil {
		t.Errorf("Expected true, got false")
	} else {
		t.Logf("Empty directory: %s", emptyDir)
	}

	nestedDir := filepath.Join(tmpDir, "nesteddir")
	err = os.Mkdir(nestedDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat(nestedDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := IsEmptyDir(emptyDir); err != nil {
		t.Errorf("Expected true, got false")
	} else {
		t.Logf("Empty directory: %s", emptyDir)
	}
}

// TestReadContent tests the ReadContent function
func TestReadContent(t *testing.T) {
	// Test case 1: File exists
	path := "path/to/file.txt"
	expectedOutput := "Hello, World!"
	actualOutput := ReadContent(path)
	if actualOutput != expectedOutput {
		t.Logf("ReadContent(%s) = %s; want %s", path, actualOutput, expectedOutput)
	}

	// Test case 2: File does not exist
	path = "path/to/nonexistent/file.txt"
	expectedOutput = ""
	actualOutput = ReadContent(path)
	if actualOutput != expectedOutput {
		t.Logf("ReadContent(%s) = %s; want %s", path, actualOutput, expectedOutput)
	}

	// Test case 3: File is a directory
	path = "path/to/directory"
	expectedOutput = ""
	actualOutput = ReadContent(path)
	if actualOutput != expectedOutput {
		t.Logf("ReadContent(%s) = %s; want %s", path, actualOutput, expectedOutput)
	}
}

func TestPrettyYAMLMarshal(t *testing.T) {
	type args struct {
		obj map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"test-1", args{map[string]interface{}{"property": ""}}, "property: \"\"\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrettyYAMLMarshal(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyYAMLMarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PrettyYAMLMarshal() got = %v, want %v", got, tt.want)
			}
		})
	}
}
