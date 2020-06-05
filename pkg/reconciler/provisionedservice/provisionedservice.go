/*
Copyright 2019 The Knative Authors

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

package provisionedservice

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	bindingsv1alpha1 "github.com/projectriff/bindings/pkg/apis/bindings/v1alpha1"
	provisionedservicereconciler "github.com/projectriff/bindings/pkg/client/injection/reconciler/bindings/v1alpha1/provisionedservice"
	"knative.dev/pkg/reconciler"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and
// reason ProvisionedServiceReconciled.
func newReconciledNormal(namespace, name string) reconciler.Event {
	return reconciler.NewEvent(corev1.EventTypeNormal, "ProvisionedServiceReconciled", "ProvisionedService reconciled: \"%s/%s\"", namespace, name)
}

// Reconciler implements provisionedservicereconciler.Interface for
// ProvisionedService resources.
type Reconciler struct{}

// Check that our Reconciler implements Interface
var _ provisionedservicereconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, o *bindingsv1alpha1.ProvisionedService) reconciler.Event {
	if o.GetDeletionTimestamp() != nil {
		// Check for a DeletionTimestamp.  If present, elide the normal reconcile logic.
		// When a controller needs finalizer handling, it would go here.
		return nil
	}
	o.Status.InitializeConditions()

	o.Status.Binding = o.Spec.Binding

	o.Status.ObservedGeneration = o.Generation
	o.Status.MarkReady()
	return newReconciledNormal(o.Namespace, o.Name)
}
