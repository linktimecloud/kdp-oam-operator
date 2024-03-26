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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	v1types "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	entity "kdp-oam-operator/pkg/apiserver/domain/entity"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/utils/log"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BigDataClusterService bigdata cluster service
type BigDataClusterService interface {
	ListBigDataClusters(ctx context.Context, listOptions v1types.ListOptions) ([]*entity.BigDataClusterEntity, error)
	GetBigDataCluster(ctx context.Context, bdcName string) (*entity.BigDataClusterEntity, error)
}

// NewBigDataClusterService new bigdata cluster service
func NewBigDataClusterService() BigDataClusterService {
	kubeConfig, err := clients.GetKubeConfig()
	if err != nil {
		log.Logger.Fatalf("get kube config failure %s", err.Error())
	}
	kubeClient, err := clients.GetKubeClient()
	if err != nil {
		log.Logger.Fatalf("get kube client failure %s", err.Error())
	}
	return &bigDataClusterServiceImpl{
		KubeClient: kubeClient,
		KubeConfig: kubeConfig,
	}
}

type bigDataClusterServiceImpl struct {
	KubeClient client.Client
	KubeConfig *rest.Config
}

func (b bigDataClusterServiceImpl) ListBigDataClusters(ctx context.Context, options v1types.ListOptions) ([]*entity.BigDataClusterEntity, error) {
	bdcList := new(bdcv1alpha1.BigDataClusterList)
	matchLabels := metav1.LabelSelector{MatchLabels: options.Labels}
	selector, err := metav1.LabelSelectorAsSelector(&matchLabels)
	if err != nil {
		return nil, err
	}
	if err := b.KubeClient.List(ctx, bdcList, &client.ListOptions{
		LabelSelector: selector,
	}); err != nil {
		return nil, err
	}
	var bdcs []*entity.BigDataClusterEntity
	for _, item := range bdcList.Items {
		for _, ns := range item.Spec.Namespaces {
			if ns.IsDefault {
				if _, ok := item.GetLabels()[constants.AnnotationBDCDefaultNamespace]; !ok {
					item.SetLabels(map[string]string{constants.AnnotationBDCDefaultNamespace: ns.Name})
					break
				}
			}
		}
		bdcs = append(bdcs, entity.Object2BigDataClusterEntity(&item))
	}
	return bdcs, nil
}

func (b bigDataClusterServiceImpl) GetBigDataCluster(ctx context.Context, bdcName string) (*entity.BigDataClusterEntity, error) {
	bdc := new(bdcv1alpha1.BigDataCluster)

	if err := b.KubeClient.Get(ctx, client.ObjectKey{Name: bdcName}, bdc); err != nil {
		return nil, err
	}
	for _, ns := range bdc.Spec.Namespaces {
		if ns.IsDefault {
			if _, ok := bdc.GetLabels()[constants.AnnotationBDCDefaultNamespace]; !ok {
				bdc.SetLabels(map[string]string{constants.AnnotationBDCDefaultNamespace: ns.Name})
				break
			}
		}
	}
	return entity.Object2BigDataClusterEntity(bdc), nil
}
