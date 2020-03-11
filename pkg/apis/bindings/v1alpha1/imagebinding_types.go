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
	ImageBindingAnnotationKey = GroupName + "/image-binding"
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

	// Check is Bindable
	_ psbinding.Bindable  = (*ImageBinding)(nil)
	_ duck.BindableStatus = (*ImageBindingStatus)(nil)
)

type ImageBindingSpec struct {
	Subject   *tracker.Reference `json:"subject,omitempty"`
	Providers []Provider         `json:"providers,omitempty"`
}

type Provider struct {
	ImageableRef  *tracker.Reference `json:"imageableRef, omitempty"`
	ContainerName string             `json:"containerName,omitempty"`
}

type ImageBindingStatus struct {
	duckv1beta1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ImageBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageBinding `json:"items"`
}

func (b *ImageBinding) Validate(ctx context.Context) (errs *apis.FieldError) {
	if b.Spec.Subject.Namespace != b.Namespace {
		errs = errs.Also(
			apis.ErrInvalidValue(b.Spec.Subject.Namespace, "spec.subject.namespace"),
		)
	}
	for i, p := range b.Spec.Providers {
		if p.ImageableRef.Namespace != b.Namespace {
			errs = errs.Also(
				apis.ErrInvalidValue(p.ImageableRef.Namespace, "imageableRef.namespace").ViaFieldIndex("spec.providers", i),
			)
		}
	}

	return errs
}

func (b *ImageBinding) SetDefaults(context.Context) {
	if b.Spec.Subject.Namespace == "" {
		// Default the subject's namespace to our namespace.
		b.Spec.Subject.Namespace = b.Namespace
	}
	for i, p := range b.Spec.Providers {
		if p.ImageableRef.Namespace == "" {
			b.Spec.Providers[i].ImageableRef.Namespace = b.Namespace
		}
	}
}

func (b *ImageBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ImageBinding")
}
