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

package contextsecret

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	conditiontype "kdp-oam-operator/api/bdc/condition"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	bdcctrl "kdp-oam-operator/pkg/controllers/bdc"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/controllers/bdc/parser"
	"kdp-oam-operator/pkg/controllers/utils/condition"
	"kdp-oam-operator/pkg/controllers/utils/dispatch"
	"kdp-oam-operator/version"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler reconciles a ContextSecret object
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

//+kubebuilder:rbac:groups=bdc.kdp.io,resources=contextsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bdc.kdp.io,resources=contextsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bdc.kdp.io,resources=contextsecrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the ContextSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.InfoS("Reconcile context-secret", "", klog.KRef(req.Namespace, req.Name))

	// Lookup the contextSecret instance for this reconcile request
	var contextSecret bdcv1alpha1.ContextSecret
	if err := r.Get(ctx, req.NamespacedName, &contextSecret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//klog.InfoS("ContextSecret", "bdc.cs", contextSecret)

	// Set BigDataCluster as metadata.ownerReferences
	var bigDataCluster bdcv1alpha1.BigDataCluster
	if err := r.Get(ctx, client.ObjectKey{Name: contextSecret.GetAnnotations()[constants.AnnotationBDCName]}, &bigDataCluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if contextSecret.GetOwnerReferences() == nil {
		contextSecret.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion:         bigDataCluster.APIVersion,
				Kind:               bigDataCluster.Kind,
				Name:               bigDataCluster.Name,
				UID:                bigDataCluster.UID,
				Controller:         pointer.Bool(true),
				BlockOwnerDeletion: pointer.Bool(true),
			},
		})
		err := r.patchOwnerReferencer(ctx, &contextSecret)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Replace template.parameter with BigDataCluster Object spec
	bdcParser := parser.NewParser(r.Client)

	// Dispatch manifests
	bdcDispatcher := dispatch.NewManifestsDispatcher(r.Client)
	bdcFile, err := bdcParser.GenerateBigDataClusterFile(ctx, &contextSecret)
	if err != nil {
		klog.Error(err, "[Generate BigDataClusterFile]")
		return ctrl.Result{}, err
	}

	manifests, err := bdcFile.PrepareManifests(ctx, req)
	if err != nil {
		klog.Error(err, "[Handle PrepareManifests]")
		return ctrl.Result{}, err
	}
	// klog.InfoS("ContextSecret", "output manifests", manifests)

	if len(manifests) > 0 {
		if err := bdcDispatcher.Dispatch(ctx, manifests...); err != nil {
			klog.Error(err, "[Handle Apply Manifests]")
		}
		klog.Info("Successfully generated manifests")
	}
	contextSecret.Status.SchemaConfigMapRef = bdcFile.BDCTemplate.FullTemplate.XDefinitionSchemaName
	// UpdateStatus
	err = r.UpdateStatus(ctx, &contextSecret)
	if err != nil {
		err = condition.PatchCondition(ctx, r, &contextSecret,
			conditiontype.ReconcileError(fmt.Errorf(constants.ErrCreateBDCResource, contextSecret.Kind, contextSecret.Name, err)))
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	err = condition.PatchCondition(ctx, r, &contextSecret, conditiontype.ReconcileSuccess())
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// UpdateStatus update Status with retry.RetryOnConflict
func (r *Reconciler) UpdateStatus(ctx context.Context, bdc *bdcv1alpha1.ContextSecret, opts ...client.SubResourceUpdateOption) error {
	status := bdc.DeepCopy().Status
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if err = r.Get(ctx, client.ObjectKey{Name: bdc.Name}, bdc); err != nil {
			return
		}
		bdc.Status = status
		return r.Status().Update(ctx, bdc, opts...)
	})
}

func (r *Reconciler) patchOwnerReferencer(ctx context.Context, bdc *bdcv1alpha1.ContextSecret) error {
	if err := r.Patch(ctx, bdc, client.Merge); err != nil {
		klog.Info(err, "unable to patch annotation")
	}
	klog.InfoS("patch", "Object", bdc.Name, "OwnerReferencer", bdc.OwnerReferences)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.concurrentReconciles,
		}).
		For(&bdcv1alpha1.ContextSecret{}).
		Complete(r)
}

func Setup(mgr ctrl.Manager, args bdcctrl.Args) error {
	r := Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("bdc-custom-secret-controller"),
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
