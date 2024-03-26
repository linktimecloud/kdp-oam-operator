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

package bigdatacluster

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	conditiontype "kdp-oam-operator/api/bdc/condition"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	bdcctrl "kdp-oam-operator/pkg/controllers/bdc"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/controllers/bdc/parser"
	"kdp-oam-operator/pkg/controllers/utils/condition"
	"kdp-oam-operator/pkg/controllers/utils/dispatch"
	"kdp-oam-operator/version"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Reconciler reconciles a BigDataCluster object
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

//+kubebuilder:rbac:groups=bdc.kdp.io,resources=bigdataclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bdc.kdp.io,resources=bigdataclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bdc.kdp.io,resources=bigdataclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the BigDataCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	klog.InfoS("Reconcile bigdata-cluster", "bdc", klog.KRef(req.Namespace, req.Name))

	// Lookup the BigDataCluster instance for this reconcile request
	var bigDataCluster bdcv1alpha1.BigDataCluster
	if err := r.Get(ctx, req.NamespacedName, &bigDataCluster); err != nil {
		if errors.IsNotFound(err) {
			klog.InfoS("[Warning]Not found", "name", req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	//klog.InfoS("BigDataCluster", "bdc", bigDataCluster)

	// Replace template.parameter with BigDataCluster Object spec
	bdcParser := parser.NewParser(r.Client)

	// Dispatch manifests
	bdcDispatcher := dispatch.NewManifestsDispatcher(r.Client)
	bdcFile, err := bdcParser.GenerateBigDataClusterFile(ctx, &bigDataCluster)
	if err != nil {
		klog.Error(err, "[Generate BigDataClusterFile]")
		return ctrl.Result{}, err
	}
	// SetOwnerReference as false, Because Finalizer logic needs to be executed when deleting downstream resources
	bdcFile.SetOwnerReference = false

	manifests, err := bdcFile.PrepareManifests(ctx, req)
	if err != nil {
		klog.Error(err, "[Handle PrepareManifests]")
		return ctrl.Result{}, err
	}
	bdcFile.ReferredObjects = manifests
	// klog.InfoS("BigDataCluster", "output manifests", manifests)

	// Handle Finalizer
	endReconcile, result, err := r.handleFinalizers(ctx, &bigDataCluster, bdcFile)
	if err != nil {
		return result, err
	}
	if endReconcile {
		return result, nil
	}

	if len(manifests) > 0 {
		if err := bdcDispatcher.Dispatch(ctx, manifests...); err != nil {
			klog.Error(err, "[Handle Apply Manifests]")
		}
		klog.Info("Successfully generated manifests")
	}
	// UpdateStatus
	bigDataCluster.Status.Status = bdcv1alpha1.ActiveBigDataCluster
	bigDataCluster.Status.SchemaConfigMapRef = bdcFile.BDCTemplate.FullTemplate.XDefinitionSchemaName
	if bigDataCluster.Spec.Disabled {
		bigDataCluster.Status.Status = bdcv1alpha1.DisabledBigDataCluster
	}
	if bigDataCluster.Spec.Frozen {
		bigDataCluster.Status.Status = bdcv1alpha1.FrozenBigDataCluster
	}
	err = r.UpdateStatus(ctx, &bigDataCluster)
	if err != nil {
		err = condition.PatchCondition(ctx, r, &bigDataCluster,
			conditiontype.ReconcileError(fmt.Errorf(constants.ErrCreateBDCResource, bigDataCluster.Kind, bigDataCluster.Name, err)))
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	err = condition.PatchCondition(ctx, r, &bigDataCluster, conditiontype.ReconcileSuccess())
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// UpdateStatus update Status with retry.RetryOnConflict
func (r *Reconciler) UpdateStatus(ctx context.Context, bdc *bdcv1alpha1.BigDataCluster, opts ...client.SubResourceUpdateOption) error {
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
		For(&bdcv1alpha1.BigDataCluster{}).
		Complete(r)
}

func Setup(mgr ctrl.Manager, args bdcctrl.Args) error {
	r := Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("bdc-controller"),
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

func (r *Reconciler) handleFinalizers(ctx context.Context, bigDataCluster *bdcv1alpha1.BigDataCluster, bdcFile *parser.BDCFile) (bool, ctrl.Result, error) {
	// name of our custom finalizer
	myFinalizerName := "ns.bdc.kdp.io/finalizer"

	// examine DeletionTimestamp to determine if object is under deletion
	if bigDataCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(bigDataCluster, myFinalizerName) {
			controllerutil.AddFinalizer(bigDataCluster, myFinalizerName)
			if err := r.Update(ctx, bigDataCluster); err != nil {
				return true, ctrl.Result{}, err
			}
			klog.InfoS("Register new finalizer for bigdatacluster", "finalizer", myFinalizerName)
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(bigDataCluster, myFinalizerName) {
			// Set status as Terminating
			bigDataCluster.Status.Status = bdcv1alpha1.TerminatingBigDataCluster
			err := r.UpdateStatus(ctx, bigDataCluster)
			if err != nil {
				return true, ctrl.Result{}, err
			}

			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, bdcFile); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return true, ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(bigDataCluster, myFinalizerName)
			if err := r.Update(ctx, bigDataCluster); err != nil {
				return true, ctrl.Result{}, err
			}
		}
		// Stop reconciliation as the item is being deleted
		return true, ctrl.Result{}, nil
	}
	return false, ctrl.Result{}, nil
}

func (r *Reconciler) deleteExternalResources(ctx context.Context, bdcFile *parser.BDCFile) error {
	//
	// delete any external resources associated with the cronJob
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple times for same object.
	klog.InfoS("Prepare to delete downstream resource", "Objects", bdcFile.ReferredObjects)
	//time.Sleep(10 * time.Second)
	for _, ro := range bdcFile.ReferredObjects {
		err := r.Client.Delete(ctx, ro)
		if err != nil {
			return err
		}
		klog.InfoS("Deleted", "Kind", ro.GroupVersionKind(), "name", ro.GetName())
	}
	klog.InfoS("Finish to delete downstream resource", "Objects", bdcFile.ReferredObjects)
	return nil
}
