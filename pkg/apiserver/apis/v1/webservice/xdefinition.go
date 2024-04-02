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
	"kdp-oam-operator/pkg/apiserver/domain/service"
	"kdp-oam-operator/pkg/utils/log"

	"github.com/emicklei/go-restful/v3"
)

func (c *BigDataClusterWebService) getDefinitionSchema(request *restful.Request, kind, defType string) (*v1dto.XDefinitionBase, error) {
	bdcName := request.PathParameter("bdcName")
	if defType == "" {
		defType = "default"
	}
	def, err := c.XDefinitionService.GetXDefinition(request.Request.Context(), service.DefinitionQueryOption{
		RelatedResourceType: defType,
		RelatedResourceKind: kind,
	}, bdcName)
	if err != nil {
		log.Logger.Errorf("get bigdata cluster definition failure %s", err.Error())
		return nil, err
	}
	defBase, err := assembler.ConvertXDefinitionEntityToDTO(def)
	if err != nil {
		log.Logger.Errorf("convert bigdata cluster to base failure %s", err.Error())
		return nil, err
	}
	return defBase, nil
}
