/*
Copyright 2023 KDP(Kubernetes Data Platform).

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

package xdefinitions

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"kdp-oam-operator/api/bdc/common"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	pkgcommon "kdp-oam-operator/pkg/common"
	bdcctrl "kdp-oam-operator/pkg/controllers/bdc"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/controllers/bdc/deftemplate"
	"kdp-oam-operator/pkg/utils"
	"kdp-oam-operator/version"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// Reconciler reconciles a XDefinition object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	options
}

type options struct {
	defRevLimit          int
	concurrentReconciles int
	ignoreDefNoCtrlReq   bool
	controllerVersion    string
}

//+kubebuilder:rbac:groups=bdc.kdp.io,resources=xdefinitions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bdc.kdp.io,resources=xdefinitions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bdc.kdp.io,resources=xdefinitions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the XDefinition object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	klog.InfoS("Reconcile bigdata-cluster XDefinition", "bdc", klog.KRef(req.Namespace, req.Name))

	apiResourceDefMap := make(map[string]string)
	var xDefinition bdcv1alpha1.XDefinition
	if err := r.Get(ctx, req.NamespacedName, &xDefinition); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// handler finalizer
	del, err := r.handlerFinalizer(ctx, xDefinition)
	if del {
		return ctrl.Result{}, err
	}

	def := deftemplate.NewCapabilityXDef(&xDefinition)
	// Store the XDefinition OpenAPI-Schema to configMap
	schemaCMName, err := def.StoreOpenAPISchema(ctx, r.Client, pkgcommon.SystemDefaultNamespace, req.Name)
	if err != nil {
		klog.InfoS("Could not store capability in ConfigMap", "err", err)
		return ctrl.Result{}, nil
	}
	// Store SchemaConfigMapRef
	xDefinition.Status.SchemaConfigMapRef = schemaCMName
	xDefinition.Status.SchemaConfigMapRefNamespace = pkgcommon.SystemDefaultNamespace

	if err := r.UpdateStatus(ctx, &xDefinition); err != nil {
		klog.InfoS("Could not update x Status", "err", err)
		return ctrl.Result{}, nil
	}
	klog.InfoS("Successfully updated the status.schemaConfigMapRef of the XDefinition", "", req.NamespacedName)

	// Store the XDefinition Vs APIResource mapping to configMap
	extractData(xDefinition, apiResourceDefMap)

	var cmName string
	cmName, err = CreateOrUpdateConfigMap(ctx, r.Client, apiResourceDefMap, false)
	if err != nil {
		return ctrl.Result{}, nil
	}
	klog.InfoS("", "", cmName)
	return ctrl.Result{}, nil
}

func extractData(xDefinition bdcv1alpha1.XDefinition, apiResourceDefMap map[string]string) map[string]string {
	if xDefinition.Spec.APIResource.Definition.Kind != "" {
		apiResourceDefinitionType := common.DefaultAPIResourceType
		if xDefinition.Spec.APIResource.Definition.Type != "" {
			apiResourceDefinitionType = xDefinition.Spec.APIResource.Definition.Type
		}
		// apiResourceDefinitionType := xDefinition.Spec.APIResource.Definition.Type
		apiResourceDefMap[fmt.Sprintf("%s-%s", apiResourceDefinitionType, xDefinition.Spec.APIResource.Definition.Kind)] = xDefinition.Name
		// apiResourceDefMap[fmt.Sprintf("Schema-%s-%s", apiResourceDefinitionType, xDefinition.Spec.APIResource.Definition.Kind)] = schemaCMName
	}

	return apiResourceDefMap
}

// UpdateStatus updates Status with retry.RetryOnConflict
func (r *Reconciler) UpdateStatus(ctx context.Context, bdc *bdcv1alpha1.XDefinition, opts ...client.SubResourceUpdateOption) error {
	status := bdc.DeepCopy().Status
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if err = r.Get(ctx, client.ObjectKey{Name: bdc.Name}, bdc); err != nil {
			return
		}
		bdc.Status = status
		return r.Status().Update(ctx, bdc, opts...)
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.concurrentReconciles,
		}).
		For(&bdcv1alpha1.XDefinition{}).
		Complete(r)
}

func (r *Reconciler) handlerFinalizer(ctx context.Context, xdef bdcv1alpha1.XDefinition) (bool, error) {
	if !xdef.DeletionTimestamp.IsZero() {
		if utils.FinalizerExists(&xdef, constants.FinalizerResourceTracker) {
			// clean info in bdc definition configMap
			_, err := CreateOrUpdateConfigMap(ctx, r.Client, extractData(xdef, make(map[string]string)), true)
			if err != nil {
				return true, err
			}
			// Remove the finalizer
			utils.RemoveFinalizer(&xdef, constants.FinalizerResourceTracker)
			err = r.Update(ctx, &xdef)
			return true, err
		}
		return true, nil
	}

	if !utils.FinalizerExists(&xdef, constants.FinalizerResourceTracker) {
		// Add the finalizer
		xdef.SetFinalizers(append(xdef.GetFinalizers(), constants.FinalizerResourceTracker))
		err := r.Update(ctx, &xdef)
		return true, err
	}

	return false, nil
}

func Setup(mgr ctrl.Manager, args bdcctrl.Args) error {
	r := Reconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		options: parseOptions(args),
	}
	return r.SetupWithManager(mgr)
}

func parseOptions(args bdcctrl.Args) options {
	return options{
		defRevLimit:          args.DefRevisionLimit,
		concurrentReconciles: args.ConcurrentReconciles,
		ignoreDefNoCtrlReq:   args.IgnoreDefinitionWithoutControllerRequirement,
		controllerVersion:    version.CoreVersion,
	}
}

func CreateOrUpdateConfigMap(ctx context.Context, k8sClient client.Client, data map[string]string, clean bool) (string, error) {
	//cmName := fmt.Sprintf("%s-%s%s", definitionType, common.CapabilityConfigMapNamePrefix, definitionName)
	cmName := fmt.Sprintf("%s", "bdc-definition-map")
	namespace := pkgcommon.SystemDefaultNamespace
	var cm v1.ConfigMap

	// No need to check the existence of namespace, if it doesn't exist, API server will return the error message
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: cmName}, &cm)
	if err != nil && apierrors.IsNotFound(err) {
		cm = v1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: namespace,
			},
			Data: data,
		}
		err = k8sClient.Create(ctx, &cm)
		if err != nil {
			return cmName, fmt.Errorf(constants.ErrUpdateDefinitionAndAPIResourceMappingConfigMap, cmName, err)
		}
		klog.InfoS("Successfully stored Capability Schema in ConfigMap", "configMap", klog.KRef(namespace, cmName))
		return cmName, nil
	}

	if !clean {
		cm.Data = utils.MergeMapOverrideWithDst(cm.Data, data)
	} else {
		for k, _ := range data {
			delete(cm.Data, k)
		}
	}

	if err = k8sClient.Update(ctx, &cm); err != nil {
		return cmName, fmt.Errorf(constants.ErrUpdateDefinitionAndAPIResourceMappingConfigMap, cmName, err)
	}
	klog.InfoS("Successfully update in ConfigMap", "configMap", klog.KRef(namespace, cmName))
	return cmName, nil
}
