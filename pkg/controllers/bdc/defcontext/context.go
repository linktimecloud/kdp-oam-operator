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

package defcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

const (
	Context = "context"
	// ContextName is the name of context
	ContextName = "name"
	// ContextBDCName is the appName of context
	ContextBDCName   = "bdcName"
	ContextBDCLabels = "bdcLabels"
	// ContextBDCAnnotations is the annotations of bdc of context
	ContextBDCAnnotations = "bdcAnnotations"
	// ContextNamespace is the namespace of the bdc
	ContextNamespace = "namespace"

	AppUuid = "app_uuid"
	Group   = "group"
	Bdc     = "bdc"
)

type ContextData struct {
	Namespace      string
	Name           string
	Cluster        string
	BDCName        string
	BDCLabels      map[string]string
	BDCAnnotations map[string]string
	Ctx            context.Context
	data           map[string]interface{}
}

// NewBDCContext creates a new bdc context
func NewBDCContext(cd ContextData) ContextData {
	cd.PushData(ContextName, cd.Name)
	cd.PushData(ContextNamespace, cd.Namespace)
	cd.PushData(ContextBDCName, cd.BDCName)
	cd.PushData(ContextBDCLabels, cd.BDCLabels)
	cd.PushData(ContextBDCAnnotations, cd.BDCAnnotations)
	return cd
}

func (ctx *ContextData) PushData(key string, data interface{}) {
	if ctx.data == nil {
		ctx.data = map[string]interface{}{key: data}
		return
	}
	ctx.data[key] = data
}

// GetData get data from context
func (ctx *ContextData) GetData(key string) interface{} {
	return ctx.data[key]
}

func (ctx *ContextData) BaseContextLabels() map[string]string {
	return map[string]string{
		ContextName: fmt.Sprint(ctx.GetData(ContextName)),
	}
}

func (ctx *ContextData) BaseContextFile() (string, error) {
	var buff string

	if ctx.data != nil {
		d, err := json.Marshal(ctx.data)
		if err != nil {
			return "", err
		}
		buff += fmt.Sprintf("\n %s", structMarshal(string(d)))
	}

	return fmt.Sprintf("%s: %s", Context, structMarshal(buff)), nil
}

func structMarshal(v string) string {
	skip := false
	v = strings.TrimFunc(v, func(r rune) bool {
		if !skip {
			if unicode.IsSpace(r) {
				return true
			}
			skip = true

		}
		return false
	})

	if strings.HasPrefix(v, "{") {
		return v
	}
	return fmt.Sprintf("{%s}", v)
}
