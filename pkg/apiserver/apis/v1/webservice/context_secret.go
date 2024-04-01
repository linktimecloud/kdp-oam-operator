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

package webservice

import (
	"kdp-oam-operator/pkg/apiserver/apis/v1/assembler"
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/apiserver/exception"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/utils/log"
	"strings"

	"github.com/emicklei/go-restful/v3"
)

func (c *BigDataClusterWebService) listContextSecrets(request *restful.Request, response *restful.Response) {
	bdcName := request.QueryParameter("bdcName")
	labels := map[string]string{}
	if request.QueryParameter("labelSelector") != "" {
		allLabels := strings.Split(request.QueryParameter("labelSelector"), ",")
		for _, label := range allLabels {
			kv := strings.Split(label, "=")
			if len(kv) == 2 {
				labels[kv[0]] = kv[1]
			}
		}
	}
	if bdcName != "" {
		labels[constants.AnnotationBDCName] = bdcName
	}
	ctxSecrets, err := c.ContextSecretService.ListContextSecrets(request.Request.Context(), v1dto.ListOptions{Labels: labels})
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	var listRtn []*v1dto.ContextSecretBase
	for _, item := range ctxSecrets {
		ctxSecretBase, err := assembler.ConvertContextSecretEntityToDTO(item)
		if err != nil {
			log.Logger.Errorf("convert bigdata cluster to base failure %s", err.Error())
			continue
		}
		listRtn = append(listRtn, ctxSecretBase)
	}
	if err := response.WriteEntity(v1dto.ListContextSecretsResponse{
		Data:    listRtn,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getContextSecret(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	ctxSecret, err := c.ContextSecretService.GetContextSecret(request.Request.Context(), name)
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	ctxSecretBase, err := assembler.ConvertContextSecretEntityToDTO(ctxSecret)
	if err != nil {
		log.Logger.Errorf("convert bigdata cluster to base failure %s", err.Error())
		exception.ReturnError(request, response, err)
		return
	}
	if err := response.WriteEntity(v1dto.GetContextSecretResponse{
		Data:    ctxSecretBase,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getContextSecretDefinitionSchema(request *restful.Request, response *restful.Response) {
	defType := request.PathParameter("defType")
	defBase, err := c.getDefinitionSchema(request, "ContextSecret", defType)
	if err != nil {
		log.Logger.Errorf("query definition schema with error %s", err.Error())
		exception.ReturnError(request, response, err)
		return
	}
	if err := response.WriteEntity(v1dto.GetXDefinitionResponse{
		Data:    defBase,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}
