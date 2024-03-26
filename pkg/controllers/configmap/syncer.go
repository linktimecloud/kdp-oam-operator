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

package configmap

import (
	"context"
	"errors"
	"fmt"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	ctrOptions "kdp-oam-operator/cmd/bdc/controller/options"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/utils"
	"sync"
	"time"

	coreV1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	coreInformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CMSyncer struct {
	QPS                 float64
	Burst               int
	InformerSyncPeriod  time.Duration
	Scheme              *runtime.Scheme
	KubeClient          client.Client
	KubeClientSet       *kubernetes.Clientset
	KubeConfig          *rest.Config
	KubeInformerFactory informers.SharedInformerFactory
	lock                sync.RWMutex
}

func (s *CMSyncer) setupInformers() {
	labelSelector := labels.SelectorFromSet(labels.Set{constants.AnnotationCtxSettingSource: "config"})

	configInformer := s.KubeInformerFactory.InformerFor(&coreV1.ConfigMap{}, func(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return coreInformers.NewFilteredConfigMapInformer(
			client,
			"",
			s.InformerSyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			func(options *metaV1.ListOptions) {
				options.LabelSelector = labelSelector.String()
			},
		)
	})
	_, err := configInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.AddFunc,
		UpdateFunc: s.UpdateFunc,
		DeleteFunc: s.DeleteFunc,
	})
	if err != nil {
		return
	}
}

func (s *CMSyncer) AddFunc(obj interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if newRes, ok := obj.(*coreV1.ConfigMap); ok {
		klog.InfoS("add configmap", "configmap namespace", newRes.Name, "name", newRes.Name, "resourceVersion", newRes.ResourceVersion)
		if err := s.SyncConfigMapToContextSetting(newRes); err != nil {
			klog.Errorln(err)
		}
	}

}

func (s *CMSyncer) UpdateFunc(oldObj, newObj interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if newRes, ok := newObj.(*coreV1.ConfigMap); ok {
		klog.InfoS("update configmap", "configmap namespace", newRes.Name, "name", newRes.Name, "resourceVersion", newRes.ResourceVersion)
		if err := s.SyncConfigMapToContextSetting(newRes); err != nil {
			klog.Errorln(err)
		}
	}

}

func (s *CMSyncer) DeleteFunc(obj interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ctx := context.Background()
	configmap := obj.(*coreV1.ConfigMap)
	klog.InfoS("delete configmap", "configmap namespace", configmap.Name, "name", configmap.Name, "resourceVersion", configmap.ResourceVersion)

	var contextSetting bdcv1alpha1.ContextSetting
	contextSetting.Name = fmt.Sprintf("%s-%s", configmap.Namespace, configmap.Name)
	if err := s.KubeClient.Get(ctx, client.ObjectKey{Name: contextSetting.Name}, &contextSetting); err != nil {
		if apierrors.IsNotFound(err) {
			klog.InfoS("Referenced ContextSetting has been deleted")
		}
	}
	klog.InfoS("will delete", "context setting", contextSetting.Name)
	if err := s.KubeClient.Delete(ctx, &contextSetting, &client.DeleteOptions{}); err != nil {
		klog.ErrorS(err, "failed to delete context setting")
	}
}

