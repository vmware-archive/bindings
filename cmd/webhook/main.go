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

package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/configmaps"
	"knative.dev/pkg/webhook/psbinding"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"

	"github.com/projectriff/bindings/pkg/apis/bindings/v1alpha1"
	"github.com/projectriff/bindings/pkg/reconciler/bindableservice"
	"github.com/projectriff/bindings/pkg/reconciler/imagebinding"
	"github.com/projectriff/bindings/pkg/reconciler/servicebinding"
)

var (
	//BindingExcludeLabel can be applied to exclude resource from webhook
	BindingExcludeLabel = "bindings.projectriff.io/exclude"
	//BindingIncludeLabel can be applied to include resource in webhook
	BindingIncludeLabel = "bindings.projectriff.io/include"

	ExclusionSelector = metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key:      BindingExcludeLabel,
			Operator: metav1.LabelSelectorOpNotIn,
			Values:   []string{"true"},
		}},
	}
	InclusionSelector = metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key:      BindingIncludeLabel,
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"true"},
		}},
	}
)
var ourTypes = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	v1alpha1.SchemeGroupVersion.WithKind("BindableService"): &v1alpha1.BindableService{},
	v1alpha1.SchemeGroupVersion.WithKind("ImageBinding"):    &v1alpha1.ImageBinding{},
	v1alpha1.SchemeGroupVersion.WithKind("ServiceBinding"):  &v1alpha1.ServiceBinding{},
}

func NewDefaultingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return defaulting.NewAdmissionController(ctx,
		// Name of the resource webhook.
		"defaulting.webhook.bindings.projectriff.io",

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to validate and default.
		ourTypes,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func NewValidationAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,
		// Name of the resource webhook.
		"validation.webhook.bindings.projectriff.io",

		// The path on which to serve the webhook.
		"/validation",

		// The resources to validate and default.
		ourTypes,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func NewConfigValidationController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return configmaps.NewAdmissionController(ctx,

		// Name of the configmap webhook.
		"config.webhook.bindings.projectriff.io",

		// The path on which to serve the webhook.
		"/config-validation",

		// The configmaps to validate.
		configmap.Constructors{
			logging.ConfigMapName(): logging.NewConfigFromConfigMap,
			metrics.ConfigMapName(): metrics.NewObservabilityConfigFromConfigMap,
		},
	)
}

func NewBindingWebhook(resource string, gla psbinding.GetListAll, wcf WithContextFactory) injection.ControllerConstructor {
	selector := psbinding.WithSelector(ExclusionSelector)
	if os.Getenv("BINDING_SELECTION_MODE") == "inclusion" {
		selector = psbinding.WithSelector(InclusionSelector)
	}
	return func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
		wc := wcf(ctx, func(types.NamespacedName) {})
		return psbinding.NewAdmissionController(ctx,
			// Name of the resource webhook.
			fmt.Sprintf("%s.webhook.bindings.projectriff.io", resource),

			// The path on which to serve the webhook.
			fmt.Sprintf("/%s", resource),

			// How to get all the Bindables for configuring the mutating webhook.
			gla,

			// How to setup the context prior to invoking Do/Undo.
			wc,
			selector,
		)
	}
}

func main() {
	// Set up a signal context with our webhook options
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "webhook",
		Port:        8443,
		SecretName:  "webhook-certs",
	})

	sharedmain.WebhookMainWithContext(ctx, "webhook",
		// Our singleton certificate controller.
		certificates.NewController,

		// Our singleton webhook admission controllers
		NewDefaultingAdmissionController,
		NewValidationAdmissionController,
		NewConfigValidationController,

		// For each binding we have a controller and a binding webhook.
		bindableservice.NewController,
		imagebinding.NewController, NewBindingWebhook("imagebindings", imagebinding.ListAll, imagebinding.WithContextFactory),
		servicebinding.NewController, NewBindingWebhook("servicebindings", servicebinding.ListAll, servicebinding.WithContextFactory),
	)
}

type WithContextFactory func(ctx context.Context, handler func(name types.NamespacedName)) psbinding.BindableContext
