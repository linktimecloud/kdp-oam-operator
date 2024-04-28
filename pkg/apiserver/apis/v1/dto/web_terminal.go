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

package dto

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type TerminalBase struct {
	Name       string      `json:"name"`
	NameSpace  string      `json:"nameSpace"`
	Phase      string      `json:"phase"`
	AccessUrl  string      `json:"accessUrl"`
	CreateTime metav1.Time `json:"createTime"`
	EndTime    metav1.Time `json:"endTime"`
	Ttl        int64       `json:"ttl"`
}

type WebTerminalResponse struct {
	Data    *TerminalBase `json:"data"`
	Message string        `json:"message"`
	Status  int           `json:"status"`
}

type ExtractionRules struct {
	Phase     []string `json:"phase"`
	AccessUrl []string `json:"accessUrl"`
	Ttl       []string `json:"ttl"`
}
