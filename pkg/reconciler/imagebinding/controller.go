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

package imagebinding

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/client/injection/ducks/duck/v1/podspecable"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"
	"knative.dev/pkg/webhook/psbinding"

	"github.com/projectriff/bindings/pkg/apis/bindings/v1alpha1"
	imginformer "github.com/projectriff/bindings/pkg/client/injection/informers/bindings/v1alpha1/imagebinding"
	"github.com/projectriff/bindings/pkg/resolver"
)

const (
	controllerAgentName = "imagebinding-controller"
)

// NewController returns a new ImageBinding reconciler.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)
	ibInformer := imginformer.Get(ctx)
	dc := dynamicclient.Get(ctx)

	psInformerFactory := podspecable.Get(ctx)
	c := &psbinding.BaseReconciler{
		GVR: v1alpha1.SchemeGroupVersion.WithResource("imagebindings"),
		Get: func(namespace string, name string) (psbinding.Bindable, error) {
			return ibInformer.Lister().ImageBindings(namespace).Get(name)
		},
		DynamicClient: dc,
		Recorder: record.NewBroadcaster().NewRecorder(
			scheme.Scheme, corev1.EventSource{Component: controllerAgentName}),
	}

	impl := controller.NewImpl(c, logger, "ImageBindings")
	c.WithContext = WithContextFactory(ctx, impl.EnqueueKey)

	logger.Info("Setting up event handlers")

	ibInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	c.Tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))
	c.Factory = &duck.CachedInformerFactory{
		Delegate: &duck.EnqueueInformerFactory{
			Delegate:     psInformerFactory,
			EventHandler: controller.HandleAll(c.Tracker.OnChanged),
		},
	}
	return impl
}

func ListAll(ctx context.Context, handler cache.ResourceEventHandler) psbinding.ListAll {
	imgInformer := imginformer.Get(ctx)

	// Whenever a ImageBinding changes our webhook programming might change.
	imgInformer.Informer().AddEventHandler(handler)

	return func() ([]psbinding.Bindable, error) {
		l, err := imgInformer.Lister().List(labels.Everything())
		if err != nil {
			return nil, err
		}
		bl := make([]psbinding.Bindable, 0, len(l))
		for _, elt := range l {
			bl = append(bl, elt)
		}
		return bl, nil
	}
}

func WithContextFactory(ctx context.Context, handler func(name types.NamespacedName)) psbinding.BindableContext {
	r := resolver.NewLatestImageResolver(ctx, handler)
	return func(ctx context.Context, b psbinding.Bindable) (context.Context, error) {
		sb := b.(*v1alpha1.ImageBinding)
		img, err := r.LatestImageFromObjectReference(sb.Spec.Provider, sb)
		if err != nil {
			return nil, err
		}
		return v1alpha1.WithLatestImage(ctx, img), nil
	}
}
