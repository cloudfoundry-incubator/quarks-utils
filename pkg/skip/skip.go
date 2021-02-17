// Package skip helps with skiping reconciles for stale resources
package skip

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	crc "sigs.k8s.io/controller-runtime/pkg/client"

	log "code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
)

// Object is used as a helper interface when passing Kubernetes resources
// between methods.
// All Kubernetes resources should implement both of these interfaces
type Object interface {
	runtime.Object
	metav1.Object
}

// Reconciles returns true if the object is stale, and shouldn't be enqueued for reconciliation
// The object can be a ConfigMap or a Secret
func Reconciles(ctx context.Context, client crc.Client, object Object) bool {
	var newResourceVersion string

	switch object := object.(type) {
	case *corev1.ConfigMap:
		cm := &corev1.ConfigMap{}
		err := client.Get(ctx, types.NamespacedName{Name: object.Name, Namespace: object.Namespace}, cm)
		if err != nil {
			log.Errorf(ctx, "Failed to get ConfigMap '%s/%s': %s", object.Namespace, object.Name, err)
			return true
		}

		newResourceVersion = cm.ResourceVersion
	case *corev1.Secret:
		s := &corev1.Secret{}
		err := client.Get(ctx, types.NamespacedName{Name: object.Name, Namespace: object.Namespace}, s)
		if err != nil {
			log.Errorf(ctx, "Failed to get Secret '%s/%s': %s", object.Namespace, object.Name, err)
			return true
		}

		newResourceVersion = s.ResourceVersion
	default:
		return false
	}

	if object.GetResourceVersion() != newResourceVersion {
		log.Debugf(ctx, "Skipping reconcile request for old resource version of '%s/%s'", object.GetNamespace(), object.GetName())
		return true
	}
	return false
}
