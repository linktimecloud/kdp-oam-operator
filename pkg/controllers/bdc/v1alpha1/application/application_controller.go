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

package application

import (
	"context"
	"fmt"
	"kdp-oam-operator/api/bdc/common"
	conditiontype "kdp-oam-operator/api/bdc/condition"
	"kdp-oam-operator/api/bdc/v1alpha1"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	bdcctrl "kdp-oam-operator/pkg/controllers/bdc"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/controllers/bdc/parser"
	"kdp-oam-operator/pkg/controllers/utils/condition"
	"kdp-oam-operator/pkg/controllers/utils/dispatch"
	"kdp-oam-operator/pkg/controllers/utils/vela"
	"kdp-oam-operator/pkg/utils"
	"kdp-oam-operator/version"

	velav1beta1 "github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler reconciles a Application object
type Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	options
}

type options struct {
	defRevLimit          int
	concurrentReconciles int
	ignoreDefNoCtrlReq   bool
	controllerVersion    string
}

//+kubebuilder:rbac:groups=bdc.kdp.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bdc.kdp.io,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bdc.kdp.io,resources=applications/finalizers,verbs=update
//+kubebuilder:rbac:groups=core.oam.dev/v1beta1,resources=applications,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (reconciler *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.InfoS("Reconcile application", "", klog.KRef(req.Namespace, req.Name))

	// Lookup the application instance for this reconcile request
	var application v1alpha1.Application
	if err := reconciler.Get(ctx, req.NamespacedName, &application); err != nil {
		klog.Errorf("[application] [namespace：%s, name: %s] Get bdc application error: %v", application.Namespace, application.Name, err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if application.Status.Status == "" {
		return ctrl.Result{}, reconciler.reconcileStatus(ctx, application, "", "")
	}

	// Set BigDataCluster as metadata.ownerReferences
	var bigDataCluster bdcv1alpha1.BigDataCluster
	if err := reconciler.Get(ctx, client.ObjectKey{Name: application.GetAnnotations()[constants.AnnotationBDCName]}, &bigDataCluster); err != nil {
		// if BigDataCluster is not existed, delete the application
		if apierrors.IsNotFound(err) && application.Status.Status != common.ApplicationInitializing {
			reconciler.deleteFinalizer(ctx, application)
			reconciler.Delete(ctx, &application)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, reconciler.reconcileStatusWithInitializeError(ctx, application, err)
	}
	application.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         bigDataCluster.APIVersion,
			Kind:               bigDataCluster.Kind,
			Name:               bigDataCluster.Name,
			UID:                bigDataCluster.UID,
			Controller:         pointer.Bool(true),
			BlockOwnerDeletion: pointer.Bool(true),
		},
	})

	if err := reconciler.patchOwnerReferencer(ctx, &application); err != nil {
		return ctrl.Result{}, reconciler.reconcileStatusWithInitializeError(ctx, application, err)
	}

	// handler finalizer
	del, err := reconciler.handlerFinalizer(ctx, application, bigDataCluster)
	if del {
		return ctrl.Result{}, err
	}

	// Replace template.parameter with BigDataCluster Object spec
	bdcParser := parser.NewParser(reconciler.Client)

	// Dispatch manifests
	bdcDispatcher := dispatch.NewManifestsDispatcher(reconciler.Client)
	bdcFile, err := bdcParser.GenerateBigDataClusterFile(ctx, &application)
	if err != nil {
		klog.Errorf("[application] [namespace：%s, name: %s] Generate BigDataClusterFile error: %v", application.Namespace, application.Name, err)
		return ctrl.Result{}, reconciler.reconcileStatusWithInitializeError(ctx, application, err)
	}

	manifests, err := bdcFile.PrepareManifests(ctx, req)
	if err != nil {
		klog.Errorf("[application] [namespace：%s, name: %s] Prepare manifests error: %v", application.Namespace, application.Name, err)
		return ctrl.Result{}, reconciler.reconcileStatusWithInitializeError(ctx, application, err)
	}
	// klog.InfoS("ContextSetting", "output manifests", manifests)

	if len(manifests) > 0 {
		mergeMetaData(&manifests, application)
		if err := bdcDispatcher.Dispatch(ctx, manifests...); err != nil {
			klog.Errorf("[application] [namespace：%s, name: %s]Handle Apply Manifests error: %v", application.Namespace, application.Name, err)
			return ctrl.Result{}, reconciler.reconcileStatusWithOutputDefError(ctx, application, err)
		}
		klog.Info("Successfully generated manifests")
	}

	if err = reconciler.reconcileStatus(ctx, application, bdcFile.BDCTemplate.FullTemplate.XDefinitionSchemaName, bdcFile.DownStreamNamespace); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (reconciler *Reconciler) handlerFinalizer(ctx context.Context, application bdcv1alpha1.Application, bigDataCluster bdcv1alpha1.BigDataCluster) (bool, error) {
	if !application.GetDeletionTimestamp().IsZero() {
		if !utils.FinalizerExists(&application, constants.FinalizerResourceTracker) {
			return true, nil
		}
		// check vela application exist
		var velaApplication velav1beta1.Application
		namespace, _ := parser.GenerateDownStreamNamespace(&bigDataCluster)
		err := reconciler.Get(ctx, client.ObjectKey{Name: application.Spec.Name, Namespace: namespace}, &velaApplication)
		if err != nil && apierrors.IsNotFound(err) {
			// if vela application deleted,the application should be removed
			reconciler.deleteFinalizer(ctx, application)
			return true, err
		}

		klog.InfoS("vela application exist, waiting for delete.", "application name", application.Name)

		application.Status.Status = common.ApplicationDeleting
		err = reconciler.UpdateStatus(ctx, &application)
		if err != nil {
			klog.Errorf("unable to update application %s status, error: %v: ", application.Name, err)
		}

		// delete vela application
		err = reconciler.Delete(ctx, &velaApplication)
		if err != nil {
			return true, err
		}
		return true, nil
	}

	// add finalizer
	if !utils.FinalizerExists(&application, constants.FinalizerResourceTracker) {
		application.SetFinalizers(append(application.GetFinalizers(), constants.FinalizerResourceTracker))
		if err := reconciler.Update(ctx, &application); err != nil {
			klog.Errorf("[application] [namespace：%s, name: %s] Add finalizer error: %v", application.Namespace, application.Name, err)
			return true, err
		}
	}

	return false, nil
}

