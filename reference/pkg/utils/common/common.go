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

package common

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/reference/pkg/cue/model/value"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/openapi"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"
)

var (
	// Scheme defines the default kdp schema
	Scheme = k8sruntime.NewScheme()
)

// CreateCustomNamespace display the create namespace message
const CreateCustomNamespace = "create new namespace"

func init() {
	_ = clientgoscheme.AddToScheme(Scheme)
	// +kubebuilder:scaffold:scheme
}

// HTTPOption define the https options
type HTTPOption struct {
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	CaFile          string `json:"caFile,omitempty"`
	CertFile        string `json:"certFile,omitempty"`
	KeyFile         string `json:"keyFile,omitempty"`
	InsecureSkipTLS bool   `json:"insecureSkipTLS,omitempty"`
}

// InitBaseRestConfig will return reset config for create controller runtime client
func InitBaseRestConfig() (Args, error) {
	args := Args{
		Schema: Scheme,
	}
	_, err := args.GetConfig()
	if err != nil && os.Getenv("IGNORE_KUBE_CONFIG") != "true" {
		fmt.Println("get kubeConfig err", err)
		os.Exit(1)
	} else if err != nil {
		return Args{}, err
	}
	return args, nil
}

// HTTPGetResponse use HTTP option and default client to send request and get raw response
func HTTPGetResponse(ctx context.Context, url string, opts *HTTPOption) (*http.Response, error) {
	// Change NewRequest to NewRequestWithContext and pass context it
	if _, err := neturl.ParseRequestURI(url); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{}
	if opts != nil && len(opts.Username) != 0 && len(opts.Password) != 0 {
		req.SetBasicAuth(opts.Username, opts.Password)
	}
	if opts != nil && opts.InsecureSkipTLS {
		httpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}} // nolint
	}
	// if specify the caFile, we cannot re-use the default httpClient, so create a new one.
	if opts != nil && (len(opts.CaFile) != 0 || len(opts.KeyFile) != 0 || len(opts.CertFile) != 0) {
		// must set MinVersion of TLS, otherwise will report GoSec error G402
		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
		tr := http.Transport{}
		if len(opts.CaFile) != 0 {
			c := x509.NewCertPool()
			if !(c.AppendCertsFromPEM([]byte(opts.CaFile))) {
				return nil, fmt.Errorf("failed to append certificates")
			}
			tlsConfig.RootCAs = c
		}
		if len(opts.CertFile) != 0 && len(opts.KeyFile) != 0 {
			cert, err := tls.X509KeyPair([]byte(opts.CertFile), []byte(opts.KeyFile))
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
		}
		tr.TLSClientConfig = tlsConfig
		defer tr.CloseIdleConnections()
		httpClient.Transport = &tr
	}
	return httpClient.Do(req)
}

// HTTPGetWithOption use HTTP option and default client to send get request
func HTTPGetWithOption(ctx context.Context, url string, opts *HTTPOption) ([]byte, error) {
	resp, err := HTTPGetResponse(ctx, url, opts)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// HTTPGetKubernetesObjects use HTTP requests to load resources from remote url
func HTTPGetKubernetesObjects(ctx context.Context, url string) ([]*unstructured.Unstructured, error) {
	resp, err := HTTPGetResponse(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer resp.Body.Close()
	decoder := yamlv3.NewDecoder(resp.Body)
	var uns []*unstructured.Unstructured
	for {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{}}
		if err := decoder.Decode(obj.Object); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed to decode object: %w", err)
		}
		uns = append(uns, obj)
	}
	return uns, nil
}

// GenOpenAPI generates OpenAPI json schema from cue.Instance
func GenOpenAPI(val *value.Value) (b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid cue definition to generate open api: %v", r)
			debug.PrintStack()
			return
		}
	}()
	if val.CueValue().Err() != nil {
		return nil, val.CueValue().Err()
	}
	paramOnlyVal, err := RefineParameterValue(val)
	if err != nil {
		return nil, err
	}
	defaultConfig := &openapi.Config{ExpandReferences: true}
	b, err = openapi.Gen(paramOnlyVal, defaultConfig)
	if err != nil {
		return nil, err
	}
	var out = &bytes.Buffer{}
	_ = json.Indent(out, b, "", "   ")
	return out.Bytes(), nil
}

