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
	"encoding/json"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	v1types "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/apiserver/domain/entity"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/utils/log"
	"time"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplicationService application service
type ApplicationService interface {
	ListApplications(ctx context.Context, listOptions v1types.ListOptions) ([]*entity.ApplicationEntity, error)
	GetApplication(ctx context.Context, appName string) (*entity.ApplicationEntity, error)
	DetailApplication(ctx context.Context, appName string) (*bdcv1alpha1.Application, error)
	CreateApplication(context.Context, v1types.CreateApplicationRequest) (*v1types.ApplicationBase, error)
	UpdateApplication(context.Context, v1types.UpdateApplicationRequest) (*v1types.ApplicationBase, error)
	DeleteApplication(ctx context.Context, appName string) error
	DeleteApplicationPod(ctx context.Context, podNamespace, podName string) error
}

// NewApplicationService new application service
func NewApplicationService() ApplicationService {
	kubeConfig, err := clients.GetKubeConfig()
	if err != nil {
		log.Logger.Fatalf("get kube config failure %s", err.Error())
	}
	kubeClient, err := clients.GetKubeClient()
	if err != nil {
		log.Logger.Fatalf("get kube client failure %s", err.Error())
	}
	return &applicationServiceImpl{
		KubeClient: kubeClient,
		KubeConfig: kubeConfig,
	}
}

type applicationServiceImpl struct {
	KubeClient client.Client
	KubeConfig *rest.Config
}

func (a applicationServiceImpl) ListApplications(ctx context.Context, options v1types.ListOptions) ([]*entity.ApplicationEntity, error) {
	list := new(bdcv1alpha1.ApplicationList)
	matchLabels := metav1.LabelSelector{MatchLabels: options.Labels}
	selector, err := metav1.LabelSelectorAsSelector(&matchLabels)
	if err != nil {
		return nil, err
	}
	if err := a.KubeClient.List(ctx, list, &client.ListOptions{
		LabelSelector: selector,
	}); err != nil {
		return nil, err
	}

	var apps []*entity.ApplicationEntity
	for _, item := range list.Items {
		apps = append(apps, entity.Object2ApplicationEntity(&item))
	}
	return apps, nil
}

func (a applicationServiceImpl) GetApplication(ctx context.Context, appName string) (*entity.ApplicationEntity, error) {
	app := new(bdcv1alpha1.Application)

	if err := a.KubeClient.Get(ctx, client.ObjectKey{Name: appName}, app); err != nil {
		return nil, err
	}
	return entity.Object2ApplicationEntity(app), nil
}

func (a applicationServiceImpl) DetailApplication(ctx context.Context, appName string) (*bdcv1alpha1.Application, error) {
	app := new(bdcv1alpha1.Application)
	if err := a.KubeClient.Get(ctx, client.ObjectKey{Name: appName}, app); err != nil {
		return nil, err
	}
	app.APIVersion = bigDataClusterAPIVersion
	app.Kind = kindApplication
	return app, nil
}

func (a applicationServiceImpl) CreateApplication(ctx context.Context, request v1types.CreateApplicationRequest) (*v1types.ApplicationBase, error) {
	relatedBDCJSON, _ := json.Marshal(request.BDC)
	app := bdcv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       kindApplication,
			APIVersion: bigDataClusterAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: request.BDC.Name + "-" + request.AppFormName,
			Labels: map[string]string{
				constants.LabelAppFormName:    request.AppFormName,
				constants.LabelAppRuntimeName: request.AppFormName,
				constants.LabelBDCName:        request.BDC.Name,
				// constants.LabelBDCOrgName is required by CUE template rendering
				constants.LabelBDCOrgName: request.BDC.OrgName,
			},
			Annotations: map[string]string{
				constants.AnnotationBDCDefaultNamespace: request.BDC.DefaultNS,
				// constants.AnnotationBDCName is required
				constants.AnnotationBDCName:                 request.BDC.Name,
				constants.AnnotationBDCAppliedConfiguration: string(relatedBDCJSON),
			},
		},
		Spec: bdcv1alpha1.ApplicationSpec{
			Name:       request.AppFormName,
			Type:       request.AppTemplateType,
			Properties: request.Properties,
		},
	}
	if err := a.KubeClient.Create(ctx, &app); err != nil {
		return nil, err
	}
	return nil, nil
}

func (a applicationServiceImpl) UpdateApplication(ctx context.Context, request v1types.UpdateApplicationRequest) (*v1types.ApplicationBase, error) {
	app := new(bdcv1alpha1.Application)

	if err := a.KubeClient.Get(ctx, client.ObjectKey{Name: request.AppName}, app); err != nil {
		return nil, err
	}
	app.Spec.Properties = request.Properties
	app.Annotations[constants.AnnotationBDCUpdatedTime] = metav1.Now().Format(time.RFC3339)
	if err := a.KubeClient.Update(ctx, app); err != nil {
		return nil, err
	}
	return nil, nil
}

func (a applicationServiceImpl) DeleteApplication(ctx context.Context, appName string) error {
	app := new(bdcv1alpha1.Application)

	if err := a.KubeClient.Get(ctx, client.ObjectKey{Name: appName}, app); err != nil {
		return err
	}
	if err := a.KubeClient.Delete(ctx, app); err != nil {
		return err
	}
	return nil
}

func (a applicationServiceImpl) DeleteApplicationPod(ctx context.Context, podNamespace, podName string) error {
	pod := new(corev1.Pod)

	if err := a.KubeClient.Get(ctx, client.ObjectKey{Name: podName, Namespace: podNamespace}, pod); err != nil {
		return err
	}
	if err := a.KubeClient.Delete(ctx, pod); err != nil {
		return err
	}
	return nil
}