func (reconciler *Reconciler) deleteFinalizer(ctx context.Context, application bdcv1alpha1.Application) {
	utils.RemoveFinalizer(&application, constants.FinalizerResourceTracker)
	if err := reconciler.Update(ctx, &application); err != nil {
		klog.Errorf("[application] [namespace：%s, name: %s] Remove finalizer error: %v", application.Namespace, application.Name, err)
	}
}

func mergeMetaData(manifest *[]*unstructured.Unstructured, application bdcv1alpha1.Application) {
	fileterKey := []string{v1.LastAppliedConfigAnnotation, constants.AnnotationLastAppliedConfig}
	for _, m := range *manifest {
		m.SetAnnotations(utils.MergeMapOverrideWithFilters(application.GetAnnotations(), m.GetAnnotations(), fileterKey))
		m.SetLabels(utils.MergeMapOverrideWithFilters(application.GetLabels(), m.GetLabels(), nil))
	}
}

func (reconciler *Reconciler) patchOwnerReferencer(ctx context.Context, application *bdcv1alpha1.Application) error {
	if err := reconciler.Patch(ctx, application, client.Merge); err != nil {
		klog.Info(err, "unable to patch annotation")
	}
	klog.InfoS("patch", "Object", application.Name, "OwnerReferencer", application.OwnerReferences)
	return nil
}

