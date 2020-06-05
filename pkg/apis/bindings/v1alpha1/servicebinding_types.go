package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/tracker"
	"knative.dev/pkg/webhook/psbinding"
)

const (
	ServiceBindingAnnotationKey = GroupName + "/service-binding"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingSpec   `json:"spec,omitempty"`
	Status ServiceBindingStatus `json:"status,omitempty"`
}

var (
	// Check that ServiceBinding can be validated and defaulted.
	_ apis.Validatable   = (*ServiceBinding)(nil)
	_ apis.Defaultable   = (*ServiceBinding)(nil)
	_ kmeta.OwnerRefable = (*ServiceBinding)(nil)

	// Check is Bindable
	_ psbinding.Bindable  = (*ServiceBinding)(nil)
	_ duck.BindableStatus = (*ServiceBindingStatus)(nil)
)

type ServiceBindingSpec struct {
	// Name of the service binding on disk, defaults to this resource's name
	Name string `json:"name,omitempty"`

	// Application resource to inject the binding into
	Application *ApplicationReference `json:"application,omitempty"`
	// Service referencing the binding secret
	Service *tracker.Reference `json:"service,omitempty"`
}

type ApplicationReference struct {
	tracker.Reference

	// Containers to target within the application. If not set, all containers
	// will be injected. Containers may be specified by index or name.
	// InitContainers may only be specified by name.
	Containers []intstr.IntOrString `json:"containers,omitempty"`
}

type ServiceBindingStatus struct {
	duckv1beta1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceBinding `json:"items"`
}

func (b *ServiceBinding) Validate(ctx context.Context) (errs *apis.FieldError) {
	errs = errs.Also(
		b.Spec.Application.Validate(ctx).ViaField("spec.application"),
	)
	if b.Spec.Application.Namespace != b.Namespace {
		errs = errs.Also(
			apis.ErrInvalidValue(b.Spec.Application.Namespace, "spec.application.namespace"),
		)
	}
	errs = errs.Also(
		b.Spec.Service.Validate(ctx).ViaField("spec.service"),
	)
	if b.Spec.Service.Namespace != b.Namespace {
		errs = errs.Also(
			apis.ErrInvalidValue(b.Spec.Service.Namespace, "spec.service.namespace"),
		)
	}
	if b.Spec.Service.Name == "" {
		errs = errs.Also(
			apis.ErrMissingField("spec.service.name"),
		)
	}

	return errs
}

func (b *ServiceBinding) SetDefaults(context.Context) {
	if b.Spec.Name == "" {
		b.Spec.Name = b.Name
	}
	if b.Spec.Application.Namespace == "" {
		// Default the application's namespace to our namespace.
		b.Spec.Application.Namespace = b.Namespace
	}
	if b.Spec.Service.Namespace == "" {
		// Default the service's namespace to our namespace.
		b.Spec.Service.Namespace = b.Namespace
	}
}

func (b *ServiceBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ServiceBinding")
}
