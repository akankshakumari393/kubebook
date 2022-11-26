/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/apis/core"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	backstorev1beta1 "akankshakumari393.github.io/kubebook/api/v1beta1"
)

// RestoreReconciler reconciles a Restore object
type RestoreReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=backstore.akankshakumari393.github.io,resources=restores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=backstore.akankshakumari393.github.io,resources=restores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=backstore.akankshakumari393.github.io,resources=restores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Restore object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *RestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	log := r.Log.WithValues("restore", req.NamespacedName)
	// TODO(user): your logic here
	var restore backstorev1beta1.Restore
	if err := r.Get(ctx, req.NamespacedName, &restore); err != nil {
		r.Log.Info("failed to get Restore resource.......... object might have been deleted")
		// Ignore NotFound errors as they will be retried automatically if the
		// resource is created in future.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var pvc v1.PersistentVolumeClaim
	r.Log.Info("checking if PVC already exists for this kind restore ")

	err := r.Client.Get(ctx, types.NamespacedName{Name: restore.Spec.VolumeSnapshotResourceName, Namespace: restore.Namespace}, &pvc)
	if apierrors.IsNotFound(err) {
		pvc := persistentVolumeClaim(restore)
		err := r.Create(ctx, pvc)
		if err != nil {
			r.Log.Error(err, "failed to create persistence volume claim resource")
			return ctrl.Result{}, err
		}
		restore.Status.Progress = fmt.Sprintf(pvc.Name + " CREATED")
		err = r.Client.Status().Update(ctx, &restore)
		if err != nil {
			r.Log.Error(err, "Error Updating Status.")
			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(&restore, core.EventTypeNormal, "Created", "Created PVC %s", pvc.Name)

		err = wait.Poll(5*time.Second, 3*time.Minute, func() (done bool, err error) {
			var temppvc v1.PersistentVolumeClaim
			err = r.Client.Get(ctx, types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, &temppvc)
			if err != nil {
				r.Log.Info("Error Getting Current State of PersistenceVolumeClaim %s.\nReason --> %s", pvc.Name, err.Error())
				return false, err
			}
			if temppvc.Status.Phase == "Bound" {
				return true, nil
			}
			log.Info("Waiting for pvc to get Bound ....")
			return false, nil
		})
		if err != nil {
			r.Log.Error(err, "Error Waiting for PVC to be created")
			return ctrl.Result{}, err
		}
		restore.Status.Progress = fmt.Sprintf(pvc.Name + " READY")
		err = r.Client.Status().Update(ctx, &restore)
		if err != nil {
			r.Log.Error(err, "Error Updating READY Status.")
			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(&restore, core.EventTypeNormal, "PVC READY", "PVC IS CREATED AND READY")
		return ctrl.Result{}, nil
	}
	if err != nil {
		r.Log.Error(err, "Cannot create the pvc resource")
		return ctrl.Result{}, err
	}
	log.Info("PVC already exist for the given resource")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&backstorev1beta1.Restore{}).
		Complete(r)
}

func persistentVolumeClaim(restore backstorev1beta1.Restore) *v1.PersistentVolumeClaim {
	storageClass := "csi-hostpath-sc"
	apiGroup := "snapshot.storage.k8s.io/v1"
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            restore.Spec.VolumeSnapshotResourceName,
			Namespace:       restore.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&restore, backstorev1beta1.GroupVersion.WithKind("Restore"))},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			DataSource: &v1.TypedLocalObjectReference{
				Kind:     "VolumeSnapshot",
				Name:     restore.Spec.VolumeSnapshotResourceName,
				APIGroup: &apiGroup,
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
		},
	}
	return pvc
}
