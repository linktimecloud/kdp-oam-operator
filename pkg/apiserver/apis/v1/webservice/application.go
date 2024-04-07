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
	"kdp-oam-operator/pkg/apiserver/apis/v1/assembler"
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/apiserver/exception"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	pkgutils "kdp-oam-operator/pkg/utils"
	"kdp-oam-operator/pkg/utils/log"
	"strings"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
)

func (c *BigDataClusterWebService) applicationMetaCacheParse(appName string) (*v1dto.ApplicationBase, error) {
	bdcApp, err := c.ApplicationService.GetApplication(context.Background(), appName)
	if err != nil {
		log.Logger.Errorw("get bigdata cluster failure", "error", err)
		return nil, err
	}
	bdcAppBase, err := assembler.ConvertApplicationEntityToDTO(bdcApp)
	if err != nil {
		klog.Errorf("convert bigdata cluster to base failure %s", err.Error())
		return nil, err
	}
	return bdcAppBase, nil
}

func (c *BigDataClusterWebService) listApplications(request *restful.Request, response *restful.Response) {
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
	bdc, err := c.bigDataClusterMetaCacheParse(bdcName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrBigDataClusterNotFound)
		return
	}
	if bdc.Name != "" {
		labels[constants.AnnotationBDCName] = bdc.Name
	}
	apps, err := c.ApplicationService.ListApplications(request.Request.Context(), v1dto.ListOptions{Labels: labels})
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	listRtn := make([]*v1dto.ApplicationBase, 0)
	for _, item := range apps {
		appBase, err := assembler.ConvertApplicationEntityToDTO(item)
		if err != nil {
			klog.Errorf("convert bigdata cluster to base failure %s", err.Error())
			continue
		}
		listRtn = append(listRtn, appBase)
	}
	if err := response.WriteEntity(v1dto.ListApplicationsResponse{
		Data:    listRtn,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getApplication(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	app, err := c.ApplicationService.GetApplication(request.Request.Context(), appName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrApplicationNotFound)
		return
	}
	appBase, err := assembler.ConvertApplicationEntityToDTO(app)
	if err != nil {
		klog.Errorf("convert bigdata cluster to base failure %s", err.Error())
		exception.ReturnError(request, response, err)
		return
	}
	if err := response.WriteEntity(v1dto.GetApplicationsResponse{
		Data:    appBase,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) detailApplication(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	app, err := c.ApplicationService.DetailApplication(request.Request.Context(), appName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrApplicationNotFound)
		return
	}
	resRaw := pkgutils.Object2RawExtension(app)
	if err := response.WriteEntity(v1dto.ApplicationRawResponse{
		Data:    resRaw,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) createApplication(request *restful.Request, response *restful.Response) {
	bdcName := request.PathParameter("bdcName")
	var createReq v1dto.CreateApplicationRequest
	if err := request.ReadEntity(&createReq); err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	if err := validate.Struct(&createReq); err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	bdc, err := c.bigDataClusterMetaCacheParse(bdcName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrBigDataClusterNotFound)
		return
	}
	createReq.BDC = bdc
	appBase, err := c.ApplicationService.CreateApplication(request.Request.Context(), createReq)
	if err != nil {
		klog.Errorf("create application failure %s", err.Error())
		exception.ReturnError(request, response, err)
		return
	}
	if err := response.WriteEntity(v1dto.GetApplicationsResponse{
		Data:    appBase,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) updateApplication(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	app, err := c.applicationMetaCacheParse(appName)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			exception.ReturnError(request, response, exception.ErrApplicationNotFound)
			return
		}
		exception.ReturnError(request, response, err)
		return
	}
	var updateReq v1dto.UpdateApplicationRequest
	if err := request.ReadEntity(&updateReq); err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	if err := validate.Struct(&updateReq); err != nil {
		exception.ReturnError(request, response, err)
		return
	}
	updateReq.AppName = app.Name
	updateReq.BDC = app.BDC
	appBase, err := c.ApplicationService.UpdateApplication(request.Request.Context(), updateReq)
	if err != nil {
		klog.Errorf("create application failure %s", err.Error())
		exception.ReturnError(request, response, err)
		return
	}
	if err := response.WriteEntity(v1dto.GetApplicationsResponse{
		Data:    appBase,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) deleteApplication(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	err := c.ApplicationService.DeleteApplication(request.Request.Context(), appName)
	if err != nil {
		klog.Errorf("create application failure %s", err.Error())
		exception.ReturnError(request, response, err)
		return
	}
	if err := response.WriteEntity(v1dto.GetApplicationsResponse{
		Data:    nil,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}

func (c *BigDataClusterWebService) getApplicationDefinitionSchema(request *restful.Request, response *restful.Response) {
	defType := request.PathParameter("defType")
	defBase, err := c.getDefinitionSchema(request, "Application", defType)
	if err != nil {
		klog.Errorf("query definition schema with error %s", err.Error())
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

func (c *BigDataClusterWebService) deleteApplicationPod(request *restful.Request, response *restful.Response) {
	appName := request.PathParameter("appName")
	app, err := c.applicationMetaCacheParse(appName)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			exception.ReturnError(request, response, exception.ErrApplicationNotFound)
			return
		}
		exception.ReturnError(request, response, err)
		return
	}
	podName := request.PathParameter("podName")
	err = c.ApplicationService.DeleteApplicationPod(request.Request.Context(), app.AppRuntimeNs, podName)
	if err != nil {
		klog.Errorf("create application failure %s", err.Error())
		exception.ReturnError(request, response, err)
		return
	}
	if err := response.WriteEntity(v1dto.GetApplicationsResponse{
		Data:    nil,
		Message: "success",
		Status:  0,
	}); err != nil {
		// bcode.ReturnError(req, res, err)
		return
	}
}
