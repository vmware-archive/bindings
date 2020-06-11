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

package servicebindingprep

import (
	"context"

	servicev1alpha1 "github.com/projectriff/bindings/pkg/apis/service/v1alpha1"
	servicebindinginformer "github.com/projectriff/bindings/pkg/client/injection/informers/service/v1alpha1/servicebinding"
	servicebindingreconciler "github.com/projectriff/bindings/pkg/client/injection/reconciler/service/v1alpha1/servicebinding"
	"github.com/projectriff/bindings/pkg/resolver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"
)

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	secretInformer := secretinformer.Get(ctx)
	serviceBindingInformer := servicebindinginformer.Get(ctx)

	r := &Reconciler{
		kubeclient:   kubeclient.Get(ctx),
		secretLister: secretInformer.Lister(),
	}
	impl := servicebindingreconciler.NewImpl(ctx, r)
	r.resolver = resolver.NewServiceableResolver(ctx, impl.EnqueueKey)

	logger.Info("Setting up event handlers.")

	serviceBindingInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	handleMatchingControllers := cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(servicev1alpha1.Kind("ServiceBinding")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	}
	secretInformer.Informer().AddEventHandler(handleMatchingControllers)

	r.tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))
	secretInformer.Informer().AddEventHandler(controller.HandleAll(
		controller.EnsureTypeMeta(
			r.tracker.OnChanged,
			corev1.SchemeGroupVersion.WithKind("Secret"),
		),
	))

	return impl
}
