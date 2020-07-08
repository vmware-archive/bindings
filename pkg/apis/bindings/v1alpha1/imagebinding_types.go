package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/tracker"
	"knative.dev/pkg/webhook/psbinding"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ImageBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageBindingSpec   `json:"spec,omitempty"`
	Status ImageBindingStatus `json:"status,omitempty"`
}

var (
	// Check that ImageBinding can be validated and defaulted.
	_ apis.Validatable   = (*ImageBinding)(nil)
	_ apis.Defaultable   = (*ImageBinding)(nil)
	_ kmeta.OwnerRefable = (*ImageBinding)(nil)
	_ duckv1.KRShaped    = (*ImageBinding)(nil)

	// Check is Bindable
	_ psbinding.Bindable  = (*ImageBinding)(nil)
	_ duck.BindableStatus = (*ImageBindingStatus)(nil)
)

type ImageBindingSpec struct {
	Subject       *tracker.Reference `json:"subject,omitempty"`
	Provider      *tracker.Reference `json:"provider,omitempty"`
	ContainerName string             `json:"containerName,omitempty"`
}

type ImageBindingStatus struct {
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ImageBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageBinding `json:"items"`
}

func (b *ImageBinding) Validate(context.Context) (errs *apis.FieldError) {
	if b.Spec.Subject.Namespace != b.Namespace {
		errs = errs.Also(
			apis.ErrInvalidValue(b.Spec.Subject.Namespace, "spec.subject.namespace"),
		)
	}
	if b.Spec.Provider.Namespace != b.Namespace {
		errs = errs.Also(
			apis.ErrInvalidValue(b.Spec.Provider.Namespace, "spec.provider.namespace"),
		)
	}

	return errs
}

func (b *ImageBinding) SetDefaults(context.Context) {
	if b.Spec.Subject.Namespace == "" {
		// Default the subject's namespace to our namespace.
		b.Spec.Subject.Namespace = b.Namespace
	}
	if b.Spec.Provider.Namespace == "" {
		// Default the provider's namespace to our namespace.
		b.Spec.Provider.Namespace = b.Namespace
	}
}

func (b *ImageBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ImageBinding")
}
