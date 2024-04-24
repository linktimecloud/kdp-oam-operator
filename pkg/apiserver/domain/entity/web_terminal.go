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

package entity

import (
	"fmt"
	csv1alpha1 "github.com/cloudtty/cloudtty/pkg/apis/cloudshell/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kdp-oam-operator/pkg/utils"
	"time"
)

// WebTerminalEntity web terminal delivery model
type WebTerminalEntity struct {
	Name       string      `json:"name"`
	NameSpace  string      `json:"nameSpace"`
	Phase      string      `json:"phase"`
	AccessUrl  string      `json:"accessUrl"`
	CreateTime metav1.Time `json:"createTime"`
	EndTime    metav1.Time `json:"endTime"`
	Ttl        int32       `json:"ttl"`
}

func Object2WebTerminalEntity(terminal *csv1alpha1.CloudShell) *WebTerminalEntity {
	tLater := terminal.CreationTimestamp.Add(time.Duration(terminal.Spec.Ttl) * time.Second)

	httpType := utils.GetEnv("HTTPTYPE", "http")
	DOMAIN := utils.GetEnv("DOMAIN", "cloudtty.kdp-e2e.io")
	url := fmt.Sprintf("%s://%s%s", httpType, DOMAIN, terminal.Status.AccessURL)
	terminalEntity := &WebTerminalEntity{
		Name:       terminal.Name,
		NameSpace:  terminal.Namespace,
		Phase:      terminal.Status.Phase,
		AccessUrl:  url,
		CreateTime: terminal.CreationTimestamp,
		EndTime:    metav1.NewTime(tLater),
		Ttl:        terminal.Spec.Ttl,
	}

	return terminalEntity
}
