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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"kdp-oam-operator/cmd/apiserver/options"
	"os"
	"os/signal"
	"syscall"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/go-openapi/spec"
	oamcore "github.com/oam-dev/kubevela/apis/core.oam.dev"
	flag "github.com/spf13/pflag"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/pkg/apiserver"
	"kdp-oam-operator/pkg/apiserver/config"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"kdp-oam-operator/pkg/utils/log"
	"kdp-oam-operator/version"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(clients.Scheme))
	_ = bdcv1alpha1.AddToScheme(clients.Scheme)
	_ = oamcore.AddToScheme(clients.Scheme)
	// 添加自定义对象到scheme
	//+kubebuilder:scaffold:scheme
}

func (s *Server) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	cfs := fss.FlagSet("api-server")
	cfs.StringVar(&s.serverConfig.BindAddr, "bind-addr", s.serverConfig.BindAddr, "The bind address used to serve the http APIs.")
	cfs.BoolVar(&s.serverConfig.SwaggerDocEnabled, "swagger-enabled", s.serverConfig.SwaggerDocEnabled, "The swagger enabled flag used to open swagger docs.")
	cfs.Float64Var(&s.serverConfig.KubeQPS, "kube-api-qps", s.serverConfig.KubeQPS, "the qps for kube clients. Low qps may lead to low throughput. High qps may give stress to api-server.")
	cfs.IntVar(&s.serverConfig.KubeBurst, "kube-api-burst", s.serverConfig.KubeBurst, "the burst for kube clients. Recommend setting it qps*3.")
	cfs.StringVar(&s.serverConfig.DefaultSystemNS, "default-system-ns", s.serverConfig.DefaultSystemNS, "the default system namespace")
	cfs.StringVar(&s.serverConfig.LeaderConfig.ID, "id", s.serverConfig.LeaderConfig.ID, "the holder identity name")
	cfs.StringVar(&s.serverConfig.LeaderConfig.LockName, "lock-name", s.serverConfig.LeaderConfig.LockName, "the lease lock resource name")
	cfs.DurationVar(&s.serverConfig.LeaderConfig.Duration, "duration", s.serverConfig.LeaderConfig.Duration, "the lease lock resource name")
	s.serverConfig.GenericOptions.AddFlags(fss.FlagSet("generic"), &s.serverConfig.GenericOptions)

	return fss
}

func main() {
	s := &Server{
		serverConfig: config.APIServerConfig{
			BindAddr:          "0.0.0.0:8000",
			SwaggerDocEnabled: false,
			KubeQPS:           100,
			KubeBurst:         300,
			GenericOptions: options.GenericOptions{
				LogLevel: "info",
			},
		},
	}

	for _, set := range s.Flags().FlagSets {
		flag.CommandLine.AddFlagSet(set)
	}

	flag.Parse()

	// Setup logger
	log.SetUp(s.serverConfig.GenericOptions)
	log.Logger.Debugw("server config", "", s.serverConfig)

	if len(os.Args) > 2 && os.Args[1] == "build-swagger" {
		func() {
			swagger, err := s.buildSwagger()
			if err != nil {
				log.Logger.Fatal(err.Error())
			}
			outData, err := json.MarshalIndent(swagger, "", "\t")
			if err != nil {
				log.Logger.Fatal(err.Error())
			}
			log.Logger.Infof(string(outData))
			swaggerFile, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				log.Logger.Fatal(err.Error())
			}
			defer func() {
				if err := swaggerFile.Close(); err != nil {
					log.Logger.Errorf("close swagger file failure %s", err.Error())
				}
			}()
			_, err = swaggerFile.Write(outData)
			if err != nil {
				log.Logger.Fatal(err.Error())
			}
			fmt.Println("build swagger config file success")
		}()
		return
	}

	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := s.run(ctx, errChan); err != nil {
			errChan <- fmt.Errorf("failed to run apiserver: %w", err)
		}
	}()
	var term = make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case <-term:
		log.Logger.Infof("Received SIGTERM, exiting gracefully...")
	case err := <-errChan:
		log.Logger.Errorf("Received an error: %s, exiting gracefully...", err.Error())
	}
	log.Logger.Infof("See you next time!")
}

// Server apiserver
type Server struct {
	serverConfig config.APIServerConfig
	// genericOptions *options.GenericOptions
}

func (s *Server) run(ctx context.Context, errChan chan error) error {
	log.Logger.Infof("apiserver information: version: %v, gitRevision: %v", version.CoreVersion, version.GitRevision)

	server, err := apiserver.New(s.serverConfig)
	if err != nil {
		klog.Error(err)
	}

	return server.Run(ctx)
}

func (s *Server) buildSwagger() (*spec.Swagger, error) {
	server, err := apiserver.New(s.serverConfig)
	if err != nil {
		return nil, fmt.Errorf("create apiserver failed : %w ", err)
	}
	return restfulspec.BuildSwagger(server.BuildRestfulConfig()), nil
}
