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

	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
	"kdp-oam-operator/pkg/utils"
)

func (c *BigDataClusterWebService) createPodTerminal(request *restful.Request, response *restful.Response) {
	kubeConfigSecretName := "pod-terminal-secret"
	podName := request.PathParameter("podName")
	containerName := request.PathParameter("containerName")
	TerminalName := podName + "-" + containerName + "-exec"

	appName := request.PathParameter("appName")
	app, err := c.applicationMetaCacheParse(appName)
	if err != nil {
		exception.ReturnError(request, response, exception.ErrApplicationNotFound)
		return
	}
	namespace := utils.GetEnv("NAMESPACE", "default")

	// If the length is greater than 40, substitute is used ShortHashID
	if len(TerminalName) > 40 {
		TerminalNameShortHashID, err := utils.GenerateShortHashID(16, app.AppRuntimeNs, podName, containerName)
		if err != nil {
			exception.ReturnError(request, response, err)
			return
		}
		TerminalName = TerminalNameShortHashID
	}

	// create pod exec cloud shell  kubeConfigSecretName, TerminalName, TerminalNameSpace, podName, containerName
	terminal, err := c.WebTerminalService.OpenTerminal(
		request.Request.Context(), kubeConfigSecretName, TerminalName, namespace, app.AppRuntimeNs, podName, containerName)
	if err != nil {
		exception.ReturnError(request, response, err)
		return
	}

	TerminalBase, err := assembler.ConvertWebTerminalEntityToDTO(terminal)
	if err != nil {
		klog.Errorf("convert terminal to base failure %s", err.Error())
		exception.ReturnError(request, response, err)
		return
	}

	if err := response.WriteEntity(v1dto.WebTerminalResponse{
		Data:    TerminalBase,
		Message: "success",
		Status:  0,
	}); err != nil {
		return
	}
}

func (c *BigDataClusterWebService) createGeneralTerminal(request *restful.Request, response *restful.Response) {
	kubeConfigSecretName := "general-terminal-secret"
	TerminalName := "general-exec"
	namespace := utils.GetEnv("NAMESPACE", "default")
	terminal, err := c.WebTerminalService.OpenTerminal(
		request.Request.Context(), kubeConfigSecretName, TerminalName, namespace, "", "", "")
	if err != nil {
		switch err.Error() {
		case "ingressCheckFailed":
			exception.ReturnError(request, response, exception.ErrIngressCheckFailed)
			return
		case "createTerminalFailed":
			exception.ReturnError(request, response, exception.ErrCreateTerminalFailed)
			return
		case "terminalNotFound":
			exception.ReturnError(request, response, exception.ErrTerminalNotFound)
			return
		case "obtainLimitRetry":
			exception.ReturnError(request, response, exception.ErrObtainLimitTry)
			return
		default:
			exception.ReturnError(request, response, err)
			return
		}
	}

	TerminalBase, err := assembler.ConvertWebTerminalEntityToDTO(terminal)
	if err != nil {
		klog.Errorf("convert terminal to base failure %s", err.Error())
		exception.ReturnError(request, response, err)
		return
	}

	if err := response.WriteEntity(v1dto.WebTerminalResponse{
		Data:    TerminalBase,
		Message: "success",
		Status:  0,
	}); err != nil {
		return
	}
}