// UpdateStatus update Status with retry.RetryOnConflict
func (reconciler *Reconciler) UpdateStatus(ctx context.Context, application *bdcv1alpha1.Application, opts ...client.SubResourceUpdateOption) error {
	status := application.DeepCopy().Status
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if err = reconciler.Get(ctx, client.ObjectKey{Name: application.Name}, application); err != nil {
			return
		}
		application.Status = status
		return reconciler.Status().Update(ctx, application, opts...)
	})
}

// SetupWithManager sets up the controller with the Manager.
func (reconciler *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bdcv1alpha1.Application{}).
		Owns(&velav1beta1.Application{}). // watch vela application for sync status
		Complete(reconciler)
}

func Setup(mgr ctrl.Manager, args bdcctrl.Args) error {
	r := Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("bdc-org-resource-control-controller"),
		options:  parseOptions(args),
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

func (reconciler *Reconciler) reconcileStatusWithInitializeError(ctx context.Context, application bdcv1alpha1.Application, err error) error {
	if err == nil {
		return nil
	}

	application.Status.Status = common.ApplicationInitializeError
	application.Status.SetConditions(conditiontype.ReconcileError(err))

	return reconciler.UpdateStatus(ctx, &application)
}

func (reconciler *Reconciler) reconcileStatusWithOutputDefError(ctx context.Context, application bdcv1alpha1.Application, err error) error {
	if err == nil {
		return nil
	}

	application.Status.Status = common.ApplicationOutputDefError
	application.Status.SetConditions(conditiontype.ReconcileError(err))

	return reconciler.UpdateStatus(ctx, &application)
}

// reconcile application status to sync vela application status.
// Actually this controller just generate vela application not for final workload(like deployment/sts),
// so it will use vela application status directly
func (reconciler *Reconciler) reconcileStatus(ctx context.Context, application bdcv1alpha1.Application, xDefinitionSchemaName string, DownStreamNamespace string) error {
	if application.Status.Status == "" {
		application.Status.Status = common.ApplicationInitializing
		// update .status.status field only, trigger next reconcile handling
		return reconciler.UpdateStatus(ctx, &application)
	}

	application.Status.SchemaConfigMapRef = xDefinitionSchemaName

	// get vela application status, set into bdc application
	var velaApplication velav1beta1.Application
	if err := reconciler.Get(ctx, client.ObjectKey{Name: application.Spec.Name, Namespace: DownStreamNamespace}, &velaApplication); err != nil {
		klog.Error(err, "[Reconcile Status] failed to get vela application")
		return condition.PatchCondition(ctx, reconciler, &application, conditiontype.ReconcileError(fmt.Errorf(constants.ErrGetVelaApplication, err)))
	}

	desiredConditions := vela.DesiredConditionFromVela(&velaApplication)
	desiredWorkflowStatus := vela.WorkflowStatusFromVela(velaApplication.Status.Workflow)
	if !vela.DiffCondition(desiredConditions, application.Status.Conditions) &&
		!vela.DiffWorkflowStatus(desiredWorkflowStatus, application.Status.Workflow) &&
		application.Status.Status == string(velaApplication.Status.Phase) {
		// status no changes
		return nil
	}

	if velaApplication.Status.Phase == "" {
		application.Status.Status = common.ApplicationStarting
	} else {
		application.Status.Status = string(velaApplication.Status.Phase)
	}
	application.Status.Conditions = desiredConditions
	application.Status.Workflow = desiredWorkflowStatus
	application.Status.AppliedResources = vela.AppliedResourcesFormVela(velaApplication.Status.AppliedResources)
	application.Status.Services = vela.ServicesFromVela(velaApplication.Status.Services)

	if err := reconciler.UpdateStatus(ctx, &application); err != nil {
		klog.Error(err, "[Reconcile Status] failed to update status")
		return err
	}

	return nil
}
