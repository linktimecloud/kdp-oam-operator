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
	"context"
	baseTypes "kdp-oam-operator/pkg/apiserver/apis/base/types"
	"kdp-oam-operator/pkg/apiserver/apis/v1/assembler"
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/apiserver/domain/service"
	"kdp-oam-operator/pkg/apiserver/exception"
	"kdp-oam-operator/pkg/utils/log"
	"strings"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
)

type BigDataClusterWebService struct {
	BigDataClusterService       service.BigDataClusterService
	ApplicationService          service.ApplicationService
	ApplicationResourcesService service.ApplicationResourcesService
	ContextSecretService        service.ContextSecretService
	ContextSettingService       service.ContextSettingService
	XDefinitionService          service.XDefinitionService
}

func NewBigDataClusterWebService(
	bigDataClusterService service.BigDataClusterService,
	applicationService service.ApplicationService,
	applicationResourcesService service.ApplicationResourcesService,
	contextSecretService service.ContextSecretService,
	contextSettingService service.ContextSettingService,
	xDefinitionService service.XDefinitionService) WebService {
	return &BigDataClusterWebService{
		BigDataClusterService:       bigDataClusterService,
		ApplicationService:          applicationService,
		ApplicationResourcesService: applicationResourcesService,
		ContextSecretService:        contextSecretService,
		ContextSettingService:       contextSettingService,
		XDefinitionService:          xDefinitionService,
	}
}

