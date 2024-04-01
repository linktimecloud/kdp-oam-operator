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
	"kdp-oam-operator/pkg/apiserver/domain/service"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-playground/validator/v10"
)

// VersionPrefix API version prefix.
var versionPrefix = "/api/v1"

var validate = validator.New()

// WebService interface
type WebService interface {
	GetWebService() *restful.WebService
}

var registeredWebService []WebService

// RegisterWebService register webservice
func RegisterWebService(ws WebService) {
	registeredWebService = append(registeredWebService, ws)
}

// GetRegisteredWebService return registeredWebService
func GetRegisteredWebService() []WebService {
	return registeredWebService
}

// Init all webservice, pass in the required parameter object.
func Init() {
	// init domain service instance
	bigDataClusterService := service.NewBigDataClusterService()
	applicationService := service.NewApplicationService()
	applicationResourcesService := service.NewApplicationResourcesService()
	contextSecretService := service.NewContextSecretService()
	contextSettingService := service.NewContextSettingService()
	xDefinitionService := service.NewXDefinitionService()

	// register webservice
	RegisterWebService(NewBigDataClusterWebService(bigDataClusterService, applicationService,
		applicationResourcesService, contextSecretService, contextSettingService, xDefinitionService))
	RegisterWebService(NewProbeService())
}
