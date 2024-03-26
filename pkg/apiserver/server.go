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

package apiserver

import (
	webservice "kdp-oam-operator/pkg/apiserver/apis/v1/webservice"
	"kdp-oam-operator/pkg/apiserver/config"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"kdp-oam-operator/pkg/apiserver/utils"
	"kdp-oam-operator/pkg/utils/log"
	"os"
	"path/filepath"
	"time"

	"context"
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
)

var _ APIServer = &RestServer{}

// APIServer interface for call api server
type APIServer interface {
	Run(context.Context) error
	BuildRestfulConfig() restfulspec.Config
}

type RestServer struct {
	webContainer *restful.Container
	cfg          config.APIServerConfig
}

// New create restServer with config data
func New(cfg config.APIServerConfig) (a APIServer, err error) {

	s := &RestServer{
		webContainer: restful.NewContainer(),
		cfg:          cfg,
	}
	return s, nil
}

func (s *RestServer) Run(ctx context.Context) error {
	err := clients.SetKubeConfig(s.cfg)
	if err != nil {
		return err
	}
	s.BuildRestfulConfig()
	return s.startHTTP(ctx)
}

func (s *RestServer) BuildRestfulConfig() restfulspec.Config {
	webservice.Init()
	/* **************************************************************  */
	/* *************       Open API Route Group     *****************  */
	/* **************************************************************  */

	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{},
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		CookiesAllowed: true,
		Container:      s.webContainer}
	s.webContainer.Filter(cors.Filter)

	// Add container filter to respond to OPTIONS
	s.webContainer.Filter(s.webContainer.OPTIONSFilter)

	// Add request log
	s.webContainer.Filter(s.requestLog)

	// Register all custom webservice
	for _, handler := range webservice.GetRegisteredWebService() {
		s.webContainer.Add(handler.GetWebService())
	}

	restFulSpecConfig := restfulspec.Config{
		WebServices:                   s.webContainer.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}
	s.webContainer.Add(restfulspec.NewOpenAPIService(restFulSpecConfig))

	if s.cfg.SwaggerDocEnabled {
		currentDir, _ := os.Getwd()
		swdist := filepath.Join(currentDir, `docs/apidoc/swagger-ui`)
		log.Logger.Infof(swdist)

		s.webContainer.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir(swdist))))
	}

	return restFulSpecConfig
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "kdp-oam-apiserver api doc",
			Description: "kdp-oam-apiserver api doc",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "kdp-oam-apiserver",
					Email: "",
					URL:   "",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache License 2.0",
					URL:  "",
				},
			},
			Version: "v1",
		},
	}
}

func (s *RestServer) requestLog(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	c := utils.NewResponseCapture(resp.ResponseWriter)
	resp.ResponseWriter = c
	chain.ProcessFilter(req, resp)
	takeTime := time.Since(start)
	log.Logger.With(
		"clientIP", utils.Sanitize(utils.ClientIP(req.Request)),
		"path", utils.Sanitize(req.Request.URL.Path),
		"parameters", req.Request.URL.RawQuery,
		"method", req.Request.Method,
		"status", c.StatusCode(),
		"time", takeTime.String(),
		"responseSize", len(c.Bytes()),
		"proto", req.Request.Proto,
		"headers", req.Request.Header,
	).Infof("request log")
	if s.cfg.GenericOptions.LogLevel == "debug" {
		log.Logger.With("responseData", c.Bytes()).Debugf("request log")
	}
}

func (s *RestServer) startHTTP(ctx context.Context) error {
	// Start HTTP api server
	log.Logger.Infof("HTTP APIs are being served on: %s, ctx: %s", s.cfg.BindAddr, ctx)
	server := &http.Server{Addr: s.cfg.BindAddr, Handler: s.webContainer}
	return server.ListenAndServe()
}
