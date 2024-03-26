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

package service

import (
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/apiserver/exception"
	pkgutils "kdp-oam-operator/pkg/utils"
	"strconv"

	"github.com/emicklei/go-restful/v3"
)

func (c *BigDataClusterWebService) getApplicationPods(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	app, err := c.applicationMetaCacheParse(appName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrApplicationNotFound)
		return
	}

	appPods, err := c.ApplicationResourcesService.GetApplicationResourcesPods(request.Request.Context(), app.AppRuntimeNs, app.AppRuntimeName)
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	resRaw := pkgutils.Object2RawExtension(appPods["podList"])
	if err = response.WriteEntity(v1dto.ApplicationResourceRawResponse{
		Data:    resRaw,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getApplicationPodsDetail(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	app, err := c.applicationMetaCacheParse(appName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrApplicationNotFound)
		return
	}

	podName := request.PathParameter("podName")

	appPodsDetail, err := c.ApplicationResourcesService.GetApplicationResourcesPodsDetail(request.Request.Context(), app.AppRuntimeNs, podName)
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	appPodsRaw := pkgutils.Object2RawExtension(appPodsDetail)
	if err = response.WriteEntity(v1dto.ApplicationResourceRawResponse{
		Data:    appPodsRaw,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getApplicationPodLogs(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	app, err := c.applicationMetaCacheParse(appName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrApplicationNotFound)
		return
	}
	podName := request.PathParameter("podName")
	containerName := request.PathParameter("containerName")
	tailLines, _ := strconv.Atoi(request.QueryParameter("tailLines"))
	if tailLines <= 0 {
		tailLines = 100
	}

	appPodLogs, err := c.ApplicationResourcesService.GetApplicationResourcesPodLogs(request.Request.Context(), app.AppRuntimeNs, podName, containerName, int32(tailLines))
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	resRaw := pkgutils.Object2RawExtension(appPodLogs)
	if err := response.WriteEntity(v1dto.ApplicationResourceRawResponse{
		Data:    resRaw,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getApplicationServiceEndpoints(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	app, err := c.applicationMetaCacheParse(appName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrApplicationNotFound)
		return
	}

	appServiceEndpoints, err := c.ApplicationResourcesService.GetApplicationServiceEndpoints(request.Request.Context(), app.AppRuntimeNs, app.AppRuntimeName)
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	resRaw := pkgutils.Object2RawExtension(appServiceEndpoints["endpoints"])
	if err = response.WriteEntity(v1dto.ApplicationResourceRawResponse{
		Data:    resRaw,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getApplicationAppliedResources(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	app, err := c.applicationMetaCacheParse(appName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrApplicationNotFound)
		return
	}

	appAppliedResources, err := c.ApplicationResourcesService.GetApplicationAppliedResources(request.Request.Context(), app.AppRuntimeNs, app.AppRuntimeName)
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	resRaw := pkgutils.Object2RawExtension(appAppliedResources["resources"])
	if err = response.WriteEntity(v1dto.ApplicationResourceRawResponse{
		Data:    resRaw,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getApplicationResourceTopology(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	app, err := c.applicationMetaCacheParse(appName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrApplicationNotFound)
		return
	}

	appResourcesTopology, err := c.ApplicationResourcesService.GetApplicationResourcesTopology(request.Request.Context(), app.AppRuntimeNs, app.AppRuntimeName)
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	resRaw := pkgutils.Object2RawExtension(appResourcesTopology["resources"])
	if err = response.WriteEntity(v1dto.ApplicationResourceRawResponse{
		Data:    resRaw,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getApplicationResourceDetail(request *restful.Request, response *restful.Response) {
	resNs := request.QueryParameter("resNs")
	resName := request.QueryParameter("resName")
	resKind := request.QueryParameter("resKind")
	resAPIVersion := request.QueryParameter("resAPIVersion")

	appResourcesDetailSpec, err := c.ApplicationResourcesService.GetApplicationResourcesDetail(request.Request.Context(), resNs, resName, resKind, resAPIVersion)
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	resRaw := pkgutils.Object2RawExtension(appResourcesDetailSpec["resource"])
	if err = response.WriteEntity(v1dto.ApplicationResourceRawResponse{
		Data:    resRaw,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}
