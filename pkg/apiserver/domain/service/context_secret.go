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
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	v1types "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	entity "kdp-oam-operator/pkg/apiserver/domain/entity"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"kdp-oam-operator/pkg/utils/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ContextSecretService context secret service
type ContextSecretService interface {
	ListContextSecrets(ctx context.Context, listOptions v1types.ListOptions) ([]*entity.ContextSecretEntity, error)
	GetContextSecret(ctx context.Context, name string) (*entity.ContextSecretEntity, error)
}

// NewContextSecretService new context secret service
func NewContextSecretService() ContextSecretService {
	kubeConfig, err := clients.GetKubeConfig()
	if err != nil {
		log.Logger.Fatalf("get kube config failure %s", err.Error())
	}
	kubeClient, err := clients.GetKubeClient()
	if err != nil {
		log.Logger.Fatalf("get kube client failure %s", err.Error())
	}
	return &contextSecretServiceImpl{
		KubeClient: kubeClient,
		KubeConfig: kubeConfig,
	}
}

type contextSecretServiceImpl struct {
	KubeClient client.Client
	KubeConfig *rest.Config
}

func (c contextSecretServiceImpl) ListContextSecrets(ctx context.Context, options v1types.ListOptions) ([]*entity.ContextSecretEntity, error) {
	list := new(bdcv1alpha1.ContextSecretList)
	matchLabels := metav1.LabelSelector{MatchLabels: options.Labels}
	selector, err := metav1.LabelSelectorAsSelector(&matchLabels)
	if err != nil {
		return nil, err
	}
	if err := c.KubeClient.List(ctx, list, &client.ListOptions{
		LabelSelector: selector,
	}); err != nil {
		return nil, err
	}

	var ctxSecrets []*entity.ContextSecretEntity
	for _, item := range list.Items {
		ctxSecrets = append(ctxSecrets, entity.Object2ContextSecretEntity(&item))
	}
	return ctxSecrets, nil
}

func (c contextSecretServiceImpl) GetContextSecret(ctx context.Context, name string) (*entity.ContextSecretEntity, error) {
	ctxSecret := new(bdcv1alpha1.ContextSecret)
	if err := c.KubeClient.Get(ctx, client.ObjectKey{Name: name}, ctxSecret); err != nil {
		return nil, err
	}
	return entity.Object2ContextSecretEntity(ctxSecret), nil
}