// GenOpenAPIWithCueX generates OpenAPI json schema from cue.Instance
func GenOpenAPIWithCueX(val cue.Value) (b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid cue definition to generate open api: %v", r)
			debug.PrintStack()
			return
		}
	}()
	if val.Err() != nil {
		return nil, val.Err()
	}
	paramOnlyVal := FillParameterDefinitionFieldIfNotExist(val)
	defaultConfig := &openapi.Config{ExpandReferences: true}
	b, err = openapi.Gen(paramOnlyVal, defaultConfig)
	if err != nil {
		return nil, err
	}
	var out = &bytes.Buffer{}
	_ = json.Indent(out, b, "", "   ")
	return out.Bytes(), nil
}

// RefineParameterValue refines cue value to merely include `parameter` identifier
func RefineParameterValue(val *value.Value) (cue.Value, error) {
	defaultValue := cuecontext.New().CompileString("#parameter: {}")
	parameterPath := cue.MakePath(cue.Def("parameter"))
	v, err := val.MakeValue("{}")
	if err != nil {
		return defaultValue, err
	}
	paramVal, err := val.LookupValue("parameter")
	if err != nil {
		// nolint:nilerr
		return defaultValue, nil
	}
	switch k := paramVal.CueValue().IncompleteKind(); k {
	case cue.BottomKind:
		return defaultValue, nil
	default:
		paramOnlyVal := v.CueValue().FillPath(parameterPath, paramVal.CueValue())
		return paramOnlyVal, nil
	}
}

// FillParameterDefinitionFieldIfNotExist refines cue value to merely include `parameter` identifier
func FillParameterDefinitionFieldIfNotExist(val cue.Value) cue.Value {
	defaultValue := cuecontext.New().CompileString("#parameter: {}")
	defPath := cue.ParsePath("#" + "parameter")
	if paramVal := val.LookupPath(cue.ParsePath("parameter")); paramVal.Exists() {
		if paramVal.IncompleteKind() == cue.BottomKind {
			return defaultValue
		}
		paramOnlyVal := val.Context().CompileString("{}").FillPath(defPath, paramVal)
		return paramOnlyVal
	}
	return defaultValue
}

// RealtimePrintCommandOutput prints command output in real time
// If logFile is "", it will prints the stdout, or it will write to local file
func RealtimePrintCommandOutput(cmd *exec.Cmd, logFile string) error {
	var writer io.Writer
	if logFile == "" {
		writer = io.MultiWriter(os.Stdout)
	} else {
		if _, err := os.Stat(filepath.Dir(logFile)); err != nil {
			return err
		}
		f, err := os.Create(filepath.Clean(logFile))
		if err != nil {
			return err
		}
		writer = io.MultiWriter(f)
	}
	cmd.Stdout = writer
	cmd.Stderr = writer
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// ReadYamlToObject will read a yaml K8s object to runtime.Object
func ReadYamlToObject(path string, object k8sruntime.Object) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, object)
}

// GenerateUnstructuredObj generate UnstructuredObj
func GenerateUnstructuredObj(name, ns string, gvk schema.GroupVersionKind) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	u.SetName(name)
	u.SetNamespace(ns)
	return u
}

// SetSpecObjIntoUnstructuredObj set UnstructuredObj spec field
func SetSpecObjIntoUnstructuredObj(spec interface{}, u *unstructured.Unstructured) error {
	bts, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	data := make(map[string]interface{})
	if err := json.Unmarshal(bts, &data); err != nil {
		return err
	}
	_ = unstructured.SetNestedMap(u.Object, data, "spec")
	return nil
}

// NewK8sClient init a local k8s clien
func NewK8sClient() (client.Client, error) {
	conf, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	scheme := k8sruntime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	k8sClient, err := client.New(conf, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}

func init() {
	v1alpha1.AddToScheme(Scheme)
}
