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
	"context"
	"fmt"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	pkgutils "kdp-oam-operator/pkg/utils"
	"kdp-oam-operator/pkg/utils/log"

	velauxapis "github.com/kubevela/velaux/pkg/server/interfaces/api/dto/v1"
	"github.com/kubevela/velaux/pkg/server/utils"
	"github.com/kubevela/velaux/pkg/server/utils/bcode"
	"github.com/kubevela/workflow/pkg/cue/packages"
	"github.com/oam-dev/kubevela/pkg/velaql"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplicationResourcesService application resources service
type ApplicationResourcesService interface {
	GetApplicationResourcesPods(ctx context.Context, appNs, appName string) (map[string]interface{}, error)
	GetApplicationResourcesPodsDetail(ctx context.Context, podNs, podName string) (map[string]interface{}, error)
	GetApplicationResourcesPodLogs(ctx context.Context, podNs, podName, containerName string, tailLines int32) (map[string]interface{}, error)
	GetApplicationServiceEndpoints(ctx context.Context, appNs, appName string) (map[string]interface{}, error)
	GetApplicationAppliedResources(ctx context.Context, appNs, appName string) (map[string]interface{}, error)
	GetApplicationResourcesTopology(ctx context.Context, appNs, appName string) (map[string]interface{}, error)
	GetApplicationResourcesDetail(ctx context.Context, resNs, resName, resKind, resAPIVersion string) (map[string]interface{}, error)
}

// NewApplicationResourcesService new application service
func NewApplicationResourcesService() ApplicationResourcesService {
	kubeConfig, err := clients.GetKubeConfig()
	if err != nil {
		log.Logger.Fatalf("get kube config failure %s", err.Error())
	}
	kubeClient, err := clients.GetKubeClient()
	if err != nil {
		log.Logger.Fatalf("get kube client failure %s", err.Error())
	}
	return &applicationResourcesServiceImpl{
		KubeClient: kubeClient,
		KubeConfig: kubeConfig,
	}
}

type applicationResourcesServiceImpl struct {
	KubeClient client.Client
	KubeConfig *rest.Config
}

func (a applicationResourcesServiceImpl) queryView(ctx context.Context, ql string) (map[string]interface{}, error) {
	query, err := velaql.ParseVelaQL(ql)
	if err != nil {
		return nil, bcode.ErrParseVelaQL
	}
	velaPD, err := packages.NewPackageDiscover(a.KubeConfig)
	if err != nil {
		if !packages.IsCUEParseErr(err) {
			return nil, err
		}
	}

	queryValue, err := velaql.NewViewHandler(a.KubeClient, a.KubeConfig, velaPD).QueryView(utils.ContextWithUserInfo(ctx), query)
	if err != nil {
		log.Logger.Errorf("fail to query the view %s", err.Error())
		return nil, bcode.ErrViewQuery
	}

	velaQLResp := velauxapis.VelaQLViewResponse{}
	err = queryValue.UnmarshalTo(&velaQLResp)
	if err != nil {
		log.Logger.Errorf("decode the velaQL response to json failure %s", err.Error())
		return nil, bcode.ErrParseQuery2Json
	}
	resp, err := pkgutils.Object2Map(&velaQLResp)
	if err != nil {
		log.Logger.Errorf("decode the velaQL response to json failure %s", err.Error())
		return nil, err
	}
	return resp, nil
}

func (a applicationResourcesServiceImpl) GetApplicationResourcesPods(ctx context.Context, appNs, appName string) (map[string]interface{}, error) {
	ql := fmt.Sprintf("component-pod-view{appNs=%s, appName=%s}.status", appNs, appName)
	return a.queryView(ctx, ql)

}

func (a applicationResourcesServiceImpl) GetApplicationResourcesPodsDetail(ctx context.Context, podNs, podName string) (map[string]interface{}, error) {
	ql := fmt.Sprintf("pod-view{namespace=%s, name=%s, cluster=local}.status", podNs, podName)
	return a.queryView(ctx, ql)
}

func (a applicationResourcesServiceImpl) GetApplicationResourcesPodLogs(ctx context.Context, podNs, podName, containerName string, tailLines int32) (map[string]interface{}, error) {
	ql := fmt.Sprintf("collect-logs{cluster=local, namespace=%s, pod=%s, container=%s, previous=false, timestamps=true, tailLines=%d}", podNs, podName, containerName, tailLines)
	return a.queryView(ctx, ql)
}

func (a applicationResourcesServiceImpl) GetApplicationServiceEndpoints(ctx context.Context, appNs, appName string) (map[string]interface{}, error) {
	ql := fmt.Sprintf("service-endpoints-view{appNs=%s, appName=%s}.status", appNs, appName)
	return a.queryView(ctx, ql)
}

func (a applicationResourcesServiceImpl) GetApplicationAppliedResources(ctx context.Context, appNs, appName string) (map[string]interface{}, error) {
	ql := fmt.Sprintf("service-applied-resources-view{appNs=%s, appName=%s}.status", appNs, appName)
	return a.queryView(ctx, ql)

}

func (a applicationResourcesServiceImpl) GetApplicationResourcesTopology(ctx context.Context, appNs, appName string) (map[string]interface{}, error) {
	ql := fmt.Sprintf("application-resource-tree-view{appNs=%s, appName=%s}.status", appNs, appName)
	return a.queryView(ctx, ql)
}

func (a applicationResourcesServiceImpl) GetApplicationResourcesDetail(ctx context.Context, resNs, resName, resKind, resAPIVersion string) (map[string]interface{}, error) {
	ql := fmt.Sprintf("application-resource-detail-view{namespace=%s, cluster=local, name=%s, kind=%s, apiVersion=%s}.status", resNs, resName, resKind, resAPIVersion)
	if resNs == "" {
		ql = fmt.Sprintf("application-resource-detail-view{cluster=local, name=%s, kind=%s, apiVersion=%s}.status", resName, resKind, resAPIVersion)
	}

	return a.queryView(ctx, ql)
}
