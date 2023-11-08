package controllers

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type finalize func() error

func CheckFinalizer(ctx context.Context, reconciler interface{}, obj client.Object, finalize finalize) error {

	if !controllerutil.ContainsFinalizer(obj, oncallFinalizer) {
		controllerutil.AddFinalizer(obj, oncallFinalizer)
		err := reconciler.(client.Writer).Update(ctx, obj)
		if err != nil {
			return err
		}
	}

	isObjMarkedToBeDeleted := obj.GetDeletionTimestamp() != nil
	if isObjMarkedToBeDeleted {
		if err := finalize(); err != nil {
			return err
		}
		// Remove memcachedFinalizer. Once all finalizers have been
		// removed, the object will be deleted.
		controllerutil.RemoveFinalizer(obj.(client.Object), oncallFinalizer)
		err := reconciler.(client.Writer).Update(ctx, obj.(client.Object))
		if err != nil {
			return err
		}
		return nil
	}
	return nil

}
