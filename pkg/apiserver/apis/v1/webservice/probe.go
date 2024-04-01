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
	baseTypes "kdp-oam-operator/pkg/apiserver/apis/base/types"

	"github.com/emicklei/go-restful/v3"
)

type ProbeService struct {
}

func NewProbeService() WebService {
	return &ProbeService{}
}

func (p *ProbeService) GetWebService() *restful.WebService {

	ws := new(restful.WebService)
	ws.Path("/").
		Consumes(restful.MIME_JSON, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_JSON)
	ws.Route(ws.GET("/readyz").To(p.probe))
	ws.Route(ws.GET("/healthz").To(p.probe))
	return ws
}

func (p *ProbeService) probe(request *restful.Request, response *restful.Response) {
	if err := response.WriteEntity(baseTypes.HTTPResponse{}); err != nil {
		return
	}
}
