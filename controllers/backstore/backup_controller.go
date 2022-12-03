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
	snapshots "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/apis/core"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	backstorev1beta1 "akankshakumari393.github.io/kubebook/apis/backstore/v1beta1"
)

// BackupReconciler reconciles a Backup object
type BackupReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=backstore.akankshakumari393.github.io,resources=backups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=backstore.akankshakumari393.github.io,resources=backups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=backstore.akankshakumari393.github.io,resources=backups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Backup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := r.Log.WithValues("backup", req.NamespacedName)
	// TODO(user): your logic here
	var backup backstorev1beta1.Backup
	if err := r.Get(ctx, req.NamespacedName, &backup); err != nil {
		r.Log.Info("failed to get Backup resource.......... object might have been deleted")
		// Ignore NotFound errors as they will be retried automatically if the
		// resource is created in future.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var vs snapshots.VolumeSnapshot
	r.Log.Info("checking if an volumesnapshot exists for this kind backup ")

	err := r.Client.Get(ctx, types.NamespacedName{Name: backup.Spec.VolumeSnapshotName, Namespace: backup.Namespace}, &vs)
	if apierrors.IsNotFound(err) {
		r.Log.Info("could not find existing Snapshot for Backup, creating one...")

		vs = *volumeSnapshot(backup)
		err = r.Client.Create(ctx, &vs)
		if err != nil {
			r.Log.Error(err, "failed to create VolumeSnapshot resource")
			return ctrl.Result{}, err
		}
		backup.Status.Progress = fmt.Sprintf(vs.Name + " CREATED")
		err = r.Client.Status().Update(ctx, &backup)
		if err != nil {
			r.Log.Error(err, "Error Updating Status.")
			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(&backup, core.EventTypeNormal, "Created", "Created VolumeSnapshot %s", vs.Name)

		err = wait.Poll(5*time.Second, 3*time.Minute, func() (done bool, err error) {
			var snapshotresource snapshots.VolumeSnapshot
			err = r.Client.Get(ctx, types.NamespacedName{Name: vs.Name, Namespace: vs.Namespace}, &snapshotresource)
			if err != nil {
				r.Log.Info("Error Getting Current State of Volume Snapshot %s.\nReason --> %s", vs.Name, err.Error())
				return false, err
			}
			if snapshotresource.Status != nil && *snapshotresource.Status.ReadyToUse {
				return true, nil
			}
			log.Info("Waiting for Volume Snapshot to get Ready ....")
			return false, nil
		})
		if err != nil {
			r.Log.Error(err, "Error Waiting for Backup to be created")
			return ctrl.Result{}, err
		}
		backup.Status.Progress = fmt.Sprintf(vs.Name + " READY")
		err = r.Client.Status().Update(ctx, &backup)
		if err != nil {
			r.Log.Error(err, "Error Updating READY Status.")
			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(&backup, core.EventTypeNormal, "SNAPSHOT READY", "SNAPSHOT IS CREATED AND READY")
		return ctrl.Result{}, nil
	}
	if err != nil {
		r.Log.Error(err, "Cannot create the resource")
		return ctrl.Result{}, err
	}
	log.Info("Volume Snapshot already exist for the given resource")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&backstorev1beta1.Backup{}).
		Complete(r)
}

func volumeSnapshot(backup backstorev1beta1.Backup) *snapshots.VolumeSnapshot {
	volumeSnapshot := &snapshots.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:            backup.Spec.VolumeSnapshotName,
			Namespace:       backup.Spec.PVCNamespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&backup, backstorev1beta1.GroupVersion.WithKind("Backup"))},
		},
		Spec: snapshots.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &backup.Spec.VolumeSnapshotClassName,
			Source: snapshots.VolumeSnapshotSource{
				PersistentVolumeClaimName: &backup.Spec.PVCName,
			},
		},
	}
	return volumeSnapshot
}
