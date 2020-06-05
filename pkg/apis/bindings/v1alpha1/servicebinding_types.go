package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	// Subject resource to bind into
	Subject *tracker.Reference `json:"subject,omitempty"`
	// Provider resoufce container the binding metadata and secret
	Provider *tracker.Reference `json:"provider,omitempty"`
	// ContainerName targets a specific container within the subject to inject
	// into. If not set, all container will be injected
	ContainerName string `json:"containerName,omitempty"`
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
		b.Spec.Subject.Validate(ctx).ViaField("spec.subject"),
	)
	if b.Spec.Subject.Namespace != b.Namespace {
		errs = errs.Also(
			apis.ErrInvalidValue(b.Spec.Subject.Namespace, "spec.subject.namespace"),
		)
	}
	errs = errs.Also(
		b.Spec.Provider.Validate(ctx).ViaField("spec.provider"),
	)
	if b.Spec.Provider.Namespace != b.Namespace {
		errs = errs.Also(
			apis.ErrInvalidValue(b.Spec.Provider.Namespace, "spec.provider.namespace"),
		)
	}
	if b.Spec.Provider.Name == "" {
		errs = errs.Also(
			apis.ErrMissingField("spec.provider.name"),
		)
	}

	return errs
}

func (b *ServiceBinding) SetDefaults(context.Context) {
	if b.Spec.Subject.Namespace == "" {
		// Default the subject's namespace to our namespace.
		b.Spec.Subject.Namespace = b.Namespace
	}
	if b.Spec.Provider.Namespace == "" {
		// Default the provider's namespace to our namespace.
		b.Spec.Provider.Namespace = b.Namespace
	}
}

func (b *ServiceBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ServiceBinding")
}