func (s *CMSyncer) SyncConfigMapToContextSetting(item *coreV1.ConfigMap) error {
	klog.InfoS("InformerFor configmap", "", klog.KRef(item.Namespace, item.Name))
	if _, ok := item.GetLabels()[constants.AnnotationCtxSettingSource]; !ok {
		klog.Errorf("configmap %s is missing necessary annotations: %s", item.Name, constants.AnnotationCtxSettingSource)
		return errors.New("configmap is missing necessary annotations")
	}
	// query configmap namespace
	var ns coreV1.Namespace
	ctx := context.Background()
	err := s.KubeClient.Get(ctx, client.ObjectKey{Name: item.Namespace}, &ns)
	if err != nil {
		return err
	}
	bdc, err := s.bigDataClusterForNSSelector(ctx, ns.Name)
	if err != nil {
		return err
	}
	var referencedBDCOrgName string
	referencedBDCName := bdc.Name
	if _, ok := bdc.GetLabels()[constants.LabelBDCOrgName]; ok {
		referencedBDCOrgName = bdc.GetLabels()[constants.LabelBDCOrgName]
	}
	if _, ok := bdc.GetAnnotations()[constants.AnnotationOrgName]; ok {
		referencedBDCOrgName = bdc.GetAnnotations()[constants.AnnotationOrgName]
	}

	klog.InfoS("take over existing configmap", "configmap", item.Name, "bigdatacluster", referencedBDCName)
	if referencedBDCName == "" {
		klog.Warning("take over existing configmap must specify a bigdatacluster instance", "configmap", item.Name)
		return errors.New("")
	}

	// get configmap referenced definition type
	ctxConfigMapDefType := item.GetAnnotations()[constants.AnnotationCtxSettingType]
	if ctxConfigMapDefType == "" {
		ctxConfigMapDefType = "default"
	}
	contextSetting := &bdcv1alpha1.ContextSetting{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContextSetting",
			APIVersion: bdcv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", item.Namespace, item.Name),
			Annotations: map[string]string{
				constants.AnnotationCtxSettingSource:    "config",
				constants.AnnotationCtxSettingAdopt:     "true",
				constants.AnnotationBDCDefaultNamespace: item.Namespace,
				constants.AnnotationBDCName:             referencedBDCName,
				constants.AnnotationOrgName:             referencedBDCOrgName,
			},
			Labels: map[string]string{
				constants.AnnotationBDCName: referencedBDCName,
				constants.AnnotationOrgName: referencedBDCOrgName,
			},
		},

		Spec: bdcv1alpha1.ContextSettingSpec{
			Name:       item.Name,
			Type:       ctxConfigMapDefType,
			Properties: utils.Object2RawExtension(item.Data),
		},
	}
	if err = s.upsert(ctx, contextSetting); err != nil {
		return err
	}
	return nil
}

func (s *CMSyncer) upsert(ctx context.Context, ctxSetting *bdcv1alpha1.ContextSetting) error {
	originCtxSetting := ctxSetting.DeepCopy()
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if err = s.KubeClient.Get(ctx, client.ObjectKey{Name: originCtxSetting.Name}, originCtxSetting); err != nil {
			if apierrors.IsNotFound(err) {
				klog.InfoS("ContextSetting not found, it will be created later", "", originCtxSetting.Name)
				err = s.KubeClient.Create(ctx, ctxSetting)
				if err != nil {
					klog.ErrorS(err, "Create ContextSetting with error")
					return
				}
				klog.V(1).InfoS("Create ContextSetting", "ContextSetting", ctxSetting)
				return
			}
			return
		}
		klog.InfoS("ContextSetting already exist", "", originCtxSetting.Name)
		ctxSetting.SetResourceVersion(originCtxSetting.GetResourceVersion())
		// set updateTime annotation
		ctxSetting.Annotations[constants.AnnotationBDCUpdatedTime] = metav1.Now().Format(time.RFC3339)
		return s.KubeClient.Update(ctx, ctxSetting)
	})
}

func (s *CMSyncer) bigDataClusterForNSSelector(ctx context.Context, ns string) (*bdcv1alpha1.BigDataCluster, error) {
	var bdcList bdcv1alpha1.BigDataClusterList
	err := s.KubeClient.List(ctx, &bdcList)
	if err != nil {
		return nil, err
	}
	for _, bdc := range bdcList.Items {
		for _, nsItem := range bdc.Spec.Namespaces {
			if nsItem.IsDefault && nsItem.Name == ns {
				klog.InfoS("found bigdatacluster", "bigdatacluster", bdc.Name)
				return &bdc, nil
			}
		}
	}
	return nil, err

}

func (s *CMSyncer) Start() error {
	stopCh := make(chan struct{})
	defer close(stopCh)

	// start Informer List and Watch
	s.KubeInformerFactory.Start(stopCh)

	// wait for all cache sync
	s.KubeInformerFactory.WaitForCacheSync(stopCh)

	<-stopCh
	klog.InfoS("configmap syncer stopped")
	return nil
}

func RunCMSyncer(mgr ctrl.Manager, ctlOptions *ctrOptions.CoreOptions) error {
	syncer := CMSyncer{
		KubeConfig:         mgr.GetConfig(),
		Scheme:             mgr.GetScheme(),
		QPS:                ctlOptions.QPS,
		Burst:              ctlOptions.Burst,
		InformerSyncPeriod: ctlOptions.InformerSyncPeriod,
	}
	syncer.KubeClient = mgr.GetClient()
	clientSet, err := kubernetes.NewForConfig(syncer.KubeConfig)
	if err != nil {
		klog.ErrorS(err, "init kubernetes client set with error")
	}
	syncer.KubeClientSet = clientSet
	syncer.KubeInformerFactory = informers.NewSharedInformerFactory(syncer.KubeClientSet, syncer.InformerSyncPeriod)
	syncer.setupInformers()

	err = syncer.Start()
	if err != nil {
		return err
	}
	return nil

}
