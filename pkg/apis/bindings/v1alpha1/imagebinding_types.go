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
	_ psbinding.Bindable = (*ImageBinding)(nil)
)

type ImageBindingSpec struct {
	Subject   *tracker.Reference `json:"subject,omitempty"`
	Providers []Provider        `json:"providers,omitempty"`
}

type Provider struct {
	ImageableRef  *tracker.Reference `json:"imageableRef, omitempty"`
	ContainerName string            `json:"containerName,omitempty"`
}

var (
	// Check that ImageBinding can be validated and defaulted.
	_ apis.Validatable   = (*ImageBinding)(nil)
	_ apis.Defaultable   = (*ImageBinding)(nil)
	_ kmeta.OwnerRefable = (*ImageBinding)(nil)

	// Check is Bindable
	_ duck.BindableStatus = (*ImageBindingStatus)(nil)
)

type ImageBindingStatus struct {
	duckv1beta1.Status `json:",inline"`

	// ContainerImages contains the current set of images in bound containers on the Source
	ContainerImages []ContainerImage `json:"containerImages,omitempty"`
}

type ContainerImage struct {
	Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ImageBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageBinding `json:"items"`
}

func (ib *ImageBinding) Validate(context.Context) *apis.FieldError {
	return nil
}

func (ib *ImageBinding) SetDefaults(context.Context) {
	if ib.Spec.Subject.Namespace == "" {
		// Default the subject's namespace to our namespace.
		ib.Spec.Subject.Namespace = ib.Namespace
	}
	for i, p := range ib.Spec.Providers {
		if p.ImageableRef.Namespace == "" {
			ib.Spec.Providers[i].ImageableRef.Namespace = ib.Namespace
		}
	}
}

func (ib *ImageBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ImageBinding")
}
