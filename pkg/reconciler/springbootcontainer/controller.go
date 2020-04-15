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

package springbootcontainer

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/projectriff/bindings/pkg/apis/bindings/v1alpha1"
	springbootcontainerinformer "github.com/projectriff/bindings/pkg/client/injection/informers/bindings/v1alpha1/springbootcontainer"
	"github.com/projectriff/bindings/pkg/mononoke/cnb"
	"github.com/projectriff/bindings/pkg/resolver"
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
)

const (
	controllerAgentName = "servicebinding-controller"
)

// NewController returns a new ServiceBinding reconciler.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)
	springBootContainerInformer := springbootcontainerinformer.Get(ctx)
	dc := dynamicclient.Get(ctx)

	psInformerFactory := podspecable.Get(ctx)
	c := &psbinding.BaseReconciler{
		GVR: v1alpha1.SchemeGroupVersion.WithResource("springbootcontainers"),
		Get: func(namespace string, name string) (psbinding.Bindable, error) {
			return springBootContainerInformer.Lister().SpringBootContainers(namespace).Get(name)
		},
		DynamicClient: dc,
		Recorder: record.NewBroadcaster().NewRecorder(
			scheme.Scheme, corev1.EventSource{Component: controllerAgentName}),
	}

	impl := controller.NewImpl(c, logger, "SpringBootContainers")
	c.WithContext = WithContextFactory(ctx, impl.EnqueueKey)

	logger.Info("Setting up event handlers")

	springBootContainerInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

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
	serviceInformer := springbootcontainerinformer.Get(ctx)

	// Whenever a ServiceBinding changes our webhook programming might change.
	serviceInformer.Informer().AddEventHandler(handler)

	return func() ([]psbinding.Bindable, error) {
		l, err := serviceInformer.Lister().List(labels.Everything())
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
	kc, err := k8schain.NewNoClient()
	if err != nil {
		panic(fmt.Errorf("unable to create k8schain: %w", err))
	}
	registry := cnb.Registry{Keychain: kc}
	r := resolver.NewPodSpecableResolver(ctx, handler)
	return func(ctx context.Context, b psbinding.Bindable) (context.Context, error) {
		sbc := b.(*v1alpha1.SpringBootContainer)
		psObjs, err := r.PodSpecableFromObjectReference(sbc.Spec.Subject, sbc)
		if err != nil {
			return ctx, fmt.Errorf("failed to get podspecable: %w", err)
		}
		mdm := map[string]cnb.BuildMetadata{}
		for _, psObj := range psObjs {
			_, idx, err := sbc.FindTargetContainer(psObj.Spec.Template.Spec)
			if err != nil {
				return ctx, err
			}
			ref := psObj.Spec.Template.Spec.Containers[idx].Image
			img, err := registry.GetImage(ref)
			if err != nil {
				return ctx, fmt.Errorf("failed to get image %s from registry: %w", ref, err)
			}
			md, err := cnb.ParseBuildMetadata(img)
			if err != nil {
				return ctx, fmt.Errorf("failed parse cnb metadata from image %s: %w", ref, err)
			}
			mdm[psObj.Name] = md
		}
		return v1alpha1.WithSpringBootContainerMetadata(ctx, mdm), nil
	}
}