func (c *BigDataClusterWebService) GetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(versionPrefix+"/").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML).
		Doc("api for bigdatacluster manage")

	tags := []string{"bigdatacluster"}

	ws.Route(ws.GET("/bigdataclusters/").To(c.listBigDataClusters).
		Doc("list objects of kind bdc").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter("labelSelector", "A selector to restrict the list of returned objects by their labels. Defaults to everything.").DataType("string").Required(false)).
		//Param(ws.QueryParameter("fieldSelector", "A selector to restrict the list of returned objects by their fields. Defaults to everything.").DataType("string").Required(false)).
		//Param(ws.QueryParameter("pretty", "If 'true', then the output is pretty printed.").DataType("string").Required(false)).
		//Param(ws.QueryParameter("timeoutSeconds", "Timeout for the list/watch call. This limits the duration of the call, regardless of any activity or inactivity.").DataType("integer").Required(false)).
		Writes(v1dto.ListBigDataClustersResponse{}).
		Returns(200, "OK", v1dto.ListBigDataClustersResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/bigdataclusters/{bdcName}").To(c.getBigDataCluster).
		Doc("get the specified bdc").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("bdcName", "name of the bigdata cluster").DataType("string").Required(true)).
		Writes(v1dto.GetBigDataClusterResponse{}).
		Returns(200, "OK", v1dto.GetBigDataClusterResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	applicationTags := []string{"application"}

	ws.Route(ws.GET("/applications").To(c.listApplications).
		Doc("list objects of kind bdc application").
		Metadata(restfulspec.KeyOpenAPITags, applicationTags).
		Param(ws.QueryParameter("bdcName", "name of the bigdata cluster").DataType("string").Required(false)).
		Param(ws.QueryParameter("labelSelector", "A selector to restrict the list of returned objects by their labels. Defaults to everything.").DataType("string").Required(false)).
		Writes(v1dto.ListApplicationsResponse{}).
		Returns(200, "OK", v1dto.ListApplicationsResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/applications/{appName}").To(c.getApplication).
		Doc("read the specified bdc application").
		Metadata(restfulspec.KeyOpenAPITags, applicationTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Writes(v1dto.GetApplicationsResponse{}).
		Returns(200, "OK", v1dto.GetApplicationsResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/bigdataclusters/{bdcName}/applications/definitions/{defType}/schema").To(c.getApplicationDefinitionSchema).
		Doc("read bdc application definition schema").
		Metadata(restfulspec.KeyOpenAPITags, applicationTags).
		Param(ws.PathParameter("bdcName", "name of the bigdata cluster").DataType("string").Required(true)).
		Param(ws.PathParameter("defType", "name of the bdc application definition type").DataType("string").Required(true)).
		Writes(v1dto.GetXDefinitionResponse{}).
		Returns(200, "OK", v1dto.GetXDefinitionResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.POST("/bigdataclusters/{bdcName}/applications").To(c.createApplication).
		Doc("create bdc application").
		Metadata(restfulspec.KeyOpenAPITags, applicationTags).
		Param(ws.PathParameter("bdcName", "name of the bigdata cluster").DataType("string").Required(true)).
		Reads(v1dto.CreateApplicationRequestModel{}).
		Writes(v1dto.GetApplicationsResponse{}).
		Returns(200, "OK", v1dto.GetApplicationsResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}))

	ws.Route(ws.PUT("/applications/{appName}").To(c.updateApplication).
		Doc("update bdc application").
		Metadata(restfulspec.KeyOpenAPITags, applicationTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Reads(v1dto.UpdateApplicationRequestModel{}).
		Writes(v1dto.ApplicationBase{}).
		Returns(200, "OK", baseTypes.HTTPResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}))

	ws.Route(ws.DELETE("/applications/{appName}").To(c.deleteApplication).
		Doc("delete the specified bdc application").
		Metadata(restfulspec.KeyOpenAPITags, applicationTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Returns(200, "OK", baseTypes.HTTPResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}))

	applicationResourcesTags := []string{"application_resources"}

	ws.Route(ws.GET("/applications/{appName}/pods").To(c.getApplicationPods).
		Doc("query application applied pods").
		Metadata(restfulspec.KeyOpenAPITags, applicationResourcesTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Writes(v1dto.ApplicationResourcesListResponse{}).
		Returns(200, "OK", v1dto.ApplicationResourcesListResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/applications/{appName}/pods/{podName}").To(c.getApplicationPodsDetail).
		Doc("query application pods detail info").
		Metadata(restfulspec.KeyOpenAPITags, applicationResourcesTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Param(ws.PathParameter("podName", "name of the bdc application pod").DataType("string").Required(true)).
		//Param(ws.QueryParameter("ql", "query statement").DataType("string")).
		Writes(v1dto.ApplicationResourceResponse{}).
		Returns(200, "OK", v1dto.ApplicationResourceResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.DELETE("/applications/{appName}/pods/{podName}").To(c.deleteApplicationPod).
		Doc("delete the specified bdc application pod").
		Metadata(restfulspec.KeyOpenAPITags, applicationResourcesTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Param(ws.PathParameter("podName", "name of the bdc application pod").DataType("string").Required(true)).
		//Param(ws.QueryParameter("ql", "query statement").DataType("string")).
		Writes(v1dto.ApplicationResourceResponse{}).
		Returns(200, "OK", v1dto.ApplicationResourceResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/applications/{appName}/pods/{podName}/containers/{containerName}/logs").To(c.getApplicationPodLogs).
		Doc("query application applied pods container logs").
		Metadata(restfulspec.KeyOpenAPITags, applicationResourcesTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Param(ws.PathParameter("podName", "name of the bdc application pod").DataType("string").Required(true)).
		Param(ws.PathParameter("containerName", "name of the bdc application pod container").DataType("string").Required(true)).
		Param(ws.QueryParameter("tailLines", "number of tail container logs").DataType("integer").Required(false)).
		Writes(v1dto.ApplicationResourceResponse{}).
		Returns(200, "OK", v1dto.ApplicationResourceResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/applications/{appName}/serviceEndpoints").To(c.getApplicationServiceEndpoints).
		Doc("query application service endpoints").
		Metadata(restfulspec.KeyOpenAPITags, applicationResourcesTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Writes(v1dto.ApplicationResourcesListResponse{}).
		Returns(200, "OK", v1dto.ApplicationResourcesListResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/applications/{appName}/resources/topology").To(c.getApplicationResourceTopology).
		Doc("query applications resource overview").
		Metadata(restfulspec.KeyOpenAPITags, applicationResourcesTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Writes(v1dto.ApplicationResourceResponse{}).
		Returns(200, "OK", v1dto.ApplicationResourceResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/applications/{appName}/resources/detail").To(c.getApplicationResourceDetail).
		Doc("query application resource detail spec").
		Metadata(restfulspec.KeyOpenAPITags, applicationResourcesTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Param(ws.QueryParameter("resNs", "namespace of the kubernetes resource").DataType("string").Required(false)).
		Param(ws.QueryParameter("resName", "name of the kubernetes resource").DataType("string").Required(true)).
		Param(ws.QueryParameter("resKind", "kind of the kubernetes resource").DataType("string").Required(true)).
		Param(ws.QueryParameter("resAPIVersion", "apiVersion of the kubernetes resource").DataType("string").Required(true)).
		Writes(v1dto.ApplicationResourceResponse{}).
		Returns(200, "OK", v1dto.ApplicationResourceResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/applications/{appName}/detail").To(c.detailApplication).
		Doc("read the specified bdc application detail specification").
		Metadata(restfulspec.KeyOpenAPITags, applicationResourcesTags).
		Param(ws.PathParameter("appName", "name of the bdc application").DataType("string").Required(true)).
		Writes(v1dto.ApplicationResourceResponse{}).
		Returns(200, "OK", v1dto.ApplicationResourceResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ctxSecretTags := []string{"contextsecret"}

	ws.Route(ws.GET("/contextsecrets/").To(c.listContextSecrets).
		Doc("list objects of kind bdc context secret").
		Metadata(restfulspec.KeyOpenAPITags, ctxSecretTags).
		Param(ws.QueryParameter("bdcName", "name of the bigdata cluster").DataType("string").Required(false)).
		Param(ws.QueryParameter("labelSelector", "A selector to restrict the list of returned objects by their labels. Defaults to everything.").DataType("string").Required(false)).
		Writes(v1dto.ListContextSecretsResponse{}).
		Returns(200, "OK", v1dto.ListContextSecretsResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/contextsecrets/{name}").To(c.getContextSecret).
		Doc("read the specified bdc context secret").
		Metadata(restfulspec.KeyOpenAPITags, ctxSecretTags).
		Param(ws.PathParameter("name", "name of the bdc context secret").DataType("string").Required(true)).
		Writes(v1dto.GetContextSecretResponse{}).
		Returns(200, "OK", v1dto.GetContextSecretResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/bigdataclusters/{bdcName}/contextsecrets/definitions/{defType}/schema").To(c.getContextSecretDefinitionSchema).
		Doc("read bdc context secret definition schema").
		Metadata(restfulspec.KeyOpenAPITags, ctxSecretTags).
		Param(ws.PathParameter("bdcName", "name of the bigdata cluster").DataType("string").Required(true)).
		Param(ws.PathParameter("defType", "name of the bdc context secret definition type").DataType("string").Required(true)).
		Writes(v1dto.GetContextSecretDefSchemaResponse{}).
		Returns(200, "OK", v1dto.GetContextSecretDefSchemaResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ctxSettingTags := []string{"contextsetting"}

	ws.Route(ws.GET("/contextsettings/").To(c.listContextSettings).
		Doc("list objects of kind bdc context setting").
		Metadata(restfulspec.KeyOpenAPITags, ctxSettingTags).
		Param(ws.QueryParameter("bdcName", "name of the bigdata cluster").DataType("string").Required(false)).
		Param(ws.QueryParameter("labelSelector", "A selector to restrict the list of returned objects by their labels. Defaults to everything.").DataType("string").Required(false)).
		Writes(v1dto.ListContextSettingsResponse{}).
		Returns(200, "OK", v1dto.ListContextSettingsResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/contextsettings/{name}").To(c.getContextSetting).
		Doc("read the specified bdc context setting").
		Metadata(restfulspec.KeyOpenAPITags, ctxSettingTags).
		Param(ws.PathParameter("name", "name of the bdc context setting").DataType("string").Required(true)).
		Writes(v1dto.GetContextSettingResponse{}).
		Returns(200, "OK", v1dto.GetContextSettingResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}).
		Returns(404, "Not found", baseTypes.NotFoundResponse{}))

	ws.Route(ws.GET("/bigdataclusters/{bdcName}/contextsettings/definitions/{defType}/schema").To(c.getContextSettingDefinitionSchema).
		Doc("read bdc context setting definition schema").
		Metadata(restfulspec.KeyOpenAPITags, ctxSettingTags).
		Param(ws.PathParameter("bdcName", "name of the bigdata cluster").DataType("string").Required(true)).
		Param(ws.PathParameter("defType", "name of the bdc context setting definition type").DataType("string").Required(true)).
		Writes(v1dto.GetContextSettingDefSchemaResponse{}).
		Returns(200, "OK", v1dto.GetContextSettingDefSchemaResponse{}).
		Returns(400, "Bad request", baseTypes.BadRequestResponse{}))

	return ws
}

func (c *BigDataClusterWebService) listBigDataClusters(request *restful.Request, response *restful.Response) {
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
	bdcs, err := c.BigDataClusterService.ListBigDataClusters(request.Request.Context(), v1dto.ListOptions{Labels: labels})
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	var listRtn []*v1dto.BigDataClusterBase
	for _, item := range bdcs {
		bdcBase, err := assembler.ConvertBigDataClusterEntityToDTO(item)
		if err != nil {
			log.Logger.Errorf("convert bigdata cluster to base failure %s", err.Error())
			continue
		}
		listRtn = append(listRtn, bdcBase)
	}
	if err = response.WriteEntity(v1dto.ListBigDataClustersResponse{
		Data:    listRtn,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getBigDataCluster(request *restful.Request, response *restful.Response) {
	bdcName := request.PathParameter("bdcName")
	bdc, err := c.BigDataClusterService.GetBigDataCluster(request.Request.Context(), bdcName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrBigDataClusterNotFound)
		return
	}
	bdcBase, err := assembler.ConvertBigDataClusterEntityToDTO(bdc)
	if err != nil {
		log.Logger.Errorf("convert bigdata cluster to base failure %s", err.Error())
		exception.ReturnError(request, response, err)
		return
	}
	if err := response.WriteEntity(v1dto.GetBigDataClusterResponse{
		Data:    bdcBase,
		Message: "success",
		Status:  0,
	}); err != nil {
		return
	}
}

func (c *BigDataClusterWebService) bigDataClusterMetaCacheParse(bdcName string) (*v1dto.BigDataClusterBase, error) {
	bdc, err := c.BigDataClusterService.GetBigDataCluster(context.Background(), bdcName)
	if err != nil {
		log.Logger.Errorw("get bigdata cluster failure", "error", err)
		return nil, err
	}
	bdcBase, err := assembler.ConvertBigDataClusterEntityToDTO(bdc)
	if err != nil {
		klog.Errorf("convert bigdata cluster to base failure %s", err.Error())
		return nil, err
	}
	return bdcBase, nil
}
